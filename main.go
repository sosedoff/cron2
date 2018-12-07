package main

import (
	"flag"
	"log"
)

var (
	configPath   string
	validateOnly bool
)

func main() {
	flag.StringVar(&configPath, "config", "", "Path to config file")
	flag.BoolVar(&validateOnly, "validate", false, "Validate config syntax")
	flag.Parse()

	if configPath == "" {
		log.Fatal("Config path is not set")
	}

	config, err := readConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	// Exit after config is validated
	if validateOnly {
		return
	}

	service, err := newService(config)
	if err != nil {
		log.Fatal(err)
	}

	go startFilewatcher(configPath)

	if err := service.start(); err != nil {
		log.Fatal(err)
	}
}
