package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
)

type jsonExporter struct {
	configuration *config
}

func newJsonExporter() Exporter {
	return new(jsonExporter)
}

func (de *jsonExporter) SetConfiguration(configuration *config) {
	de.configuration = configuration
}

func (de *jsonExporter) IsEnabled() bool {
	return true
}

func (de *jsonExporter) Export(resScenario []scenarioResult) {
	jsonData, err := json.MarshalIndent(resScenario, "", "  ")
	if err != nil {
		fmt.Printf("Error exporting to json: %v\n", err)
		return
	}

	outputFile := de.configuration.JsonExporterFilePath
	if outputFile == "" {
		outputFile = filepath.Join(de.configuration.Path, fmt.Sprintf("jsonexporter_%d.json", rand.Int()))
	}
	err = os.WriteFile(outputFile, jsonData, os.ModePerm)
	if err != nil {
		fmt.Printf("Error exporting to json: %v\n", err)
		return
	}

	fmt.Printf("The Json file '%s' was exported.\n", outputFile)
}
