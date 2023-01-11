package main

import (
	"context"
	"fmt"

	ddtesting "github.com/DataDog/dd-sdk-go-testing"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type datadogExporter struct {
	configuration *config
}

func newDatadogExporter() Exporter {
	return new(datadogExporter)
}

func (de *datadogExporter) SetConfiguration(configuration *config) {
	de.configuration = configuration
}

func (de *datadogExporter) IsEnabled() bool {
	return de.configuration.EnableDatadog
}

func (de *datadogExporter) Export(resScenario []scenarioResult) {
	if len(resScenario) > 0 {
		finalizer := ddtesting.Initialize(tracer.WithLogger(myLogger{}), tracer.WithAnalytics(true))
		defer finalizer()

		cfg := de.configuration

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
			startSpanOptions = append(startSpanOptions, tracer.StartTime(scenario.Start))
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
					ddtesting.WithFinishOptions(tracer.FinishTime(scenario.End), tracer.WithError(scenario.Error)),
				},
			})

			testFinish(ddtesting.StatusPass, nil)
		}
	}
}
