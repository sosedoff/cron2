package main

import (
	"log"
	"os"

	"gopkg.in/robfig/cron.v2"
)

type Service struct {
	config    *Config
	scheduler *cron.Cron
}

func newService(config *Config) (*Service, error) {
	return &Service{
		config:    config,
		scheduler: cron.New(),
	}, nil
}

func (s *Service) addJobs() error {
	if len(s.config.Jobs) == 0 {
		log.Println("no jobs found")
		return nil
	}

	for _, config := range s.config.Jobs {
		log.Printf("adding job %q\n", config.Name)
		_, err := s.scheduler.AddJob(config.fullSpec(), Job{config: config})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) start() error {
	if err := s.addJobs(); err != nil {
		return err
	}

	log.Println("starting scheduler")
	defer log.Println("scheduler has stopped")

	for _, j := range s.config.Jobs {
		Job{config: j}.Run()
	}

	os.Exit(1)

	s.scheduler.Start()
	defer s.scheduler.Stop()

	// TODO: add channels to handle shutdown
	select {}

	return nil
}
