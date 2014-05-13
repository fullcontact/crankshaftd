crankshaftd
===========

Simple Go agent to ingest streaming data from Turbine via SSE and push it into StatsD as a gauge or to InfluxDB as a series.

## Property Munging

All metrics related to properties (such as max threadpool size, queue lengths, etc.) are stripped by default.

Properties of the following patterns are kept:

-current*
-rollingCount*
-latencyTotal_*
-isCircuitBreakerOpen

Normal keys are pushed under `my-prefix.my-cluster.MyCommandImpl.rollingCountSuccess`
Latencie histograms are pushed under `my-prefix.my-cluster.MyCommandImpl.latencyTotal.xx_pct`

isCircuitBreakerOpen will be either 0 or 1.

## StatsD caveat

All metric values are cast from float64 -> int64. The InfluxDB backend does not have this limitation.

## Running

Create a config.toml with your config

```toml
# Turbine connection information
host = "turbine.yourcompany.com"
port = 80
tls_enabled = false
clusters = ["cluster1", "cluster2"]

# Backend to stream updates to, 'influxdb' or 'statsd'
backendtype = "influxdb"

[statsd]
host = "127.0.0.1"
port = 8125
prefix = "hystrix"

[influxdb]
host = "127.0.0.1"
port = 8086
username = "crankshaft"
password = "test"
database = "crankshaft"
```

Build the project

    go build
  
And deploy the resulting binary & config

    ./crankshaftd
    
Or run with a different config

    ./crankshaftd dev.toml
