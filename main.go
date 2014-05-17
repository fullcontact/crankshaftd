package main

import (
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/xorlev/crankshaftd/crankshaft"
	"os"
)

var (
	configFile = flag.String("config", "config.toml", "Configuration file")
	config     crankshaft.Config
)

func usage() {
	fmt.Fprintln(os.Stderr, "usage: crankshaft [opts]")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		fmt.Println(err)
		return
	}

	if len(config.Host) == 0 {
		fmt.Fprintln(os.Stderr, "Error: Must specify a turbine host")
		usage()
	}

	if len(config.Clusters) == 0 {
		fmt.Fprintln(os.Stderr, "Error: Must specify at least one cluster")
		usage()
	}

	crankshaft.MonitorClusters(config)
}
