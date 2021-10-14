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

|   CALLSITE   |  CALLTARGET  | CALLTARGET+INLINING |
|--------------|--------------|---------------------|
| 119.130035ms | 119.10754ms  |    119.217201ms     |
| 119.238398ms | 125.035465ms |     120.51311ms     |
| 126.918113ms | 120.08652ms  |    119.083756ms     |
| 118.426531ms | 120.170514ms |    122.841455ms     |
| 126.048032ms | 118.535054ms |    119.682682ms     |
| 123.072542ms | 124.199708ms |    119.248027ms     |
| 118.384106ms | 122.606583ms |    126.828431ms     |
| 120.134643ms | 121.041395ms |    121.203776ms     |
| 116.521411ms | 118.906824ms |    120.818347ms     |
| 122.002839ms | 123.853296ms |    118.549224ms     |
| 128.277316ms | 121.762915ms |    119.450521ms     |
| 121.917339ms | 118.628801ms |    122.980554ms     |
| 126.844401ms | 124.679492ms |    121.049312ms     |
| 120.166566ms | 118.367894ms |    121.602012ms     |
| 118.115965ms | 123.022728ms |    124.196663ms     |
| 117.013621ms | 120.456397ms |    139.255776ms     |
| 118.691795ms | 118.651605ms |    140.538136ms     |
| 121.133966ms | 131.439088ms |    121.776168ms     |
| 119.685326ms | 125.850559ms |    121.966693ms     |
| 119.308807ms | 121.585758ms |    119.463395ms     |
| 122.593448ms | 120.399615ms |    125.151463ms     |
| 120.226574ms | 118.120778ms |    120.759333ms     |
| 117.263485ms | 123.414227ms |    127.689951ms     |
| 119.358762ms | 118.84807ms  |    119.649914ms     |
| 118.910029ms |  118.5974ms  |    123.053303ms     |
| 121.079215ms | 124.069845ms |    121.315848ms     |
| 123.73573ms  | 121.635202ms |     124.58714ms     |
| 130.309125ms | 120.175662ms |    121.213297ms     |
| 119.190409ms | 117.187886ms |    120.886893ms     |
| 119.180313ms | 117.956876ms |     126.48195ms     |
| 117.896747ms | 119.853782ms |    131.673977ms     |
| 121.092208ms | 121.460973ms |    122.644634ms     |
| 118.619176ms | 120.181574ms |    122.273389ms     |
| 119.133024ms | 126.056741ms |    119.923673ms     |
| 127.031195ms | 120.264438ms |    124.443369ms     |
| 116.182796ms | 121.980519ms |     119.18143ms     |
| 119.175325ms | 119.643024ms |    120.156021ms     |
| 118.581375ms | 119.462836ms |    127.159352ms     |
| 118.013229ms | 118.023562ms |    129.074734ms     |
| 121.526104ms | 125.721079ms |    122.948787ms     |
| 120.14532ms  | 120.29757ms  |    118.251873ms     |
| 120.057369ms | 118.932892ms |    136.166251ms     |
| 121.336601ms | 124.971402ms |    127.573809ms     |
| 127.508395ms | 122.292931ms |    119.869997ms     |
| 121.212915ms | 120.563003ms |    140.682755ms     |
| 115.390453ms | 121.601226ms |    120.045392ms     |
| 122.039628ms | 118.425654ms |    128.651437ms     |
| 118.746416ms | 122.464719ms |    124.734426ms     |
| 120.761814ms | 118.94632ms  |    123.702086ms     |
| 120.761814ms | 121.977843ms |    123.881871ms     |

### Outliers

|   CALLSITE   | CALLTARGET | CALLTARGET+INLINING |
|--------------|------------|---------------------|
| 151.090684ms |     -      |    165.977812ms     |
| 161.033725ms |     -      |          -          |

### Summary

