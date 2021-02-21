package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/montanaflynn/stats"
	"github.com/olekukonko/tablewriter"
)

type (
	processData struct {
		ProcessName          *string           `json:"processName"`
		ProcessArguments     *string           `json:"processArguments"`
		WorkingDirectory     *string           `json:"workingDirectory"`
		EnvironmentVariables map[string]string `json:"environmentVariables"`
	}
	scenario struct {
		processData
		Name string `json:"name"`
	}
	config struct {
		processData
		WarmUpCount int        `json:"warmUpCount"`
		Count       int        `json:"count"`
		Scenarios   []scenario `json:"scenarios"`
	}
	scenarioResult struct {
		scenario
		Data  []float64
		Mean  float64
		Stdev float64
		P99   float64
		P95   float64
		P90   float64
	}
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
			resultRow = append(resultRow, fmt.Sprint(time.Duration(resScenario[scidx].Data[idx])))
		}
		resultTable.Append(resultRow)
	}
	resultTable.Render()

	fmt.Println("\n### Summary\n")
	summaryTable := tablewriter.NewWriter(os.Stdout)
	summaryTable.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	summaryTable.SetCenterSeparator("|")
	summaryTable.SetHeader([]string{"Name", "Mean", "Stdev", "P99", "P95", "P90"})
	for scidx := 0; scidx < len(resScenario); scidx++ {
		summaryTable.Append([]string{
			resScenario[scidx].Name,
			fmt.Sprint(time.Duration(resScenario[scidx].Mean)),
			fmt.Sprint(time.Duration(resScenario[scidx].Stdev)),
			fmt.Sprint(time.Duration(resScenario[scidx].P99)),
			fmt.Sprint(time.Duration(resScenario[scidx].P95)),
			fmt.Sprint(time.Duration(resScenario[scidx].P90)),
		})
	}
	summaryTable.Render()
	fmt.Println()
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
	cmd := getProcessCmd(scenario, cfg)
	fmt.Printf("Scenario: %v => %v\n", scenario.Name, cmd.Args)
	fmt.Print("  Warming up")
	res := runScenario(cfg.WarmUpCount, scenario, cfg)
	fmt.Print("  Run")
	res = runScenario(cfg.Count, scenario, cfg)
	fmt.Println()
	mean, err := stats.Mean(res)
	if err != nil {
		fmt.Println(err)
	}
	stdev, err := stats.StandardDeviation(res)
	if err != nil {
		fmt.Println(err)
	}
	p99, err := stats.Percentile(res, 99)
	if err != nil {
		fmt.Println(err)
	}
	p95, err := stats.Percentile(res, 95)
	if err != nil {
		fmt.Println(err)
	}
	p90, err := stats.Percentile(res, 90)
	if err != nil {
		fmt.Println(err)
	}
	return scenarioResult{
		scenario: *scenario,
		Data:     res,
		Mean:     mean,
		Stdev:    stdev,
		P99:      p99,
		P95:      p95,
		P90:      p90,
	}
}

func runScenario(count int, scenario *scenario, cfg *config) []float64 {
	var res []float64
	for i := 0; i < count; i++ {
		exec := timeCmd(getProcessCmd(scenario, cfg))
		res = append(res, exec)
		fmt.Print(".")
	}
	fmt.Println()
	return res
}

func getProcessCmd(scenario *scenario, cfg *config) *exec.Cmd {
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

	var cmd *exec.Cmd
	if len(cmdArguments) > 0 {
		cmd = exec.Command(cmdString, strings.Split(cmdArguments, " ")...)
	} else {
		cmd = exec.Command(cmdString)
	}
	cmd.Dir = workingDirectory
	cmd.Env = cmdEnv
	return cmd
}

func timeCmd(cmd *exec.Cmd) float64 {
	now := time.Now()
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
	return float64(time.Now().Sub(now))
}
