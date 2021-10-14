# timeit
Command execution time meter allows to configure multiple scenarios to run benchmarks over CLI apps, output are available in markdown and as datadog traces.

## Installation

```bash
go install github.com/tonyredondo/timeit@latest
```

### Usage
```bash
timeit [configuration file.json]
```

## Sample Configuration

```json
{
  "enableDatadog": true,
  "warmUpCount": 10,
  "count": 50,
  "scenarios": [
    {
      "name": "Callsite",
      "environmentVariables": {
        "DD_TRACE_CALLTARGET_ENABLED": "false",
        "DD_CLR_ENABLE_INLINING": "false"
      }
    },
    {
      "name": "CallTarget",
      "environmentVariables": {
        "DD_TRACE_CALLTARGET_ENABLED": "true",
        "DD_CLR_ENABLE_INLINING": "false"
      }
    },
    {
      "name": "CallTarget\u002BInlining",
      "environmentVariables": {
        "DD_TRACE_CALLTARGET_ENABLED": "true",
        "DD_CLR_ENABLE_INLINING": "true"
      }
    }
  ],
  "processName": "dotnet",
  "processArguments": "--version",
  "processTimeout": 15,
  "environmentVariables": {
    "CORECLR_ENABLE_PROFILING": "1",
    "CORECLR_PROFILER": "{846F5F1C-F9AE-4B07-969E-05C26BC060D8}",
    "CORECLR_PROFILER_PATH": "/Datadog.Trace.ClrProfiler.Native.dylib",
    "DD_DOTNET_TRACER_HOME": "/",
    "DD_INTEGRATIONS": "/integrations.json"
  },
  "tags": {
    "runtime.architecture" : "x86",
    "runtime.name" : ".NET Framework",
    "runtime.version" : "4.6.1",
    "benchmark.job.runtime.name" : ".NET Framework 4.6.1",
    "benchmark.job.runtime.moniker" : "net461"
  },
  "metricsFilePath": "metrics.json",
  "timeout" : {
    "maxDuration": 15,
    "processName": "dotnet-dump",
    "processArguments": "collect --process-id %pid%"
  }
}
```

## Sample output

