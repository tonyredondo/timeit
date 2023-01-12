package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/loov/hrtime"
	"github.com/montanaflynn/stats"
	"github.com/olekukonko/tablewriter"
)

type (
	scenarioResult struct {
		scenario
		scenarioDataPoint
		WarmUpCount int                  `json:"warmUpCount"`
		Count       int                  `json:"count"`
		Data        []scenarioDataPoint  `json:"data"`
		DataFloat   []float64            `json:"durations"`
		Outliers    []float64            `json:"outliers"`
		Mean        float64              `json:"mean"`
		Max         float64              `json:"max"`
		Min         float64              `json:"min"`
		Stdev       float64              `json:"stdev"`
		StdErr      float64              `json:"stderr"`
		P99         float64              `json:"p99"`
		P95         float64              `json:"p95"`
		P90         float64              `json:"p90"`
		Metrics     map[string]float64   `json:"metrics"`
		MetricsData map[string][]float64 `json:"metricsData"`
	}
	scenarioDataPoint struct {
		Start          time.Time     `json:"start"`
		End            time.Time     `json:"end"`
		Duration       time.Duration `json:"duration"`
		Error          error         `json:"error"`
		metrics        map[string]float64
		shouldContinue bool
	}
	metricsItem struct {
		key   string
		value []float64
	}
)

var currentWorkingDirectory *string
var exporters []Exporter

func main() {
	fmt.Print("TimeIt by Tony Redondo\n\n")
	cfg, err := loadConfiguration()
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
		return
	}

	cfg.JsonExporterFilePath = replaceCustomVars(cfg.JsonExporterFilePath)
	exporters = append(exporters, newDatadogExporter(), newJsonExporter())

	fmt.Printf("Warmup count: %v\n", cfg.WarmUpCount)
	fmt.Printf("Count: %v\n", cfg.Count)
	fmt.Printf("Number of scenarios: %v\n\n", len(cfg.Scenarios))

	// process each scenario
	var resScenario []scenarioResult
	scenarioWithErrors := 0
	if cfg.Count > 0 && len(cfg.Scenarios) > 0 {
		for _, sce := range cfg.Scenarios {
			// Prepare scenario
			prepareScenario(&sce, cfg)

			// Process scenario
			res := processScenario(&sce, cfg)
			if res.Error != nil {
				scenarioWithErrors++
			}
			resScenario = append(resScenario, res)
		}
	}

	if scenarioWithErrors < len(cfg.Scenarios) {
		// print results in a table
		printResultsTable(resScenario, cfg)

		// Export data
		for _, ex := range exporters {
			ex.SetConfiguration(cfg)
			if ex.IsEnabled() {
				ex.Export(resScenario)
			}
		}

	} else {
		for scidx := 0; scidx < len(resScenario); scidx++ {
			if resScenario[scidx].Error != nil {
				fmt.Printf("Error in Scenario: %v\n", scidx)
				fmt.Println(resScenario[scidx].Error.Error())
			}
		}

		os.Exit(1)
	}
}

func prepareScenario(sce *scenario, cfg *config) {
	// let's fill the scenario with the required data
	if (sce.ProcessName == nil || *sce.ProcessName == "") && cfg.ProcessName != nil {
		sce.ProcessName = cfg.ProcessName
	}
	if sce.ProcessName != nil {
		*sce.ProcessName = replaceCustomVars(*sce.ProcessName)
	}

	if (sce.ProcessArguments == nil || *sce.ProcessArguments == "") && cfg.ProcessArguments != nil {
		sce.ProcessArguments = cfg.ProcessArguments
	}
	if sce.ProcessArguments != nil {
		*sce.ProcessArguments = replaceCustomVars(*sce.ProcessArguments)
	}

	if (sce.WorkingDirectory == nil || *sce.WorkingDirectory == "") && cfg.WorkingDirectory != nil {
		sce.WorkingDirectory = cfg.WorkingDirectory
	}
	if sce.WorkingDirectory != nil {
		*sce.WorkingDirectory = replaceCustomVars(*sce.WorkingDirectory)
	}

	for k, v := range sce.EnvironmentVariables {
		sce.EnvironmentVariables[k] = replaceCustomVars(v)
	}
	for k, v := range cfg.EnvironmentVariables {
		v = replaceCustomVars(v)
		_, hasKey := sce.EnvironmentVariables[k]
		if !hasKey {
			sce.EnvironmentVariables[k] = v
		}
	}

	if sce.Timeout.MaxDuration <= 0 && cfg.Timeout.MaxDuration > 0 {
		sce.Timeout.MaxDuration = cfg.Timeout.MaxDuration
	}

	if (sce.Timeout.ProcessName == nil || *sce.Timeout.ProcessName == "") && cfg.Timeout.ProcessName != nil {
		sce.Timeout.ProcessName = cfg.Timeout.ProcessName
	}
	if sce.Timeout.ProcessName != nil {
		*sce.Timeout.ProcessName = replaceCustomVars(*sce.Timeout.ProcessName)
	}

	if (sce.Timeout.ProcessArguments == nil || *sce.Timeout.ProcessArguments == "") && cfg.Timeout.ProcessArguments != nil {
		sce.Timeout.ProcessArguments = cfg.Timeout.ProcessArguments
	}
	if sce.Timeout.ProcessArguments != nil {
		*sce.Timeout.ProcessArguments = replaceCustomVars(*sce.Timeout.ProcessArguments)
	}

	if (sce.MetricsFilePath == nil || *sce.MetricsFilePath == "") && cfg.MetricsFilePath != nil {
		sce.MetricsFilePath = cfg.MetricsFilePath
	}
	if sce.MetricsFilePath != nil {
		*sce.MetricsFilePath = replaceCustomVars(*sce.MetricsFilePath)
	}
}

