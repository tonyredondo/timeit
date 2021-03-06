package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/montanaflynn/stats"
	"github.com/olekukonko/tablewriter"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type (
	timeout struct {
		MaxDuration      int     `json:"maxDuration"`
		ProcessName      *string `json:"processName"`
		ProcessArguments *string `json:"processArguments"`
	}
	processData struct {
		ProcessName          *string           `json:"processName"`
		ProcessArguments     *string           `json:"processArguments"`
		WorkingDirectory     *string           `json:"workingDirectory"`
		EnvironmentVariables map[string]string `json:"environmentVariables"`
		Timeout              timeout           `json:"timeout"`
	}
	scenario struct {
		processData
		Name string `json:"name"`
	}
	config struct {
		processData
		WarmUpCount   int        `json:"warmUpCount"`
		Count         int        `json:"count"`
		EnableDatadog bool       `json:"enableDatadog"`
		Scenarios     []scenario `json:"scenarios"`
	}
	scenarioResult struct {
		scenario
		scenarioDataPoint
		Data      []scenarioDataPoint
		DataFloat []float64
		Outliers  []float64
		Mean      float64
		Stdev     float64
		P99       float64
		P95       float64
		P90       float64
	}
	scenarioDataPoint struct {
		start    time.Time
		end      time.Time
		duration time.Duration
		error    error
	}
	myLogger struct{}
)

func main() {
	cfg, err := loadConfiguration()
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
		return
	}

	fmt.Printf("Warmup count: %v\n", cfg.WarmUpCount)
	fmt.Printf("Count: %v\n", cfg.Count)
	fmt.Printf("Number of scenarios: %v\n\n", len(cfg.Scenarios))

	// process each scenario
	var resScenario []scenarioResult
	if cfg.Count > 0 && len(cfg.Scenarios) > 0 {
		for _, scenario := range cfg.Scenarios {
			resScenario = append(resScenario, processScenario(&scenario, cfg))
		}
	}

	// print results in a table
	printResultsTable(resScenario, cfg)

	if cfg.EnableDatadog {
		// send traces and span data
		sendTraceData(resScenario, cfg)
	}
}

func sendTraceData(resScenario []scenarioResult, cfg *config) {
	if len(resScenario) > 0 {
		tracer.Start(tracer.WithLogger(myLogger{}), tracer.WithAnalytics(true))
		defer tracer.Stop()

		for _, scenario := range resScenario {
			var pName string
			var pArgs string

			if scenario.ProcessName != nil {
				pName = *scenario.ProcessName
			} else if cfg.ProcessName != nil {
				pName = *cfg.ProcessName
			}

			if scenario.ProcessArguments != nil {
				pArgs = *scenario.ProcessArguments
			} else if cfg.ProcessArguments != nil {
				pArgs = *cfg.ProcessArguments
			}

			span := tracer.StartSpan("time-it", tracer.StartTime(scenario.start))

			for _, datum := range scenario.Data {
				child := tracer.StartSpan("time-it-run", tracer.ChildOf(span.Context()), tracer.StartTime(datum.start))
				child.SetTag(ext.ResourceName, fmt.Sprintf("%v execution of %v %v", scenario.Name, pName, pArgs))
				child.SetTag(ext.ServiceName, pName)
				child.SetTag(ext.SpanType, "benchmark")
				child.SetTag(ext.AnalyticsEvent, true)
				child.SetTag(ext.ManualKeep, true)
				child.SetTag("process.name", pName)
				child.SetTag("process.arguments", pArgs)
				child.Finish(tracer.FinishTime(datum.end), tracer.WithError(datum.error))
			}

			span.SetTag(ext.ResourceName, fmt.Sprintf("%v (%v %v)", scenario.Name, pName, pArgs))
			span.SetTag(ext.ServiceName, pName)
			span.SetTag(ext.SpanType, "benchmark")
			span.SetTag(ext.AnalyticsEvent, true)
			span.SetTag(ext.ManualKeep, true)
			span.SetTag("scenario.name", scenario.Name)
			span.SetTag("scenario.mean", scenario.Mean)
			span.SetTag("scenario.stdev", scenario.Stdev)
			span.SetTag("scenario.p90", scenario.P90)
			span.SetTag("scenario.p95", scenario.P95)
			span.SetTag("scenario.p99", scenario.P99)
			span.SetTag("scenario.outliers", len(scenario.Outliers))
			span.SetTag("configuration.count", cfg.Count)
			span.SetTag("configuration.warmup_count", cfg.WarmUpCount)
			span.SetTag("process.name", pName)
			span.SetTag("process.arguments", pArgs)
			span.Finish(tracer.FinishTime(scenario.end), tracer.WithError(scenario.error))
		}
	}
}

