package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
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
)

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
