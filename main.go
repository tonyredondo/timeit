package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	ddtesting "github.com/DataDog/dd-sdk-go-testing"
	"github.com/loov/hrtime"
	"github.com/montanaflynn/stats"
	"github.com/olekukonko/tablewriter"
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
		Tags                 map[string]string `json:"tags"`
		MetricsFilePath      *string           `json:"metricsFilePath"`
	}
	scenario struct {
		processData
		Name string `json:"name"`
	}
	config struct {
		processData
		FilePath      string
		Path          string
		FileName      string
		WarmUpCount   int        `json:"warmUpCount"`
		Count         int        `json:"count"`
		EnableDatadog bool       `json:"enableDatadog"`
		Scenarios     []scenario `json:"scenarios"`
	}
	scenarioResult struct {
		scenario
		scenarioDataPoint
		Data        []scenarioDataPoint
		DataFloat   []float64
		Outliers    []float64
		Mean        float64
		Max         float64
		Min         float64
		Stdev       float64
		StdErr      float64
		P99         float64
		P95         float64
		P90         float64
		Metrics     map[string]float64
		MetricsData map[string][]float64
	}
	scenarioDataPoint struct {
		start    time.Time
		end      time.Time
		duration time.Duration
		metrics  map[string]float64
		error    error
	}
	metricsItem struct {
		key   string
		value []float64
	}
	myLogger struct{}
)

var currentWorkingDirectory *string

func main() {
	fmt.Println("TimeIt by Tony Redondo\n")
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
	scenarioWithErrors := 0
	if cfg.Count > 0 && len(cfg.Scenarios) > 0 {
		for _, scenario := range cfg.Scenarios {
			res := processScenario(&scenario, cfg)
			if res.error != nil {
				scenarioWithErrors++
			}
			resScenario = append(resScenario, res)
		}
	}

	if scenarioWithErrors < len(cfg.Scenarios) {
		// print results in a table
		printResultsTable(resScenario, cfg)

		if cfg.EnableDatadog {
			// send traces and span data
			sendTraceData(resScenario, cfg)
		}
	} else {
		for scidx := 0; scidx < len(resScenario); scidx++ {
			if resScenario[scidx].error != nil {
				fmt.Printf("Error in Scenario: %v\n", scidx)
				fmt.Println(resScenario[scidx].error.Error())
			}
		}

		os.Exit(1)
	}
}