|                       NAME                        |     MEAN      |   STDDEV   |  STDERR   |      P99      |      P95      |      P90      | OUTLIERS |
|---------------------------------------------------|---------------|------------|-----------|---------------|---------------|---------------|----------|
| Callsite                                          | 120.761814ms  | 3.256263ms | 460.505µs | 129.29322ms   | 127.269795ms  | 126.844401ms  |        2 |
| └─>metric.runtime.dotnet.gc.count.compacting_gen2 |             1 |          0 |         0 |             1 |             1 |             1 |          |
| └─>metric.runtime.dotnet.gc.count.gen0            |            75 |          0 |         0 |            75 |            75 |            75 |          |
| └─>metric.runtime.dotnet.gc.count.gen1            |             0 |          0 |         0 |             0 |             0 |             0 |          |
| └─>metric.runtime.dotnet.gc.count.gen2            |          1277 |          0 |         0 |          1277 |          1277 |          1277 |          |
| └─>metric.runtime.dotnet.gc.memory_load           |            45 |          0 |         0 |            45 |            45 |            45 |          |
| └─>metric.runtime.dotnet.gc.pause_time            |      250.7494 |          0 |         0 |      250.7494 |      250.7494 |      250.7494 |          |
| └─>metric.runtime.dotnet.gc.size.gen0             | 3.591296e+06  |          0 |         0 | 3.591296e+06  | 3.591296e+06  | 3.591296e+06  |          |
| └─>metric.runtime.dotnet.gc.size.gen1             |        656528 |          0 |         0 |        656528 |        656528 |        656528 |          |
| └─>metric.runtime.dotnet.gc.size.gen2             |            24 |          0 |         0 |            24 |            24 |            24 |          |
| └─>metric.runtime.dotnet.gc.size.loh              | 1.2266504e+07 |          0 |         0 | 1.2266504e+07 | 1.2266504e+07 | 1.2266504e+07 |          |
| └─>metric.runtime.dotnet.threads.contention_count |             1 |          0 |         0 |             1 |             1 |             1 |          |
| └─>metric.runtime.dotnet.threads.contention_time  |        1.6882 |          0 |         0 |        1.6882 |        1.6882 |        1.6882 |          |
| └─>metric.runtime.process.private_bytes           | 9.080832e+07  |          0 |         0 | 9.080832e+07  | 9.080832e+07  | 9.080832e+07  |          |
| └─>metric.runtime.process.processor_time          |      8765.625 |          0 |         0 |      8765.625 |      8765.625 |      8765.625 |          |
|                                                   |               |            |           |               |               |               |          |
| CallTarget                                        | 121.230315ms  | 2.757125ms | 389.916µs | 128.747914ms  | 125.785819ms  | 124.971402ms  |        0 |
| └─>metric.runtime.dotnet.gc.count.compacting_gen2 |             1 |          0 |         0 |             1 |             1 |             1 |          |
| └─>metric.runtime.dotnet.gc.count.gen0            |            75 |          0 |         0 |            75 |            75 |            75 |          |
| └─>metric.runtime.dotnet.gc.count.gen1            |             0 |          0 |         0 |             0 |             0 |             0 |          |
| └─>metric.runtime.dotnet.gc.count.gen2            |          1277 |          0 |         0 |          1277 |          1277 |          1277 |          |
| └─>metric.runtime.dotnet.gc.memory_load           |            45 |          0 |         0 |            45 |            45 |            45 |          |
| └─>metric.runtime.dotnet.gc.pause_time            |      250.7494 |          0 |         0 |      250.7494 |      250.7494 |      250.7494 |          |
| └─>metric.runtime.dotnet.gc.size.gen0             | 3.591296e+06  |          0 |         0 | 3.591296e+06  | 3.591296e+06  | 3.591296e+06  |          |
| └─>metric.runtime.dotnet.gc.size.gen1             |        656528 |          0 |         0 |        656528 |        656528 |        656528 |          |
| └─>metric.runtime.dotnet.gc.size.gen2             |            24 |          0 |         0 |            24 |            24 |            24 |          |
| └─>metric.runtime.dotnet.gc.size.loh              | 1.2266504e+07 |          0 |         0 | 1.2266504e+07 | 1.2266504e+07 | 1.2266504e+07 |          |
| └─>metric.runtime.dotnet.threads.contention_count |             1 |          0 |         0 |             1 |             1 |             1 |          |
| └─>metric.runtime.dotnet.threads.contention_time  |        1.6882 |          0 |         0 |        1.6882 |        1.6882 |        1.6882 |          |
| └─>metric.runtime.process.private_bytes           | 9.080832e+07  |          0 |         0 | 9.080832e+07  | 9.080832e+07  | 9.080832e+07  |          |
| └─>metric.runtime.process.processor_time          |      8765.625 |          0 |         0 |      8765.625 |      8765.625 |      8765.625 |          |
|                                                   |               |            |           |               |               |               |          |
| CallTarget+Inlining                               | 123.881871ms  | 5.454844ms | 771.431µs | 140.610445ms  | 137.711013ms  | 129.074734ms  |        1 |
| └─>metric.runtime.dotnet.gc.count.compacting_gen2 |             1 |          0 |         0 |             1 |             1 |             1 |          |
| └─>metric.runtime.dotnet.gc.count.gen0            |            75 |          0 |         0 |            75 |            75 |            75 |          |
| └─>metric.runtime.dotnet.gc.count.gen1            |             0 |          0 |         0 |             0 |             0 |             0 |          |
| └─>metric.runtime.dotnet.gc.count.gen2            |          1277 |          0 |         0 |          1277 |          1277 |          1277 |          |
| └─>metric.runtime.dotnet.gc.memory_load           |            45 |          0 |         0 |            45 |            45 |            45 |          |
| └─>metric.runtime.dotnet.gc.pause_time            |      250.7494 |          0 |         0 |      250.7494 |      250.7494 |      250.7494 |          |
| └─>metric.runtime.dotnet.gc.size.gen0             | 3.591296e+06  |          0 |         0 | 3.591296e+06  | 3.591296e+06  | 3.591296e+06  |          |
| └─>metric.runtime.dotnet.gc.size.gen1             |        656528 |          0 |         0 |        656528 |        656528 |        656528 |          |
| └─>metric.runtime.dotnet.gc.size.gen2             |            24 |          0 |         0 |            24 |            24 |            24 |          |
| └─>metric.runtime.dotnet.gc.size.loh              | 1.2266504e+07 |          0 |         0 | 1.2266504e+07 | 1.2266504e+07 | 1.2266504e+07 |          |
| └─>metric.runtime.dotnet.threads.contention_count |             1 |          0 |         0 |             1 |             1 |             1 |          |
| └─>metric.runtime.dotnet.threads.contention_time  |        1.6882 |          0 |         0 |        1.6882 |        1.6882 |        1.6882 |          |
| └─>metric.runtime.process.private_bytes           | 9.080832e+07  |          0 |         0 | 9.080832e+07  | 9.080832e+07  | 9.080832e+07  |          |
| └─>metric.runtime.process.processor_time          |      8765.625 |          0 |         0 |      8765.625 |      8765.625 |      8765.625 |          |
|                                                   |               |            |           |               |               |               |          |
```

### Tables are markdown compatible:

### Summary

|                       NAME                        |     MEAN      |   STDDEV   |  STDERR   |      P99      |      P95      |      P90      | OUTLIERS |
|---------------------------------------------------|---------------|------------|-----------|---------------|---------------|---------------|----------|
| Callsite                                          | 120.761814ms  | 3.256263ms | 460.505µs | 129.29322ms   | 127.269795ms  | 126.844401ms  |        2 |
| └─>metric.runtime.dotnet.gc.count.compacting_gen2 |             1 |          0 |         0 |             1 |             1 |             1 |          |
| └─>metric.runtime.dotnet.gc.count.gen0            |            75 |          0 |         0 |            75 |            75 |            75 |          |
| └─>metric.runtime.dotnet.gc.count.gen1            |             0 |          0 |         0 |             0 |             0 |             0 |          |
| └─>metric.runtime.dotnet.gc.count.gen2            |          1277 |          0 |         0 |          1277 |          1277 |          1277 |          |
| └─>metric.runtime.dotnet.gc.memory_load           |            45 |          0 |         0 |            45 |            45 |            45 |          |
| └─>metric.runtime.dotnet.gc.pause_time            |      250.7494 |          0 |         0 |      250.7494 |      250.7494 |      250.7494 |          |
| └─>metric.runtime.dotnet.gc.size.gen0             | 3.591296e+06  |          0 |         0 | 3.591296e+06  | 3.591296e+06  | 3.591296e+06  |          |
| └─>metric.runtime.dotnet.gc.size.gen1             |        656528 |          0 |         0 |        656528 |        656528 |        656528 |          |
| └─>metric.runtime.dotnet.gc.size.gen2             |            24 |          0 |         0 |            24 |            24 |            24 |          |
| └─>metric.runtime.dotnet.gc.size.loh              | 1.2266504e+07 |          0 |         0 | 1.2266504e+07 | 1.2266504e+07 | 1.2266504e+07 |          |
| └─>metric.runtime.dotnet.threads.contention_count |             1 |          0 |         0 |             1 |             1 |             1 |          |
| └─>metric.runtime.dotnet.threads.contention_time  |        1.6882 |          0 |         0 |        1.6882 |        1.6882 |        1.6882 |          |
| └─>metric.runtime.process.private_bytes           | 9.080832e+07  |          0 |         0 | 9.080832e+07  | 9.080832e+07  | 9.080832e+07  |          |
| └─>metric.runtime.process.processor_time          |      8765.625 |          0 |         0 |      8765.625 |      8765.625 |      8765.625 |          |
|                                                   |               |            |           |               |               |               |          |
| CallTarget                                        | 121.230315ms  | 2.757125ms | 389.916µs | 128.747914ms  | 125.785819ms  | 124.971402ms  |        0 |
| └─>metric.runtime.dotnet.gc.count.compacting_gen2 |             1 |          0 |         0 |             1 |             1 |             1 |          |
| └─>metric.runtime.dotnet.gc.count.gen0            |            75 |          0 |         0 |            75 |            75 |            75 |          |
| └─>metric.runtime.dotnet.gc.count.gen1            |             0 |          0 |         0 |             0 |             0 |             0 |          |
| └─>metric.runtime.dotnet.gc.count.gen2            |          1277 |          0 |         0 |          1277 |          1277 |          1277 |          |
| └─>metric.runtime.dotnet.gc.memory_load           |            45 |          0 |         0 |            45 |            45 |            45 |          |
| └─>metric.runtime.dotnet.gc.pause_time            |      250.7494 |          0 |         0 |      250.7494 |      250.7494 |      250.7494 |          |
| └─>metric.runtime.dotnet.gc.size.gen0             | 3.591296e+06  |          0 |         0 | 3.591296e+06  | 3.591296e+06  | 3.591296e+06  |          |
| └─>metric.runtime.dotnet.gc.size.gen1             |        656528 |          0 |         0 |        656528 |        656528 |        656528 |          |
| └─>metric.runtime.dotnet.gc.size.gen2             |            24 |          0 |         0 |            24 |            24 |            24 |          |
| └─>metric.runtime.dotnet.gc.size.loh              | 1.2266504e+07 |          0 |         0 | 1.2266504e+07 | 1.2266504e+07 | 1.2266504e+07 |          |
| └─>metric.runtime.dotnet.threads.contention_count |             1 |          0 |         0 |             1 |             1 |             1 |          |
| └─>metric.runtime.dotnet.threads.contention_time  |        1.6882 |          0 |         0 |        1.6882 |        1.6882 |        1.6882 |          |
| └─>metric.runtime.process.private_bytes           | 9.080832e+07  |          0 |         0 | 9.080832e+07  | 9.080832e+07  | 9.080832e+07  |          |
| └─>metric.runtime.process.processor_time          |      8765.625 |          0 |         0 |      8765.625 |      8765.625 |      8765.625 |          |
|                                                   |               |            |           |               |               |               |          |
| CallTarget+Inlining                               | 123.881871ms  | 5.454844ms | 771.431µs | 140.610445ms  | 137.711013ms  | 129.074734ms  |        1 |
| └─>metric.runtime.dotnet.gc.count.compacting_gen2 |             1 |          0 |         0 |             1 |             1 |             1 |          |
| └─>metric.runtime.dotnet.gc.count.gen0            |            75 |          0 |         0 |            75 |            75 |            75 |          |
| └─>metric.runtime.dotnet.gc.count.gen1            |             0 |          0 |         0 |             0 |             0 |             0 |          |
| └─>metric.runtime.dotnet.gc.count.gen2            |          1277 |          0 |         0 |          1277 |          1277 |          1277 |          |
| └─>metric.runtime.dotnet.gc.memory_load           |            45 |          0 |         0 |            45 |            45 |            45 |          |
| └─>metric.runtime.dotnet.gc.pause_time            |      250.7494 |          0 |         0 |      250.7494 |      250.7494 |      250.7494 |          |
| └─>metric.runtime.dotnet.gc.size.gen0             | 3.591296e+06  |          0 |         0 | 3.591296e+06  | 3.591296e+06  | 3.591296e+06  |          |
| └─>metric.runtime.dotnet.gc.size.gen1             |        656528 |          0 |         0 |        656528 |        656528 |        656528 |          |
| └─>metric.runtime.dotnet.gc.size.gen2             |            24 |          0 |         0 |            24 |            24 |            24 |          |
| └─>metric.runtime.dotnet.gc.size.loh              | 1.2266504e+07 |          0 |         0 | 1.2266504e+07 | 1.2266504e+07 | 1.2266504e+07 |          |
| └─>metric.runtime.dotnet.threads.contention_count |             1 |          0 |         0 |             1 |             1 |             1 |          |
| └─>metric.runtime.dotnet.threads.contention_time  |        1.6882 |          0 |         0 |        1.6882 |        1.6882 |        1.6882 |          |
| └─>metric.runtime.process.private_bytes           | 9.080832e+07  |          0 |         0 | 9.080832e+07  | 9.080832e+07  | 9.080832e+07  |          |
| └─>metric.runtime.process.processor_time          |      8765.625 |          0 |         0 |      8765.625 |      8765.625 |      8765.625 |          |
|                                                   |               |            |           |               |               |               |          |

## Datadog Exporter

Spans for each run are created and sent to datadog backend:

![img.png](img.png)