```bash
C:\github\timeit>go run main.go config_example.json
TimeIt by Tony Redondo

Warmup count: 10
Count: 50
Number of scenarios: 3

Scenario: Callsite
  Warming up ..........
    Duration: 1.215486747s
  Run ..................................................
    Duration: 6.154432435s

Scenario: CallTarget
  Warming up ..........
    Duration: 1.299867951s
  Run ..................................................
    Duration: 6.109837042s

Scenario: CallTarget+Inlining
  Warming up ..........
    Duration: 1.240914529s
  Run ..................................................
    Duration: 6.288758182s

### Results

| CALLSITE  |  CALLTARGET  | CALLTARGET+INLINING |
|-----------|--------------|---------------------|
| 128.034ms |  125.864ms   |      126.914ms      |
| 138.707ms |  126.272ms   |      127.021ms      |
| 130.258ms |  125.752ms   |      124.386ms      |
| 127.361ms |  125.083ms   |      127.811ms      |
| 125.851ms |  126.789ms   |      130.488ms      |
| 127.766ms |  127.596ms   |      125.745ms      |
| 131.872ms |   129.95ms   |      124.737ms      |
| 125.591ms |  141.593ms   |      133.974ms      |
| 123.506ms |  128.277ms   |      137.706ms      |
| 130.39ms  |   126.88ms   |      144.609ms      |
|  127.3ms  |  126.483ms   |      143.98ms       |
| 133.602ms |  125.806ms   |      142.537ms      |
| 127.134ms |   126.63ms   |      162.578ms      |
| 124.488ms |   131.37ms   |      134.155ms      |
| 134.913ms |  125.712ms   |      138.437ms      |
| 124.473ms |  131.456ms   |      138.875ms      |
| 135.563ms |  130.313ms   |      141.265ms      |
| 140.098ms |  133.843ms   |      142.847ms      |
| 141.159ms |  125.109ms   |      136.943ms      |
| 134.457ms |  127.016ms   |      135.778ms      |
| 130.949ms |  123.801ms   |      133.704ms      |
| 130.772ms |  131.163ms   |      140.031ms      |
| 138.526ms |  129.056ms   |      138.047ms      |
| 124.77ms  |  137.756ms   |      124.335ms      |
| 132.375ms |  135.918ms   |      132.817ms      |
| 126.138ms |  126.635ms   |      128.572ms      |
| 131.648ms |   128.45ms   |      127.273ms      |
| 130.234ms |  133.968ms   |      126.095ms      |
| 125.204ms |  129.032ms   |      130.77ms       |
| 129.563ms |  125.773ms   |      131.315ms      |
| 125.14ms  |  131.564ms   |      127.136ms      |
| 125.721ms |  146.129ms   |      128.173ms      |
| 127.555ms |  130.744ms   |      143.971ms      |
| 122.78ms  |  136.311ms   |      129.546ms      |
| 125.692ms |  126.496ms   |      129.619ms      |
|  131.7ms  |  129.541ms   |      138.907ms      |
| 131.013ms |  124.613ms   |      133.643ms      |
| 149.901ms |  127.396ms   |      133.793ms      |
| 130.009ms |   141.9ms    |      127.747ms      |
| 123.178ms |  131.656ms   |      128.96ms       |
| 125.873ms |  123.407ms   |      136.835ms      |
| 123.978ms |  128.807ms   |      128.943ms      |
| 124.097ms |  127.554ms   |      125.688ms      |
| 136.379ms |  145.953ms   |      132.308ms      |
| 130.106ms |  129.138ms   |      128.701ms      |
| 126.853ms |  154.454ms   |      130.537ms      |
| 127.426ms |  135.273ms   |      140.839ms      |
| 125.552ms | 130.644297ms |      146.52ms       |
| 127.945ms | 130.644297ms |      136.477ms      |
| 124.243ms | 130.644297ms |      127.21ms       |

### Outliers

| CALLSITE | CALLTARGET | CALLTARGET+INLINING |
|----------|------------|---------------------|
|    -     | 159.336ms  |          -          |
|    -     | 167.635ms  |          -          |
|    -     | 177.363ms  |          -          |

### Summary

|                       NAME                       |     MEAN      |   STDDEV   |   STDERR   |      P99      |      P95      |      P90      | OUTLIERS |
|--------------------------------------------------|---------------|------------|------------|---------------|---------------|---------------|----------|
| Callsite                                         | 129.55686ms   | 5.423834ms | 767.045µs  | 145.53ms      | 139.4025ms    | 136.379ms     |        0 |
| └>metric.runtime.dotnet.gc.count.compacting_gen2 |             1 |          0 |          0 |             1 |             1 |             1 |          |
| └>metric.runtime.dotnet.gc.count.gen0            |            75 |          0 |          0 |            75 |            75 |            75 |          |
| └>metric.runtime.dotnet.gc.count.gen1            |             0 |          0 |          0 |             0 |             0 |             0 |          |
| └>metric.runtime.dotnet.gc.count.gen2            |          1277 |          0 |          0 |          1277 |          1277 |          1277 |          |
| └>metric.runtime.dotnet.gc.memory_load           |            45 |          0 |          0 |            45 |            45 |            45 |          |
| └>metric.runtime.dotnet.gc.pause_time            |      250.7494 |          0 |          0 |      250.7494 |      250.7494 |      250.7494 |          |
| └>metric.runtime.dotnet.gc.size.gen0             | 3.591296e+06  |          0 |          0 | 3.591296e+06  | 3.591296e+06  | 3.591296e+06  |          |
| └>metric.runtime.dotnet.gc.size.gen1             |        656528 |          0 |          0 |        656528 |        656528 |        656528 |          |
| └>metric.runtime.dotnet.gc.size.gen2             |            24 |          0 |          0 |            24 |            24 |            24 |          |
| └>metric.runtime.dotnet.gc.size.loh              | 1.2266504e+07 |          0 |          0 | 1.2266504e+07 | 1.2266504e+07 | 1.2266504e+07 |          |
| └>metric.runtime.dotnet.threads.contention_count |             1 |          0 |          0 |             1 |             1 |             1 |          |
| └>metric.runtime.dotnet.threads.contention_time  |        1.6882 |          0 |          0 |        1.6882 |        1.6882 |        1.6882 |          |
| └>metric.runtime.process.private_bytes           | 9.080832e+07  |          0 |          0 | 9.080832e+07  | 9.080832e+07  | 9.080832e+07  |          |
| └>metric.runtime.process.processor_time          |      8765.625 |          0 |          0 |      8765.625 |      8765.625 |      8765.625 |          |
|                                                  |               |            |            |               |               |               |          |
| CallTarget                                       | 130.644297ms  | 6.228898ms | 880.899µs  | 150.2915ms    | 143.9265ms    | 137.756ms     |        3 |
| └>metric.runtime.dotnet.gc.count.compacting_gen2 |             1 |          0 |          0 |             1 |             1 |             1 |          |
| └>metric.runtime.dotnet.gc.count.gen0            |            75 |          0 |          0 |            75 |            75 |            75 |          |
| └>metric.runtime.dotnet.gc.count.gen1            |             0 |          0 |          0 |             0 |             0 |             0 |          |
| └>metric.runtime.dotnet.gc.count.gen2            |          1277 |          0 |          0 |          1277 |          1277 |          1277 |          |
| └>metric.runtime.dotnet.gc.memory_load           |            45 |          0 |          0 |            45 |            45 |            45 |          |
| └>metric.runtime.dotnet.gc.pause_time            |      250.7494 |          0 |          0 |      250.7494 |      250.7494 |      250.7494 |          |
| └>metric.runtime.dotnet.gc.size.gen0             | 3.591296e+06  |          0 |          0 | 3.591296e+06  | 3.591296e+06  | 3.591296e+06  |          |
| └>metric.runtime.dotnet.gc.size.gen1             |        656528 |          0 |          0 |        656528 |        656528 |        656528 |          |
| └>metric.runtime.dotnet.gc.size.gen2             |            24 |          0 |          0 |            24 |            24 |            24 |          |
| └>metric.runtime.dotnet.gc.size.loh              | 1.2266504e+07 |          0 |          0 | 1.2266504e+07 | 1.2266504e+07 | 1.2266504e+07 |          |
| └>metric.runtime.dotnet.threads.contention_count |             1 |          0 |          0 |             1 |             1 |             1 |          |
| └>metric.runtime.dotnet.threads.contention_time  |        1.6882 |          0 |          0 |        1.6882 |        1.6882 |        1.6882 |          |
| └>metric.runtime.process.private_bytes           | 9.080832e+07  |          0 |          0 | 9.080832e+07  | 9.080832e+07  | 9.080832e+07  |          |
| └>metric.runtime.process.processor_time          |      8765.625 |          0 |          0 |      8765.625 |      8765.625 |      8765.625 |          |
|                                                  |               |            |            |               |               |               |          |
| CallTarget+Inlining                              | 133.78596ms   | 7.358837ms | 1.040696ms | 154.549ms     | 144.2945ms    | 142.847ms     |        0 |
| └>metric.runtime.dotnet.gc.count.compacting_gen2 |             1 |          0 |          0 |             1 |             1 |             1 |          |
| └>metric.runtime.dotnet.gc.count.gen0            |            75 |          0 |          0 |            75 |            75 |            75 |          |
| └>metric.runtime.dotnet.gc.count.gen1            |             0 |          0 |          0 |             0 |             0 |             0 |          |
| └>metric.runtime.dotnet.gc.count.gen2            |          1277 |          0 |          0 |          1277 |          1277 |          1277 |          |
| └>metric.runtime.dotnet.gc.memory_load           |            45 |          0 |          0 |            45 |            45 |            45 |          |
| └>metric.runtime.dotnet.gc.pause_time            |      250.7494 |          0 |          0 |      250.7494 |      250.7494 |      250.7494 |          |
| └>metric.runtime.dotnet.gc.size.gen0             | 3.591296e+06  |          0 |          0 | 3.591296e+06  | 3.591296e+06  | 3.591296e+06  |          |
| └>metric.runtime.dotnet.gc.size.gen1             |        656528 |          0 |          0 |        656528 |        656528 |        656528 |          |
| └>metric.runtime.dotnet.gc.size.gen2             |            24 |          0 |          0 |            24 |            24 |            24 |          |
| └>metric.runtime.dotnet.gc.size.loh              | 1.2266504e+07 |          0 |          0 | 1.2266504e+07 | 1.2266504e+07 | 1.2266504e+07 |          |
| └>metric.runtime.dotnet.threads.contention_count |             1 |          0 |          0 |             1 |             1 |             1 |          |
| └>metric.runtime.dotnet.threads.contention_time  |        1.6882 |          0 |          0 |        1.6882 |        1.6882 |        1.6882 |          |
| └>metric.runtime.process.private_bytes           | 9.080832e+07  |          0 |          0 | 9.080832e+07  | 9.080832e+07  | 9.080832e+07  |          |
| └>metric.runtime.process.processor_time          |      8765.625 |          0 |          0 |      8765.625 |      8765.625 |      8765.625 |          |
|                                                  |               |            |            |               |               |               |          |
```

