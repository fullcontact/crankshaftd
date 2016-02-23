package crankshaft

import (
	"github.com/cactus/go-statsd-client/statsd"
	"log"
	"strconv"
	"strings"
)

type statsdBackend struct {
	*statsd.Client
}

func GetStatsClient() *statsdBackend {
	backend := config.Statsd.Host + ":" + strconv.Itoa(config.Statsd.Port)
	prefix := config.Statsd.Prefix

	log.Println("Opening StatsD Backend to", backend, "prefix:", prefix)

	client, err := statsd.New(backend, prefix)
	if err != nil {
		log.Println("Error creating StatsD client")
	}

	return &statsdBackend{client}
}

func (client *statsdBackend) WriteEvent(event *TurbineEvent) {
	name := event.data["name"].(string)
	resourceType := event.data["type"].(string)

	if _, ok := event.data["reportingHosts"]; !ok {
		return
	}

	reportingHosts := event.data["reportingHosts"].(float64)

	for k, v := range event.data {
		// This are the only properties we want per command/pool.
		if !strings.HasPrefix(k, "rollingCount") && !strings.HasPrefix(k, "current") &&
			!strings.HasPrefix(k, "isCircuitBreakerOpen") && !strings.HasPrefix(k, "latencyExecute") &&
			!strings.HasPrefix(k, "latencyTotal") {
			continue
		}

		statKey := buildStatKey(event.clusterName, name, resourceType, k)

		switch v := v.(type) {
		default:
			log.Printf("unexpected data element %T, %s", v, v)
		case string:
			// ignored
		case map[string]interface{}:
			for pct, val := range v {
				client.Gauge(statKey+"."+strings.Replace(pct, ".", "_", -1)+"_pct", int64(val.(float64)/reportingHosts), 1.0)
			}
		case bool:
			if v {
				client.Gauge(statKey, 1, 1.0)
			} else {
				client.Gauge(statKey, 0, 1.0)
			}
		case int64:
			client.Gauge(statKey, v/int64(reportingHosts), 1.0)
		case float64:
			client.Gauge(statKey, int64(v/reportingHosts), 1.0)
		}
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
