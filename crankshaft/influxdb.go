package crankshaft

import (
	"github.com/influxdb/influxdb-go"
	"log"
	"strconv"
	"strings"
	"time"
)

type influxdbBackend struct {
	*influxdb.Client
}

func GetInfluxClient() *influxdbBackend {
	backend := config.InfluxDB.Host + ":" + strconv.Itoa(config.InfluxDB.Port)
	username := config.InfluxDB.Username
	password := config.InfluxDB.Password
	database := config.InfluxDB.Database

	log.Println("Opening InfluxDB Backend to", backend, "user:", username, "database:", database)

	client, err := influxdb.NewClient(&influxdb.ClientConfig{
		Host:     backend,
		Username: username,
		Password: password,
		Database: database,
	})
	if err != nil {
		log.Println("Error creating InfluxDB client")
	}

	return &influxdbBackend{client}
}

func (client *influxdbBackend) WriteEvent(event *TurbineEvent) {
	name := event.data["name"].(string)
	resourceType := event.data["type"].(string)

	series := []*influxdb.Series{}
	now := getCurrentTime()
	for k, v := range event.data {
		// This are the only properties we want per command/pool.
		if !strings.HasPrefix(k, "rollingCount") && !strings.HasPrefix(k, "current") &&
			!strings.HasPrefix(k, "isCircuitBreakerOpen") && !strings.HasPrefix(k, "latencyTotal") {
			continue
		}

		statKey := buildStatKey(event.clusterName, name, resourceType, k)

		switch v := v.(type) {
		default:
			log.Printf("unexpected data element %T, %s", v, v)
		case string:
			// ignored
		case map[string]interface{}:
			// for pct, val := range v {
			//      client.Gauge(statKey+"."+strings.Replace(pct, ".", "_", -1)+"_pct", int64(val.(float64)), 1.0)
			// }
		case bool:
			// if v {
			//      client.Gauge(statKey, 1, 1.0)
			// } else {
			//      client.Gauge(statKey, 0, 1.0)
			// }
		case int64:
			// client.Gauge(statKey, v, 1.0)
		case float64:
			// client.Gauge(statKey, int64(v), 1.0)
			series = append(series, &influxdb.Series{
				Name:    statKey,
				Columns: []string{"time", "value"},
				Points:  [][]interface{}{{now, v}},
			})
		}
	}

	if err := client.WriteSeries(series); err != nil {
		log.Println(err)
	}
}

func buildStatKey(clusterName string, name string, resourceType string, key string) string {
	resourceType = strings.TrimPrefix(strings.ToLower(resourceType), "hystrix")

	if name != "meta" {
		return clusterName + "." + resourceType + "." + name + "." + key
	} else {
		return clusterName + "." + name + "." + key
	}
}

func getCurrentTime() int64 {
	return time.Now().UnixNano() / 1000000
}
