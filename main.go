package main

import (
	"encoding/json"
	"fmt"
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
		Name    string
		Data    []float64
		Average float64
	}
)

func main() {
	jsonFile, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	jsonBytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		fmt.Println(err)
	}

	var cfg config
	err = json.Unmarshal(jsonBytes, &cfg)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Warmup count: %v\n", cfg.WarmUpCount)
	fmt.Printf("Count: %v\n", cfg.Count)
	fmt.Printf("Number of scenarios: %v\n\n", len(cfg.Scenarios))

	var resScenario []scenarioResult
	if cfg.Count > 0 && len(cfg.Scenarios) > 0 {
		for _, scenario := range cfg.Scenarios {
			resScenario = append(resScenario, processScenario(&scenario, &cfg))
		}
	}

	// ****
	fmt.Println("\nResults in ms:\n")
	for scidx := 0; scidx < len(resScenario); scidx++ {
		fmt.Printf(" %v \t", resScenario[scidx].Name)
	}
	fmt.Println()
	for idx := 0; idx < cfg.Count; idx++ {
		for scidx := 0; scidx < len(resScenario); scidx++ {
			fmt.Printf("  %v \t", resScenario[scidx].Data[idx])
		}
		fmt.Println()
	}
	fmt.Println("\n= AVG =")
	for scidx := 0; scidx < len(resScenario); scidx++ {
		fmt.Printf("  %.4f \t", resScenario[scidx].Average)
	}
	fmt.Println()
}

func processScenario(scenario *scenario, cfg *config) scenarioResult {
	cmd := getProcessCmd(scenario, cfg)
	fmt.Printf("Scenario: %v => %v\n", scenario.Name, cmd.Args)
	fmt.Print("  Warming up... ")
	res, avg := runScenario(cfg.WarmUpCount, scenario, cfg)
	fmt.Print("  Run... ")
	res, avg = runScenario(cfg.Count, scenario, cfg)
	fmt.Println()
	return scenarioResult{
		Name:    scenario.Name,
		Data:    res,
		Average: avg,
	}
}

func runScenario(count int, scenario *scenario, cfg *config) ([]float64, float64) {
	var res []float64
	for i := 0; i < count; i++ {
		exec := timeCmd(getProcessCmd(scenario, cfg))
		res = append(res, exec)
	}
	fmt.Printf(" %v = ", res)
	var total float64 = 0
	for _, value := range res {
		total += value
	}
	avg := total / float64(len(res))
	fmt.Println(avg)
	return res, avg
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
	return float64(time.Now().Sub(now).Nanoseconds()) / 1_000_000
}