func sendTraceData(resScenario []scenarioResult, cfg *config) {
	if len(resScenario) > 0 {
		finalizer := ddtesting.Initialize(tracer.WithLogger(myLogger{}), tracer.WithAnalytics(true))
		defer finalizer()

		for _, scenario := range resScenario {
			var pName string
			var pArgs string
			var tags map[string]string

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

			tags = cfg.Tags
			for k, v := range scenario.Tags {
				tags[k] = v
			}

			var startSpanOptions []tracer.StartSpanOption
			startSpanOptions = append(startSpanOptions, tracer.StartTime(scenario.start))
			startSpanOptions = append(startSpanOptions, tracer.Tag("benchmark.job.description", scenario.Name))
			startSpanOptions = append(startSpanOptions, tracer.Tag("benchmark.runs", cfg.Count))
			startSpanOptions = append(startSpanOptions, tracer.Tag("benchmark.warmup_count", cfg.WarmUpCount))
			startSpanOptions = append(startSpanOptions, tracer.Tag("benchmark.duration.mean", scenario.Mean))
			startSpanOptions = append(startSpanOptions, tracer.Tag("benchmark.statistics.n", cfg.Count))
			startSpanOptions = append(startSpanOptions, tracer.Tag("benchmark.statistics.mean", scenario.Mean))
			startSpanOptions = append(startSpanOptions, tracer.Tag("benchmark.statistics.max", scenario.Max))
			startSpanOptions = append(startSpanOptions, tracer.Tag("benchmark.statistics.min", scenario.Min))
			startSpanOptions = append(startSpanOptions, tracer.Tag("benchmark.statistics.std_dev", scenario.Stdev))
			startSpanOptions = append(startSpanOptions, tracer.Tag("benchmark.statistics.std_err", scenario.StdErr))
			startSpanOptions = append(startSpanOptions, tracer.Tag("benchmark.statistics.p90", scenario.P90))
			startSpanOptions = append(startSpanOptions, tracer.Tag("benchmark.statistics.p95", scenario.P95))
			startSpanOptions = append(startSpanOptions, tracer.Tag("benchmark.statistics.p99", scenario.P99))
			startSpanOptions = append(startSpanOptions, tracer.Tag("benchmark.statistics.outliers", len(scenario.Outliers)))
			startSpanOptions = append(startSpanOptions, tracer.Tag("process.name", pName))
			startSpanOptions = append(startSpanOptions, tracer.Tag("process.arguments", pArgs))
			startSpanOptions = append(startSpanOptions, tracer.Tag("test.file.path", cfg.Path))
			startSpanOptions = append(startSpanOptions, tracer.Tag("test.file.name", cfg.FileName))
			startSpanOptions = append(startSpanOptions, tracer.Tag("test.file.file_path", cfg.FilePath))
			startSpanOptions = append(startSpanOptions, tracer.Tag("test.scenario", scenario.Name))
			startSpanOptions = append(startSpanOptions, tracer.Tag("test.framework", "time-it"))
			for k, v := range tags {
				startSpanOptions = append(startSpanOptions, tracer.Tag(k, v))
			}
			for k, v := range scenario.Metrics {
				startSpanOptions = append(startSpanOptions, tracer.Tag(fmt.Sprintf("benchmark.%v", k), v))
			}

			_, testFinish := ddtesting.StartCustomTestOrBenchmark(context.Background(), ddtesting.TestData{
				Type:  ddtesting.TypeBenchmark,
				Suite: fmt.Sprintf("time-it.%v", scenario.Name),
				Name:  cfg.FilePath,
				Options: []ddtesting.Option{
					ddtesting.WithSpanOptions(startSpanOptions...),
					ddtesting.WithFinishOptions(tracer.FinishTime(scenario.end), tracer.WithError(scenario.error)),
				},
			})

			testFinish(ddtesting.StatusPass, nil)
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
	summaryTable.SetHeader([]string{"Name", "Mean", "StdDev", "StdErr", "P99", "P95", "P90", "Outliers"})
	for scidx := 0; scidx < len(resScenario); scidx++ {
		summaryTable.Append([]string{
			resScenario[scidx].Name,
			fmt.Sprint(time.Duration(resScenario[scidx].Mean)),
			fmt.Sprint(time.Duration(resScenario[scidx].Stdev)),
			fmt.Sprint(time.Duration(resScenario[scidx].StdErr)),
			fmt.Sprint(time.Duration(resScenario[scidx].P99)),
			fmt.Sprint(time.Duration(resScenario[scidx].P95)),
			fmt.Sprint(time.Duration(resScenario[scidx].P90)),
			fmt.Sprint(len(resScenario[scidx].Outliers)),
		})

		if len(resScenario[scidx].MetricsData) > 0 {
			for _, item := range orderByKey(resScenario[scidx].MetricsData) {
				mMean, _ := stats.Mean(item.value)
				mStdDev, _ := stats.StandardDeviation(item.value)
				mStdErr := mStdDev / math.Sqrt(float64(len(resScenario[scidx].DataFloat)))
				mP99, _ := stats.Percentile(item.value, 99)
				mP95, _ := stats.Percentile(item.value, 95)
				mP90, _ := stats.Percentile(item.value, 90)

				summaryTable.Append([]string{
					fmt.Sprintf("└─>%v", item.key),
					fmt.Sprint(toFixed(mMean, 6)),
					fmt.Sprint(toFixed(mStdDev, 6)),
					fmt.Sprint(toFixed(mStdErr, 6)),
					fmt.Sprint(toFixed(mP99, 6)),
					fmt.Sprint(toFixed(mP95, 6)),
					fmt.Sprint(toFixed(mP90, 6)),
					"",
				})
			}

			summaryTable.Append([]string{"", "", "", "", "", "", "", ""})
		}
	}
	summaryTable.Render()
	fmt.Println()
}

func loadConfiguration() (*config, error) {
	if len(os.Args) < 2 {
		return nil, errors.New("missing argument with the configuration file")
	}

	configurationFilePath := os.Args[1]

	jsonFile, err := os.Open(configurationFilePath)
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

	cfg.FilePath = configurationFilePath
	cfg.Path = filepath.Dir(configurationFilePath)
	cfg.FileName = filepath.Base(configurationFilePath)

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
	metricsData := map[string][]float64{}
	mapErrors := make(map[string]bool)
	for _, item := range res {
		durations = append(durations, float64(item.duration))
		if item.error != nil && item.error != context.DeadlineExceeded {
			mapErrors[item.error.Error()] = true
		}
		for k, v := range item.metrics {
			metricsData[k] = append(metricsData[k], v)
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
	max, _ := stats.Max(durations)
	min, _ := stats.Min(durations)

	// Add the missing datapoints removed from outliers
	missingDurations := durationsCount - len(durations)
	for i := 0; i < missingDurations; i++ {
		durations = append(durations, mean)
	}

	stdev, _ := stats.StandardDeviation(durations)
	p99, _ := stats.Percentile(durations, 99)
	p95, _ := stats.Percentile(durations, 95)
	p90, _ := stats.Percentile(durations, 90)
	stderr := stdev / math.Sqrt(float64(durationsCount))

	// Calculate metrics stats
	metricsStats := map[string]float64{}
	for k, v := range metricsData {
		mMean, _ := stats.Mean(v)
		mMax, _ := stats.Max(v)
		mMin, _ := stats.Min(v)
		mStdDev, _ := stats.StandardDeviation(v)
		mStdErr := mStdDev / math.Sqrt(float64(durationsCount))
		mP99, _ := stats.Percentile(v, 99)
		mP95, _ := stats.Percentile(v, 95)
		mP90, _ := stats.Percentile(v, 90)

		metricsStats[fmt.Sprintf("%v.mean", k)] = mMean
		metricsStats[fmt.Sprintf("%v.max", k)] = mMax
		metricsStats[fmt.Sprintf("%v.min", k)] = mMin
		metricsStats[fmt.Sprintf("%v.std_dev", k)] = mStdDev
		metricsStats[fmt.Sprintf("%v.std_err", k)] = mStdErr
		metricsStats[fmt.Sprintf("%v.p99", k)] = mP99
		metricsStats[fmt.Sprintf("%v.p95", k)] = mP95
		metricsStats[fmt.Sprintf("%v.p90", k)] = mP90
	}

	return scenarioResult{
		scenario: *scenario,
		scenarioDataPoint: scenarioDataPoint{
			start:    start,
			end:      end,
			duration: end.Sub(start),
			error:    error,
		},
		Data:        res,
		DataFloat:   durations,
		Outliers:    extremeOutliers,
		Mean:        mean,
		Max:         max,
		Min:         min,
		Stdev:       stdev,
		StdErr:      stderr,
		P99:         p99,
		P95:         p95,
		P90:         p90,
		Metrics:     metricsStats,
		MetricsData: metricsData,
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
	cmdString = replaceCustomVars(cmdString)

	var cmdArguments string
	if scenario.ProcessArguments != nil {
		cmdArguments = *scenario.ProcessArguments
	} else if cfg.ProcessArguments != nil {
		cmdArguments = *cfg.ProcessArguments
	}
	cmdArguments = replaceCustomVars(cmdArguments)

	var workingDirectory string
	if scenario.WorkingDirectory != nil {
		workingDirectory = *scenario.WorkingDirectory
	} else if cfg.WorkingDirectory != nil {
		workingDirectory = *cfg.WorkingDirectory
	}
	workingDirectory = replaceCustomVars(workingDirectory)

	cmdEnv := os.Environ()
	for k, v := range cfg.EnvironmentVariables {
		v = replaceCustomVars(v)
		cmdEnv = append(cmdEnv, fmt.Sprintf("%s=%s", k, v))
	}
	for k, v := range scenario.EnvironmentVariables {
		v = replaceCustomVars(v)
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
	timeoutCmdString = replaceCustomVars(timeoutCmdString)

	var timeoutCmdArguments string
	if scenario.Timeout.ProcessArguments != nil {
		timeoutCmdArguments = *scenario.Timeout.ProcessArguments
	} else if cfg.Timeout.ProcessArguments != nil {
		timeoutCmdArguments = *cfg.Timeout.ProcessArguments
	}
	timeoutCmdArguments = replaceCustomVars(timeoutCmdArguments)

	var metricsFilePath string
	if scenario.MetricsFilePath != nil {
		metricsFilePath = *scenario.MetricsFilePath
	} else if cfg.MetricsFilePath != nil {
		metricsFilePath = *cfg.MetricsFilePath
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
	startDur := hrtime.Now()
	err := cmd.Run()
	endDur := hrtime.Now()
	end := time.Now()

	if ctx.Err() == context.DeadlineExceeded {
		err = ctx.Err()
	}

	metricsData := map[string]float64{}
	if metricsFilePath != "" {
		if _, err := os.Stat(metricsFilePath); err == nil {
			data, err := os.ReadFile(metricsFilePath)
			if err == nil {
				var metricsJson []map[string]string
				_ = json.Unmarshal(data, &metricsJson)

				for _, item := range metricsJson {
					for k, v := range item {
						if s, err := strconv.ParseFloat(v, 64); err == nil {
							metricsData[k] = s
						}
					}
				}
			}
		}
	}

	return scenarioDataPoint{
		start:    start,
		end:      end,
		duration: endDur - startDur,
		metrics:  metricsData,
		error:    err,
	}
}

func (l myLogger) Log(msg string) {
	if strings.Index(msg, "INFO:") == -1 {
		fmt.Printf("%v\n", msg)
	}
}

func replaceCustomVars(value string) string {
	if currentWorkingDirectory == nil {
		wd, _ := os.Getwd()
		currentWorkingDirectory = &wd
	}
	value = strings.ReplaceAll(value, "$(CWD)", *currentWorkingDirectory)

	return value
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

func orderByKey(data map[string][]float64) []metricsItem {
	var value []metricsItem
	for k, v := range data {
		value = append(value, metricsItem{
			key:   k,
			value: v,
		})
	}
	sort.Slice(value, func(i, j int) bool {
		return value[i].key < value[j].key
	})
	return value
}
