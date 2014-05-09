package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/cactus/go-statsd-client/statsd"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	host       = flag.String("host", "", "Turbine Host")
	port       = flag.Int("port", 80, "Turbine Port")
	clusters   = flag.String("clusters", "", "Comma-separated clusters to monitor")
	maxretries = flag.Int("maxretries", 5, "Max number of times to retry on errors")
	statsHost  = flag.String("host", "", "StatsD Host")
	statsPort  = flag.Int("host", 8125, "StatsD Port")
)

func usage() {
	fmt.Fprintln(os.Stderr, "usage: crankshaft [opts]")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if len(*host) == 0 {
		fmt.Fprintln(os.Stderr, "Error: Must specify a turbine host")
		usage()
	}

	if len(*statsHost) == 0 {
		fmt.Fprintln(os.Stderr, "Error: Must specify a StatsD host")
		usage()
	}

	if *port == 443 {
		fmt.Fprintln(os.Stderr, "Error: Crankshaft currently does not handle TLS connections")
		usage()
	}

	if len(*clusters) == 0 {
		fmt.Fprintln(os.Stderr, "Error: Must specify at least one cluster")
		usage()
	}

	clusterList := strings.Split(*clusters, ",")
	channels := make([]chan error, len(clusterList))

	for i := range clusterList {
		go turbine(channels[i], strings.TrimSpace(clusterList[i]))

	}

	for i := range channels {
		<-channels[i]
	}
}

func turbine(c chan error, clusterName string) {
	defer close(c)

	for {
		err := attachToTurbine(clusterName)
		if err != nil {
			log.Println("Turbine session ended with error", err, "restarting...")
			time.Sleep(3 * time.Second) // wait
		}
	}

	c <- nil
}

func attachToTurbine(clusterName string) error {
	client, err := statsd.New(*statsHost+":"+strconv.Itoa(*statsPort), "hystrix")
	if err != nil {
		log.Println("Error creating StatsD client")
		return err
	}

	log.Println("Opening Turbine connection for", clusterName)

	// TODO: urlencode
	req, err := http.NewRequest("GET", "/turbine.stream?cluster="+clusterName, nil)

	if err != nil {
		log.Println("Error creating HTTP request", err)
		return err
	}

	conn, err := openConnection()
	defer conn.Close()
	if err != nil {
		log.Println("Error opening TCP socket", err)
		return err
	}

	clientConn := httputil.NewClientConn(conn, nil)
	resp, err := clientConn.Do(req)

	if err != nil {
		log.Println("Error sending HTTP request", err)
		return err
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')

		if err != nil {
			log.Println("Error reading bytes from stream", err)
			return err
		}

		line = bytes.TrimSpace(line)
		s := string(line[:])

		if strings.HasPrefix(s, "data: ") {
			s = strings.TrimPrefix(s, "data: ")

			data, err := unmarshalJson(s)

			if err != nil {
				log.Println("Error decoding JSON", err)
				return err
			}

			writeStats(clusterName, data, client)
		}
	}
}

func writeStats(clusterName string, data map[string]interface{}, client statsd.Statter) {
	name := data["name"].(string)
	resourceType := data["type"].(string)

	for k, v := range data {
		// This are the only properties we want per command/pool.
		if !strings.HasPrefix(k, "rollingCount") && !strings.HasPrefix(k, "current") &&
			!strings.HasPrefix(k, "isCircuitBreakerOpen") && !strings.HasPrefix(k, "latencyTotal") {
			continue
		}

		statKey := buildStatKey(clusterName, name, resourceType, k)

		switch v := v.(type) {
		default:
			log.Printf("unexpected type %T, %s", v, v) // %T prints whatever type t has
		case string:
			// ignored
		case map[string]interface{}:
			for pct, val := range v {
				client.Gauge(statKey+"."+strings.Replace(pct, ".", "_", -1)+"_pct", int64(val.(float64)), 1.0)
			}
		case bool:
			if v {
				client.Gauge(statKey, 1, 1.0)
			} else {
				client.Gauge(statKey, 0, 1.0)
			}
		case int64:
			client.Gauge(statKey, v, 1.0)
		case float64:
			client.Gauge(statKey, int64(v), 1.0)
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

func unmarshalJson(payload string) (map[string]interface{}, error) {
	var f map[string]interface{}
	err := json.Unmarshal([]byte(payload), &f)

	return f, err
}

func openConnection() (net.Conn, error) {
	conn, err := net.Dial("tcp", *host+":"+strconv.Itoa(*port))

	return conn, err
}
