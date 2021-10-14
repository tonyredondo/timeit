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
    Duration: 1.390445894s
  Run ..................................................
    Duration: 6.520697665s

Scenario: CallTarget
  Warming up ..........
    Duration: 1.270528732s
  Run ..................................................
    Duration: 6.394368526s

Scenario: CallTarget+Inlining
  Warming up ..........
    Duration: 1.308253413s
  Run ..................................................
    Duration: 6.409844452s


### Results

|  CALLSITE   | CALLTARGET | CALLTARGET+INLINING |
|-------------|------------|---------------------|
|  129.569ms  | 125.165ms  |      137.792ms      |
|  124.808ms  | 132.932ms  |      125.173ms      |
|  125.416ms  | 132.878ms  |      123.902ms      |
|  132.15ms   | 126.826ms  |      123.685ms      |
|  125.592ms  | 125.429ms  |      125.77ms       |
|  123.276ms  | 127.891ms  |      130.504ms      |
|  126.182ms  | 127.487ms  |      133.925ms      |
|  130.078ms  | 122.943ms  |      124.347ms      |
|  125.842ms  | 124.606ms  |      125.367ms      |
|  131.573ms  | 126.308ms  |      123.41ms       |
|  136.984ms  | 130.747ms  |      123.37ms       |
|  125.796ms  | 126.213ms  |      124.29ms       |
|  126.268ms  |  124.89ms  |      125.259ms      |
|  131.165ms  | 125.863ms  |      125.248ms      |
|  123.263ms  |  123.7ms   |      128.468ms      |
|  130.384ms  | 124.945ms  |      123.389ms      |
|  126.502ms  | 125.433ms  |      124.24ms       |
|  127.123ms  | 124.325ms  |      124.586ms      |
|  128.616ms  | 128.441ms  |      122.694ms      |
|  125.139ms  | 123.991ms  |      122.271ms      |
|  125.637ms  | 126.635ms  |      125.187ms      |
|  126.11ms   | 126.428ms  |      127.693ms      |
|  126.702ms  | 124.989ms  |      127.874ms      |
|  139.174ms  | 122.386ms  |      123.033ms      |
|  126.297ms  | 126.779ms  |      123.805ms      |
|  124.045ms  |  124.08ms  |      122.164ms      |
|  130.46ms   | 125.003ms  |      129.686ms      |
|  132.922ms  | 132.043ms  |      123.357ms      |
|  130.397ms  | 129.094ms  |      125.618ms      |
|  120.883ms  | 133.399ms  |       127.2ms       |
|  127.159ms  | 127.114ms  |      123.603ms      |
|  126.274ms  | 124.352ms  |      128.501ms      |
|  122.927ms  | 124.717ms  |      124.779ms      |
|  127.177ms  | 127.315ms  |      124.859ms      |
|  130.461ms  | 139.018ms  |      122.452ms      |
|  125.033ms  | 135.024ms  |      126.443ms      |
|  124.671ms  | 130.445ms  |      128.769ms      |
|  124.411ms  | 128.692ms  |      124.606ms      |
|  130.343ms  | 128.612ms  |      126.529ms      |
|  122.163ms  | 125.065ms  |      129.012ms      |
|  126.279ms  |  124.65ms  |      124.187ms      |
|  125.405ms  |  124.77ms  |      125.702ms      |
|  130.551ms  | 126.038ms  |      124.699ms      |
|  125.695ms  | 128.479ms  |      124.971ms      |
|  126.862ms  | 127.332ms  |      128.273ms      |
|  125.762ms  | 124.975ms  |      125.716ms      |
|  130.139ms  |  123.91ms  |      124.323ms      |
| 127.43968ms |  124.77ms  |      132.308ms      |
| 127.43968ms | 123.572ms  |    125.896645ms     |
| 127.43968ms | 122.235ms  |    125.896645ms     |

### Outliers

| CALLSITE  | CALLTARGET | CALLTARGET+INLINING |
|-----------|------------|---------------------|
| 145.858ms |     -      |      144.161ms      |
| 151.439ms |     -      |      152.944ms      |
| 184.091ms |     -      |          -          |

### Summary