### Tables are markdown compatible:

### Summary

|                       NAME                       |     MEAN      |   STDDEV   |   STDERR   |      P99      |      P95      |      P90      | OUTLIERS |
|--------------------------------------------------|---------------|------------|------------|---------------|---------------|---------------|----------|
| Callsite                                         | 129.55686ms   | 5.423834ms | 767.045µs  | 145.53ms      | 139.4025ms    | 136.379ms     |        0 |
| └>metric.runtime.dotnet.gc.count.compacting_gen2 |             1 |          0 |          0 |             1 |             1 |             1 |          |
| └>metric.runtime.dotnet.gc.count.gen0            |            75 |          0 |          0 |            75 |            75 |            75 |          |
| └>metric.runtime.dotnet.gc.count.gen1            |             0 |          0 |          0 |             0 |             0 |             0 |          |
| └>metric.runtime.dotnet.gc.count.gen2            |          1277 |          0 |          0 |          1277 |          1277 |          1277 |          |
| └>metric.runtime.dotnet.gc.memory_load           |            45 |          0 |          0 |            45 |            45 |            45 |          |
| └>metric.runtime.dotnet.gc.pause_time            |      250.7494 |          0 |          0 |      250.7494 |      250.7494 |      250.7494 |          |
| └>metric.runtime.dotnet.gc.size.gen0             | 3.591296e+06  |          0 |          0 | 3.591296e+06  | 3.591296e+06  | 3.591296e+06  |          |
| └>metric.runtime.dotnet.gc.size.gen1             |        656528 |          0 |          0 |        656528 |        656528 |        656528 |          |
| └>metric.runtime.dotnet.gc.size.gen2             |            24 |          0 |          0 |            24 |            24 |            24 |          |
| └>metric.runtime.dotnet.gc.size.loh              | 1.2266504e+07 |          0 |          0 | 1.2266504e+07 | 1.2266504e+07 | 1.2266504e+07 |          |
| └>metric.runtime.dotnet.threads.contention_count |             1 |          0 |          0 |             1 |             1 |             1 |          |
| └>metric.runtime.dotnet.threads.contention_time  |        1.6882 |          0 |          0 |        1.6882 |        1.6882 |        1.6882 |          |
| └>metric.runtime.process.private_bytes           | 9.080832e+07  |          0 |          0 | 9.080832e+07  | 9.080832e+07  | 9.080832e+07  |          |
| └>metric.runtime.process.processor_time          |      8765.625 |          0 |          0 |      8765.625 |      8765.625 |      8765.625 |          |
|                                                  |               |            |            |               |               |               |          |
| CallTarget                                       | 130.644297ms  | 6.228898ms | 880.899µs  | 150.2915ms    | 143.9265ms    | 137.756ms     |        3 |
| └>metric.runtime.dotnet.gc.count.compacting_gen2 |             1 |          0 |          0 |             1 |             1 |             1 |          |
| └>metric.runtime.dotnet.gc.count.gen0            |            75 |          0 |          0 |            75 |            75 |            75 |          |
| └>metric.runtime.dotnet.gc.count.gen1            |             0 |          0 |          0 |             0 |             0 |             0 |          |
| └>metric.runtime.dotnet.gc.count.gen2            |          1277 |          0 |          0 |          1277 |          1277 |          1277 |          |
| └>metric.runtime.dotnet.gc.memory_load           |            45 |          0 |          0 |            45 |            45 |            45 |          |
| └>metric.runtime.dotnet.gc.pause_time            |      250.7494 |          0 |          0 |      250.7494 |      250.7494 |      250.7494 |          |
| └>metric.runtime.dotnet.gc.size.gen0             | 3.591296e+06  |          0 |          0 | 3.591296e+06  | 3.591296e+06  | 3.591296e+06  |          |
| └>metric.runtime.dotnet.gc.size.gen1             |        656528 |          0 |          0 |        656528 |        656528 |        656528 |          |
| └>metric.runtime.dotnet.gc.size.gen2             |            24 |          0 |          0 |            24 |            24 |            24 |          |
| └>metric.runtime.dotnet.gc.size.loh              | 1.2266504e+07 |          0 |          0 | 1.2266504e+07 | 1.2266504e+07 | 1.2266504e+07 |          |
| └>metric.runtime.dotnet.threads.contention_count |             1 |          0 |          0 |             1 |             1 |             1 |          |
| └>metric.runtime.dotnet.threads.contention_time  |        1.6882 |          0 |          0 |        1.6882 |        1.6882 |        1.6882 |          |
| └>metric.runtime.process.private_bytes           | 9.080832e+07  |          0 |          0 | 9.080832e+07  | 9.080832e+07  | 9.080832e+07  |          |
| └>metric.runtime.process.processor_time          |      8765.625 |          0 |          0 |      8765.625 |      8765.625 |      8765.625 |          |
|                                                  |               |            |            |               |               |               |          |
| CallTarget+Inlining                              | 133.78596ms   | 7.358837ms | 1.040696ms | 154.549ms     | 144.2945ms    | 142.847ms     |        0 |
| └>metric.runtime.dotnet.gc.count.compacting_gen2 |             1 |          0 |          0 |             1 |             1 |             1 |          |
| └>metric.runtime.dotnet.gc.count.gen0            |            75 |          0 |          0 |            75 |            75 |            75 |          |
| └>metric.runtime.dotnet.gc.count.gen1            |             0 |          0 |          0 |             0 |             0 |             0 |          |
| └>metric.runtime.dotnet.gc.count.gen2            |          1277 |          0 |          0 |          1277 |          1277 |          1277 |          |
| └>metric.runtime.dotnet.gc.memory_load           |            45 |          0 |          0 |            45 |            45 |            45 |          |
| └>metric.runtime.dotnet.gc.pause_time            |      250.7494 |          0 |          0 |      250.7494 |      250.7494 |      250.7494 |          |
| └>metric.runtime.dotnet.gc.size.gen0             | 3.591296e+06  |          0 |          0 | 3.591296e+06  | 3.591296e+06  | 3.591296e+06  |          |
| └>metric.runtime.dotnet.gc.size.gen1             |        656528 |          0 |          0 |        656528 |        656528 |        656528 |          |
| └>metric.runtime.dotnet.gc.size.gen2             |            24 |          0 |          0 |            24 |            24 |            24 |          |
| └>metric.runtime.dotnet.gc.size.loh              | 1.2266504e+07 |          0 |          0 | 1.2266504e+07 | 1.2266504e+07 | 1.2266504e+07 |          |
| └>metric.runtime.dotnet.threads.contention_count |             1 |          0 |          0 |             1 |             1 |             1 |          |
| └>metric.runtime.dotnet.threads.contention_time  |        1.6882 |          0 |          0 |        1.6882 |        1.6882 |        1.6882 |          |
| └>metric.runtime.process.private_bytes           | 9.080832e+07  |          0 |          0 | 9.080832e+07  | 9.080832e+07  | 9.080832e+07  |          |
| └>metric.runtime.process.processor_time          |      8765.625 |          0 |          0 |      8765.625 |      8765.625 |      8765.625 |          |
|                                                  |               |            |            |               |               |               |          |

## Datadog Exporter

Spans for each run are created and sent to datadog backend:

![img.png](img.png)
