package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/montanaflynn/stats"
	"github.com/olekukonko/tablewriter"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"
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
		Data  []time.Duration
		Mean  time.Duration
		Stdev time.Duration
		P99   time.Duration
		P90   time.Duration
	}
)

func main() {
	cfg, err := loadConfiguration()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Warmup count: %v\n", cfg.WarmUpCount)
	fmt.Printf("Count: %v\n", cfg.Count)
	fmt.Printf("Number of scenarios: %v\n\n", len(cfg.Scenarios))

	var resScenario []scenarioResult
	if cfg.Count > 0 && len(cfg.Scenarios) > 0 {
		for _, scenario := range cfg.Scenarios {
			resScenario = append(resScenario, processScenario(&scenario, cfg))
		}
	}

	// ****
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
			resultRow = append(resultRow, fmt.Sprint(resScenario[scidx].Data[idx]))
		}
		resultTable.Append(resultRow)
	}
	resultTable.Render()

	fmt.Println("\n### Summary\n")
	summaryTable := tablewriter.NewWriter(os.Stdout)
	summaryTable.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	summaryTable.SetCenterSeparator("|")
	summaryTable.SetHeader([]string{"Name", "Mean", "Stdev", "P99", "P90"})
	for scidx := 0; scidx < len(resScenario); scidx++ {
		summaryTable.Append([]string{
			resScenario[scidx].Name,
			fmt.Sprint(resScenario[scidx].Mean),
			fmt.Sprint(resScenario[scidx].Stdev),
			fmt.Sprint(resScenario[scidx].P99),
			fmt.Sprint(resScenario[scidx].P90),
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
	fmt.Print("  Warming up... ")
	resDuration, resFloat := runScenario(cfg.WarmUpCount, scenario, cfg)
	fmt.Print("  Run... ")
	resDuration, resFloat = runScenario(cfg.Count, scenario, cfg)
	fmt.Println()
	mean, err := stats.Mean(resFloat)
	if err != nil {
		fmt.Println(err)
	}
	stdev, err := stats.StandardDeviation(resFloat)
	if err != nil {
		fmt.Println(err)
	}
	p99, err := stats.Percentile(resFloat, 99)
	if err != nil {
		fmt.Println(err)
	}
	p90, err := stats.Percentile(resFloat, 90)
	if err != nil {
		fmt.Println(err)
	}
	return scenarioResult{
		scenario: *scenario,
		Data:     resDuration,
		Mean:     time.Duration(mean),
		Stdev:    time.Duration(stdev),
		P99:      time.Duration(p99),
		P90:      time.Duration(p90),
	}
}

func runScenario(count int, scenario *scenario, cfg *config) ([]time.Duration, []float64) {
	var resDuration []time.Duration
	var resFloat []float64
	for i := 0; i < count; i++ {
		exec := timeCmd(getProcessCmd(scenario, cfg))
		resDuration = append(resDuration, exec)
		resFloat = append(resFloat, float64(exec))
	}
	fmt.Printf(" %v\n", resDuration)
	return resDuration, resFloat
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

func timeCmd(cmd *exec.Cmd) time.Duration {
	now := time.Now()
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
	return time.Now().Sub(now)
}
