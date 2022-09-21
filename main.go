package main

import (
	"flag"
	"log"
	"workserver/config"
	"workserver/serverd"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

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


