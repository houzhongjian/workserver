package main

import (
	"flag"
	"workserver/config"
	"workserver/serverd"
)

func main() {
	parseFlag()

	config.Load(configDir)

	runServer()
}

func runServer() {
	serverd.Run()
}

var configDir string

func parseFlag() {
	flag.StringVar(&configDir, "conf", "./conf", "config path")
	flag.Parse()
}