func printResultsTable(resScenario []scenarioResult, cfg *config) {
	fmt.Println("\n### Results\n")
	resultTable := tablewriter.NewWriter(os.Stdout)
	resultTable.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	resultTable.SetCenterSeparator("|")
	resultTable.SetHeaderAlignment(tablewriter.ALIGN_CENTER)
	resultTable.SetAlignment(tablewriter.ALIGN_CENTER)
	var resultHeader []string
	for scidx := 0; scidx < len(resScenario); scidx++ {
		resultHeader = append(resultHeader, resScenario[scidx].Name)
	}
	resultTable.SetHeader(resultHeader)
	for idx := 0; idx < cfg.Count; idx++ {
		var resultRow []string
		for scidx := 0; scidx < len(resScenario); scidx++ {
			resultRow = append(resultRow, fmt.Sprint(time.Duration(resScenario[scidx].DataFloat[idx])))
		}
		resultTable.Append(resultRow)
	}
	resultTable.Render()

	fmt.Println("\n### Outliers\n")
	outliersTable := tablewriter.NewWriter(os.Stdout)
	outliersTable.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	outliersTable.SetCenterSeparator("|")
	outliersTable.SetHeaderAlignment(tablewriter.ALIGN_CENTER)
	outliersTable.SetAlignment(tablewriter.ALIGN_CENTER)
	outliersTable.SetHeader(resultHeader)
	maxOutliersLength := 0
	for scidx := 0; scidx < len(resScenario); scidx++ {
		outLength := len(resScenario[scidx].Outliers)
		if maxOutliersLength < outLength {
			maxOutliersLength = outLength
		}
	}
	for idx := 0; idx < maxOutliersLength; idx++ {
		var resultRow []string
		for scidx := 0; scidx < len(resScenario); scidx++ {
			outliersArray := resScenario[scidx].Outliers
			if idx < len(outliersArray) {
				resultRow = append(resultRow, fmt.Sprint(time.Duration(outliersArray[idx])))
			} else {
				resultRow = append(resultRow, " - ")
			}
		}
		outliersTable.Append(resultRow)
	}
	outliersTable.Render()

	fmt.Println("\n### Summary\n")
	summaryTable := tablewriter.NewWriter(os.Stdout)
	summaryTable.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	summaryTable.SetCenterSeparator("|")
	summaryTable.SetHeader([]string{"Name", "Mean", "Stdev", "P99", "P95", "P90", "Outliers"})
	for scidx := 0; scidx < len(resScenario); scidx++ {
		summaryTable.Append([]string{
			resScenario[scidx].Name,
			fmt.Sprint(time.Duration(resScenario[scidx].Mean)),
			fmt.Sprint(time.Duration(resScenario[scidx].Stdev)),
			fmt.Sprint(time.Duration(resScenario[scidx].P99)),
			fmt.Sprint(time.Duration(resScenario[scidx].P95)),
			fmt.Sprint(time.Duration(resScenario[scidx].P90)),
			fmt.Sprint(len(resScenario[scidx].Outliers)),
		})
	}
	summaryTable.Render()
	fmt.Println()

	for scidx := 0; scidx < len(resScenario); scidx++ {
		if resScenario[scidx].error != nil {
			fmt.Printf("Error in Scenario: %v\n", scidx)
			fmt.Println(resScenario[scidx].error.Error())
		}
	}
}

