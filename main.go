package main

import (
	"flag"
	"log"
)

var (
	configPath string
	testConfig bool
)

func main() {
	flag.StringVar(&configPath, "c", "", "Path to config file")
	flag.BoolVar(&testConfig, "t", false, "Test config synax")
	flag.Parse()

	if configPath == "" {
		log.Fatal("Config path is not set")
	}

	config, err := readConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	// Exit after config is validated
	if testConfig {
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
