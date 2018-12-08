package main

import (
	"flag"
	"log"
)

var (
	configPath   string
	socketPath   string
	validateOnly bool
	triggerName  string
	listJobs     bool
)

func main() {
	flag.StringVar(&configPath, "config", "/etc/cron2", "Path to config file")
	flag.StringVar(&socketPath, "socket", "/var/run/cron2.sock", "Path to unix socket")
	flag.BoolVar(&validateOnly, "validate", false, "Validate config syntax")
	flag.StringVar(&triggerName, "trigger", "", "Trigger a job")
	flag.BoolVar(&listJobs, "list", false, "Show all jobs")
	flag.Parse()

	// Trigger a job from the command line
	if triggerName != "" {
		if err := triggerJob(socketPath, triggerName); err != nil {
			log.Fatal(err)
		}
		return
	}

	// Show all configured jobs
	if listJobs {
		if err := listCurrentJobs(socketPath); err != nil {
			log.Fatal(err)
		}
		return
	}

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

	go startFilewatcher(service, configPath)
	go startListener(service, socketPath)

	if err := service.start(); err != nil {
		log.Fatal(err)
	}
}
