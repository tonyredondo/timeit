package main

type Exporter interface {
	SetConfiguration(configuration *config)
	IsEnabled() bool
	Export(resScenario []scenarioResult)
}