func loadConfiguration() (*config, error) {
	if len(os.Args) < 2 {
		return nil, errors.New("missing argument with the configuration file")
	}

	jsonFile, err := os.Open(os.Args[1])
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	jsonBytes, err2 := ioutil.ReadAll(jsonFile)
	if err2 != nil {
		return nil, err2
	}

	var cfg config
	err = json.Unmarshal(jsonBytes, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func processScenario(scenario *scenario, cfg *config) scenarioResult {
	fmt.Printf("Scenario: %v\n", scenario.Name)
	fmt.Print("  Warming up")
	start := time.Now()
	_ = runScenario(cfg.WarmUpCount, scenario, cfg)
	end := time.Now()
	fmt.Printf("    Duration: %v\n", end.Sub(start))
	fmt.Print("  Run")
	start = time.Now()
	res := runScenario(cfg.Count, scenario, cfg)
	end = time.Now()
	fmt.Printf("    Duration: %v\n", end.Sub(start))
	fmt.Println()

	var durations []float64
	mapErrors := make(map[string]bool)
	for _, item := range res {
		durations = append(durations, float64(item.duration))
		if item.error != nil && item.error != context.DeadlineExceeded {
			mapErrors[item.error.Error()] = true
		}
	}
	var errorString string
	for k, _ := range mapErrors {
		errorString += fmt.Sprintln(k)
	}
	var error error
	if errorString != "" {
		error = errors.New(errorString)
	}

	// Get outliers
	outliers, _ := stats.QuartileOutliers(durations)
	extremeOutliers := outliers.Extreme

	durationsCount := len(durations)
	outliersCount := len(extremeOutliers)

	// Remove outliers
	var newDurations []float64
	for x := 0; x < durationsCount; x++ {
		add := true
		for j := 0; j < outliersCount; j++ {
			if durations[x] == extremeOutliers[j] {
				add = false
				break
			}
		}
		if add {
			newDurations = append(newDurations, durations[x])
		}
	}

	durations = newDurations
	mean, _ := stats.Mean(durations)

	// Add the missing datapoints removed from outliers
	missingDurations := durationsCount - len(durations)
	for i := 0; i < missingDurations; i++ {
		durations = append(durations, mean)
	}

	stdev, _ := stats.StandardDeviation(durations)
	p99, _ := stats.Percentile(durations, 99)
	p95, _ := stats.Percentile(durations, 95)
	p90, _ := stats.Percentile(durations, 90)

	return scenarioResult{
		scenario: *scenario,
		scenarioDataPoint: scenarioDataPoint{
			start:    start,
			end:      end,
			duration: end.Sub(start),
			error:    error,
		},
		Data:      res,
		DataFloat: durations,
		Outliers:  extremeOutliers,
		Mean:      mean,
		Stdev:     stdev,
		P99:       p99,
		P95:       p95,
		P90:       p90,
	}
}

func runScenario(count int, scenario *scenario, cfg *config) []scenarioDataPoint {
	var res []scenarioDataPoint
	fmt.Print(" ")
	for i := 0; i < count; i++ {
		currentRun := runProcessCmd(scenario, cfg)
		res = append(res, currentRun)
		if currentRun.error != nil {
			fmt.Print("x")
		} else {
			fmt.Print(".")
		}
	}
	fmt.Println()
	return res
}

func runProcessCmd(scenario *scenario, cfg *config) scenarioDataPoint {
	var cmdString string
	if scenario.ProcessName != nil {
		cmdString = *scenario.ProcessName
	} else if cfg.ProcessName != nil {
		cmdString = *cfg.ProcessName
	}

	var cmdArguments string
	if scenario.ProcessArguments != nil {
		cmdArguments = *scenario.ProcessArguments
	} else if cfg.ProcessArguments != nil {
		cmdArguments = *cfg.ProcessArguments
	}

	var workingDirectory string
	if scenario.WorkingDirectory != nil {
		workingDirectory = *scenario.WorkingDirectory
	} else if cfg.WorkingDirectory != nil {
		workingDirectory = *cfg.WorkingDirectory
	}

	cmdEnv := os.Environ()
	for k, v := range cfg.EnvironmentVariables {
		cmdEnv = append(cmdEnv, fmt.Sprintf("%s=%s", k, v))
	}
	for k, v := range scenario.EnvironmentVariables {
		cmdEnv = append(cmdEnv, fmt.Sprintf("%s=%s", k, v))
	}

	cmdTimeout := 0
	if scenario.Timeout.MaxDuration > 0 {
		cmdTimeout = scenario.Timeout.MaxDuration
	} else if cfg.Timeout.MaxDuration > 0 {
		cmdTimeout = cfg.Timeout.MaxDuration
	}

	var timeoutCmdString string
	if scenario.Timeout.ProcessName != nil {
		timeoutCmdString = *scenario.Timeout.ProcessName
	} else if cfg.Timeout.ProcessName != nil {
		timeoutCmdString = *cfg.Timeout.ProcessName
	}

	var timeoutCmdArguments string
	if scenario.Timeout.ProcessArguments != nil {
		timeoutCmdArguments = *scenario.Timeout.ProcessArguments
	} else if cfg.Timeout.ProcessArguments != nil {
		timeoutCmdArguments = *cfg.Timeout.ProcessArguments
	}

	defer runtime.GC()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var cmd *exec.Cmd
	if len(cmdArguments) > 0 {
		cmd = exec.CommandContext(ctx, cmdString, strings.Split(cmdArguments, " ")...)
	} else {
		cmd = exec.CommandContext(ctx, cmdString)
	}
	cmd.Dir = workingDirectory
	cmd.Env = cmdEnv

	if cmdTimeout > 0 {
		go func() {
			select {
			case <-time.After(time.Duration(cmdTimeout) * time.Second):
				if timeoutCmdString != "" {
					fmt.Printf("[%v]", cmd.Process.Pid)
					timeoutCmdString = strings.ReplaceAll(timeoutCmdString, "%pid%", fmt.Sprint(cmd.Process.Pid))
					timeoutCmdArguments = strings.ReplaceAll(timeoutCmdArguments, "%pid%", fmt.Sprint(cmd.Process.Pid))
					var timeoutCmd *exec.Cmd
					if len(timeoutCmdArguments) > 0 {
						timeoutCmd = exec.CommandContext(ctx, timeoutCmdString, strings.Split(timeoutCmdArguments, " ")...)
					} else {
						timeoutCmd = exec.CommandContext(ctx, timeoutCmdString)
					}
					err := timeoutCmd.Run()
					if err != nil {
						fmt.Println(err)
					}
				}
				cancel()
			case <-ctx.Done():
				cancel()
			}
		}()
	}

	start := time.Now()
	err := cmd.Run()
	end := time.Now()

	if ctx.Err() == context.DeadlineExceeded {
		err = ctx.Err()
	}
	return scenarioDataPoint{
		start:    start,
		end:      end,
		duration: end.Sub(start),
		error:    err,
	}
}

func (l myLogger) Log(msg string) {
	if strings.Index(msg, "INFO:") == -1 {
		fmt.Printf("%v\n", msg)
	}
}