func processScenario(scenario *scenario, cfg *config) scenarioResult {
	fmt.Printf("Scenario: %v\n", scenario.Name)
	fmt.Print("  Warming up")
	start := time.Now()
	_ = runScenario(cfg.WarmUpCount, scenario)
	end := time.Now()
	fmt.Printf("    Duration: %v\n", end.Sub(start))
	fmt.Print("  Run")
	start = time.Now()
	res := runScenario(cfg.Count, scenario)
	end = time.Now()
	fmt.Printf("    Duration: %v\n", end.Sub(start))
	fmt.Println()

	var durations []float64
	metricsData := map[string][]float64{}
	mapErrors := make(map[string]bool)
	for _, item := range res {
		durations = append(durations, float64(item.Duration))
		if item.Error != nil && item.Error != context.DeadlineExceeded {
			mapErrors[item.Error.Error()] = true
		}
		for k, v := range item.metrics {
			metricsData[k] = append(metricsData[k], v)
		}
	}
	var errorString string
	for k := range mapErrors {
		errorString += fmt.Sprintln(k)
	}
	var sceError error
	if errorString != "" {
		sceError = errors.New(errorString)
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
			Start:    start,
			End:      end,
			Duration: end.Sub(start),
			Error:    sceError,
		},
		WarmUpCount: cfg.WarmUpCount,
		Count:       cfg.Count,
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

func runScenario(count int, scenario *scenario) []scenarioDataPoint {
	var res []scenarioDataPoint
	fmt.Print(" ")
	for i := 0; i < count; i++ {
		currentRun := runProcessCmd(scenario)
		res = append(res, currentRun)
		if !currentRun.shouldContinue {
			break
		}
		if currentRun.Error != nil {
			fmt.Print("x")
		} else {
			fmt.Print(".")
		}
	}
	fmt.Println()
	return res
}

func runProcessCmd(sce *scenario) scenarioDataPoint {
	var cmdString string
	var cmdArguments string
	var workingDirectory string
	var timeoutCmdString string
	var timeoutCmdArguments string

	if sce.ProcessName != nil {
		cmdString = *sce.ProcessName
	}

	if sce.ProcessArguments != nil {
		cmdArguments = *sce.ProcessArguments
	}

	if sce.WorkingDirectory != nil {
		workingDirectory = *sce.WorkingDirectory
	}

	cmdTimeout := sce.Timeout.MaxDuration

	if sce.Timeout.ProcessName != nil {
		timeoutCmdString = *sce.Timeout.ProcessName
	}

	if sce.Timeout.ProcessArguments != nil {
		timeoutCmdArguments = *sce.Timeout.ProcessArguments
	}

	cmdEnv := os.Environ()
	for k, v := range sce.EnvironmentVariables {
		cmdEnv = append(cmdEnv, fmt.Sprintf("%s=%s", k, v))
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

	var b bytes.Buffer
	cmd.Stdout = &b
	cmd.Stderr = &b

	shouldContinue := true
	start := time.Now()
	startDur := hrtime.Now()
	err := cmd.Run()
	endDur := hrtime.Now()
	end := time.Now()

	if ctx.Err() == context.DeadlineExceeded {
		err = ctx.Err()
	} else if err != nil {
		err = errors.New(fmt.Sprintf("\n%s%s", b.String(), err.Error()))
	}

	// Since metrics file(s) are created during or at the end of the process
	// we have to look for them once the process is finished
	var metricsFilesPath []string
	if sce.MetricsFilePath != nil {
		metricsFilesPath = resolveWildcard(*sce.MetricsFilePath, workingDirectory)
	}

	metricsData := map[string]float64{}
	if len(metricsFilesPath) > 0 {
		for _, metricsFilePath := range metricsFilesPath {
			if _, lerr := os.Stat(metricsFilePath); lerr == nil {
				data, lerr := os.ReadFile(metricsFilePath)
				if lerr == nil {
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
			} else if os.IsNotExist(lerr) {
				err = errors.New(fmt.Sprintf("MetricsFilePath '%v' not found.", metricsFilePath))
				shouldContinue = false
			}
		}
	}

	return scenarioDataPoint{
		Start:          start,
		End:            end,
		Duration:       endDur - startDur,
		Error:          err,
		metrics:        metricsData,
		shouldContinue: shouldContinue,
	}
}

func printResultsTable(resScenario []scenarioResult, cfg *config) {
	fmt.Print("\n### Results\n\n")
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

	fmt.Print("\n### Outliers\n\n")
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

	fmt.Print("\n### Summary\n\n")
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

		totalNum := len(resScenario[scidx].MetricsData)
		if totalNum > 0 {
			for idx, item := range orderByKey(resScenario[scidx].MetricsData) {
				mMean, _ := stats.Mean(item.value)
				mStdDev, _ := stats.StandardDeviation(item.value)
				mStdErr := mStdDev / math.Sqrt(float64(len(resScenario[scidx].DataFloat)))
				mP99, _ := stats.Percentile(item.value, 99)
				mP95, _ := stats.Percentile(item.value, 95)
				mP90, _ := stats.Percentile(item.value, 90)

				var name string
				if idx < totalNum-1 {
					name = fmt.Sprintf("├>%v", item.key)
				} else {
					name = fmt.Sprintf("└>%v", item.key)
				}

				summaryTable.Append([]string{
					name,
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