|                       NAME                       |     MEAN      |   STDDEV   |  STDERR   |      P99      |      P95      |      P90      | OUTLIERS |
|--------------------------------------------------|---------------|------------|-----------|---------------|---------------|---------------|----------|
| Callsite                                         | 127.43968ms   | 3.449743ms | 487.867µs | 138.079ms     | 132.536ms     | 131.165ms     |        3 |
| ├>metric.runtime.dotnet.gc.count.compacting_gen2 |             1 |          0 |         0 |             1 |             1 |             1 |          |
| ├>metric.runtime.dotnet.gc.count.gen0            |            75 |          0 |         0 |            75 |            75 |            75 |          |
| ├>metric.runtime.dotnet.gc.count.gen1            |             0 |          0 |         0 |             0 |             0 |             0 |          |
| ├>metric.runtime.dotnet.gc.count.gen2            |          1277 |          0 |         0 |          1277 |          1277 |          1277 |          |
| ├>metric.runtime.dotnet.gc.memory_load           |            45 |          0 |         0 |            45 |            45 |            45 |          |
| ├>metric.runtime.dotnet.gc.pause_time            |      250.7494 |          0 |         0 |      250.7494 |      250.7494 |      250.7494 |          |
| ├>metric.runtime.dotnet.gc.size.gen0             | 3.591296e+06  |          0 |         0 | 3.591296e+06  | 3.591296e+06  | 3.591296e+06  |          |
| ├>metric.runtime.dotnet.gc.size.gen1             |        656528 |          0 |         0 |        656528 |        656528 |        656528 |          |
| ├>metric.runtime.dotnet.gc.size.gen2             |            24 |          0 |         0 |            24 |            24 |            24 |          |
| ├>metric.runtime.dotnet.gc.size.loh              | 1.2266504e+07 |          0 |         0 | 1.2266504e+07 | 1.2266504e+07 | 1.2266504e+07 |          |
| ├>metric.runtime.dotnet.threads.contention_count |             1 |          0 |         0 |             1 |             1 |             1 |          |
| ├>metric.runtime.dotnet.threads.contention_time  |        1.6882 |          0 |         0 |        1.6882 |        1.6882 |        1.6882 |          |
| ├>metric.runtime.process.private_bytes           | 9.080832e+07  |          0 |         0 | 9.080832e+07  | 9.080832e+07  | 9.080832e+07  |          |
| └>metric.runtime.process.processor_time          |      8765.625 |          0 |         0 |      8765.625 |      8765.625 |      8765.625 |          |
|                                                  |               |            |           |               |               |               |          |
| CallTarget                                       | 126.85868ms   | 3.384404ms | 478.627µs | 137.021ms     | 133.1655ms    | 132.043ms     |        0 |
| ├>metric.runtime.dotnet.gc.count.compacting_gen2 |             1 |          0 |         0 |             1 |             1 |             1 |          |
| ├>metric.runtime.dotnet.gc.count.gen0            |            75 |          0 |         0 |            75 |            75 |            75 |          |
| ├>metric.runtime.dotnet.gc.count.gen1            |             0 |          0 |         0 |             0 |             0 |             0 |          |
| ├>metric.runtime.dotnet.gc.count.gen2            |          1277 |          0 |         0 |          1277 |          1277 |          1277 |          |
| ├>metric.runtime.dotnet.gc.memory_load           |            45 |          0 |         0 |            45 |            45 |            45 |          |
| ├>metric.runtime.dotnet.gc.pause_time            |      250.7494 |          0 |         0 |      250.7494 |      250.7494 |      250.7494 |          |
| ├>metric.runtime.dotnet.gc.size.gen0             | 3.591296e+06  |          0 |         0 | 3.591296e+06  | 3.591296e+06  | 3.591296e+06  |          |
| ├>metric.runtime.dotnet.gc.size.gen1             |        656528 |          0 |         0 |        656528 |        656528 |        656528 |          |
| ├>metric.runtime.dotnet.gc.size.gen2             |            24 |          0 |         0 |            24 |            24 |            24 |          |
| ├>metric.runtime.dotnet.gc.size.loh              | 1.2266504e+07 |          0 |         0 | 1.2266504e+07 | 1.2266504e+07 | 1.2266504e+07 |          |
| ├>metric.runtime.dotnet.threads.contention_count |             1 |          0 |         0 |             1 |             1 |             1 |          |
| ├>metric.runtime.dotnet.threads.contention_time  |        1.6882 |          0 |         0 |        1.6882 |        1.6882 |        1.6882 |          |
| ├>metric.runtime.process.private_bytes           | 9.080832e+07  |          0 |         0 | 9.080832e+07  | 9.080832e+07  | 9.080832e+07  |          |
| └>metric.runtime.process.processor_time          |      8765.625 |          0 |         0 |      8765.625 |      8765.625 |      8765.625 |          |
|                                                  |               |            |           |               |               |               |          |
| CallTarget+Inlining                              | 125.896645ms  | 3.012693ms | 426.059µs | 135.8585ms    | 131.406ms     | 129.012ms     |        2 |
| ├>metric.runtime.dotnet.gc.count.compacting_gen2 |             1 |          0 |         0 |             1 |             1 |             1 |          |
| ├>metric.runtime.dotnet.gc.count.gen0            |            75 |          0 |         0 |            75 |            75 |            75 |          |
| ├>metric.runtime.dotnet.gc.count.gen1            |             0 |          0 |         0 |             0 |             0 |             0 |          |
| ├>metric.runtime.dotnet.gc.count.gen2            |          1277 |          0 |         0 |          1277 |          1277 |          1277 |          |
| ├>metric.runtime.dotnet.gc.memory_load           |            45 |          0 |         0 |            45 |            45 |            45 |          |
| ├>metric.runtime.dotnet.gc.pause_time            |      250.7494 |          0 |         0 |      250.7494 |      250.7494 |      250.7494 |          |
| ├>metric.runtime.dotnet.gc.size.gen0             | 3.591296e+06  |          0 |         0 | 3.591296e+06  | 3.591296e+06  | 3.591296e+06  |          |
| ├>metric.runtime.dotnet.gc.size.gen1             |        656528 |          0 |         0 |        656528 |        656528 |        656528 |          |
| ├>metric.runtime.dotnet.gc.size.gen2             |            24 |          0 |         0 |            24 |            24 |            24 |          |
| ├>metric.runtime.dotnet.gc.size.loh              | 1.2266504e+07 |          0 |         0 | 1.2266504e+07 | 1.2266504e+07 | 1.2266504e+07 |          |
| ├>metric.runtime.dotnet.threads.contention_count |             1 |          0 |         0 |             1 |             1 |             1 |          |
| ├>metric.runtime.dotnet.threads.contention_time  |        1.6882 |          0 |         0 |        1.6882 |        1.6882 |        1.6882 |          |
| ├>metric.runtime.process.private_bytes           | 9.080832e+07  |          0 |         0 | 9.080832e+07  | 9.080832e+07  | 9.080832e+07  |          |
| └>metric.runtime.process.processor_time          |      8765.625 |          0 |         0 |      8765.625 |      8765.625 |      8765.625 |          |
|                                                  |               |            |           |               |               |               |          |
```

### Tables are markdown compatible:

### Summary

|                       NAME                       |     MEAN      |   STDDEV   |  STDERR   |      P99      |      P95      |      P90      | OUTLIERS |
|--------------------------------------------------|---------------|------------|-----------|---------------|---------------|---------------|----------|
| Callsite                                         | 127.43968ms   | 3.449743ms | 487.867µs | 138.079ms     | 132.536ms     | 131.165ms     |        3 |
| ├>metric.runtime.dotnet.gc.count.compacting_gen2 |             1 |          0 |         0 |             1 |             1 |             1 |          |
| ├>metric.runtime.dotnet.gc.count.gen0            |            75 |          0 |         0 |            75 |            75 |            75 |          |
| ├>metric.runtime.dotnet.gc.count.gen1            |             0 |          0 |         0 |             0 |             0 |             0 |          |
| ├>metric.runtime.dotnet.gc.count.gen2            |          1277 |          0 |         0 |          1277 |          1277 |          1277 |          |
| ├>metric.runtime.dotnet.gc.memory_load           |            45 |          0 |         0 |            45 |            45 |            45 |          |
| ├>metric.runtime.dotnet.gc.pause_time            |      250.7494 |          0 |         0 |      250.7494 |      250.7494 |      250.7494 |          |
| ├>metric.runtime.dotnet.gc.size.gen0             | 3.591296e+06  |          0 |         0 | 3.591296e+06  | 3.591296e+06  | 3.591296e+06  |          |
| ├>metric.runtime.dotnet.gc.size.gen1             |        656528 |          0 |         0 |        656528 |        656528 |        656528 |          |
| ├>metric.runtime.dotnet.gc.size.gen2             |            24 |          0 |         0 |            24 |            24 |            24 |          |
| ├>metric.runtime.dotnet.gc.size.loh              | 1.2266504e+07 |          0 |         0 | 1.2266504e+07 | 1.2266504e+07 | 1.2266504e+07 |          |
| ├>metric.runtime.dotnet.threads.contention_count |             1 |          0 |         0 |             1 |             1 |             1 |          |
| ├>metric.runtime.dotnet.threads.contention_time  |        1.6882 |          0 |         0 |        1.6882 |        1.6882 |        1.6882 |          |
| ├>metric.runtime.process.private_bytes           | 9.080832e+07  |          0 |         0 | 9.080832e+07  | 9.080832e+07  | 9.080832e+07  |          |
| └>metric.runtime.process.processor_time          |      8765.625 |          0 |         0 |      8765.625 |      8765.625 |      8765.625 |          |
|                                                  |               |            |           |               |               |               |          |
| CallTarget                                       | 126.85868ms   | 3.384404ms | 478.627µs | 137.021ms     | 133.1655ms    | 132.043ms     |        0 |
| ├>metric.runtime.dotnet.gc.count.compacting_gen2 |             1 |          0 |         0 |             1 |             1 |             1 |          |
| ├>metric.runtime.dotnet.gc.count.gen0            |            75 |          0 |         0 |            75 |            75 |            75 |          |
| ├>metric.runtime.dotnet.gc.count.gen1            |             0 |          0 |         0 |             0 |             0 |             0 |          |
| ├>metric.runtime.dotnet.gc.count.gen2            |          1277 |          0 |         0 |          1277 |          1277 |          1277 |          |
| ├>metric.runtime.dotnet.gc.memory_load           |            45 |          0 |         0 |            45 |            45 |            45 |          |
| ├>metric.runtime.dotnet.gc.pause_time            |      250.7494 |          0 |         0 |      250.7494 |      250.7494 |      250.7494 |          |
| ├>metric.runtime.dotnet.gc.size.gen0             | 3.591296e+06  |          0 |         0 | 3.591296e+06  | 3.591296e+06  | 3.591296e+06  |          |
| ├>metric.runtime.dotnet.gc.size.gen1             |        656528 |          0 |         0 |        656528 |        656528 |        656528 |          |
| ├>metric.runtime.dotnet.gc.size.gen2             |            24 |          0 |         0 |            24 |            24 |            24 |          |
| ├>metric.runtime.dotnet.gc.size.loh              | 1.2266504e+07 |          0 |         0 | 1.2266504e+07 | 1.2266504e+07 | 1.2266504e+07 |          |
| ├>metric.runtime.dotnet.threads.contention_count |             1 |          0 |         0 |             1 |             1 |             1 |          |
| ├>metric.runtime.dotnet.threads.contention_time  |        1.6882 |          0 |         0 |        1.6882 |        1.6882 |        1.6882 |          |
| ├>metric.runtime.process.private_bytes           | 9.080832e+07  |          0 |         0 | 9.080832e+07  | 9.080832e+07  | 9.080832e+07  |          |
| └>metric.runtime.process.processor_time          |      8765.625 |          0 |         0 |      8765.625 |      8765.625 |      8765.625 |          |
|                                                  |               |            |           |               |               |               |          |
| CallTarget+Inlining                              | 125.896645ms  | 3.012693ms | 426.059µs | 135.8585ms    | 131.406ms     | 129.012ms     |        2 |
| ├>metric.runtime.dotnet.gc.count.compacting_gen2 |             1 |          0 |         0 |             1 |             1 |             1 |          |
| ├>metric.runtime.dotnet.gc.count.gen0            |            75 |          0 |         0 |            75 |            75 |            75 |          |
| ├>metric.runtime.dotnet.gc.count.gen1            |             0 |          0 |         0 |             0 |             0 |             0 |          |
| ├>metric.runtime.dotnet.gc.count.gen2            |          1277 |          0 |         0 |          1277 |          1277 |          1277 |          |
| ├>metric.runtime.dotnet.gc.memory_load           |            45 |          0 |         0 |            45 |            45 |            45 |          |
| ├>metric.runtime.dotnet.gc.pause_time            |      250.7494 |          0 |         0 |      250.7494 |      250.7494 |      250.7494 |          |
| ├>metric.runtime.dotnet.gc.size.gen0             | 3.591296e+06  |          0 |         0 | 3.591296e+06  | 3.591296e+06  | 3.591296e+06  |          |
| ├>metric.runtime.dotnet.gc.size.gen1             |        656528 |          0 |         0 |        656528 |        656528 |        656528 |          |
| ├>metric.runtime.dotnet.gc.size.gen2             |            24 |          0 |         0 |            24 |            24 |            24 |          |
| ├>metric.runtime.dotnet.gc.size.loh              | 1.2266504e+07 |          0 |         0 | 1.2266504e+07 | 1.2266504e+07 | 1.2266504e+07 |          |
| ├>metric.runtime.dotnet.threads.contention_count |             1 |          0 |         0 |             1 |             1 |             1 |          |
| ├>metric.runtime.dotnet.threads.contention_time  |        1.6882 |          0 |         0 |        1.6882 |        1.6882 |        1.6882 |          |
| ├>metric.runtime.process.private_bytes           | 9.080832e+07  |          0 |         0 | 9.080832e+07  | 9.080832e+07  | 9.080832e+07  |          |
| └>metric.runtime.process.processor_time          |      8765.625 |          0 |         0 |      8765.625 |      8765.625 |      8765.625 |          |
|                                                  |               |            |           |               |               |               |          |

## Datadog Exporter

Spans for each run are created and sent to datadog backend:

![img.png](img.png)
