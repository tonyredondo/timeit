package main

import (
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func replaceCustomVars(value string) string {
	if currentWorkingDirectory == nil {
		wd, _ := os.Getwd()
		currentWorkingDirectory = &wd
	}
	value = strings.ReplaceAll(value, "$(CWD)", *currentWorkingDirectory)

	return value
}

func resolveWildcard(value string, workingDirOnRelativePath string) []string {
	value = replaceCustomVars(value)
	if !filepath.IsAbs(value) {
		value = filepath.Join(workingDirOnRelativePath, value)
	}
	values, _ := filepath.Glob(value)
	return values
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
