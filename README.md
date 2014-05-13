crankshaftd
===========

Simple Go agent to ingest streaming data from Turbine via SSE and push it into StatsD as a gauge or to InfluxDB.


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
