package main

import (
	"log"
	"sync"

	"gopkg.in/robfig/cron.v2"
)

type Service struct {
	config     *Config
	configLock *sync.Mutex
	scheduler  *cron.Cron
}

func newService(config *Config) (*Service, error) {
	return &Service{
		config:     config,
		configLock: new(sync.Mutex),
		scheduler:  cron.New(),
	}, nil
}

func (s *Service) addJobs() error {
	s.configLock.Lock()
	defer s.configLock.Unlock()

	for _, e := range s.scheduler.Entries() {
		s.scheduler.Remove(e.ID)
	}

	if len(s.config.Jobs) == 0 {
		log.Println("no jobs found")
		return nil
	}

	for _, config := range s.config.Jobs {
		if config.Disabled {
			log.Printf("job %q is disabled, skipping\n", config.Name)
			continue
		}

		log.Printf("adding job %q\n", config.Name)
		entry, err := s.scheduler.AddJob(config.fullSpec(), Job{config: config})
		if err != nil {
			return err
		}
		config.ID = int(entry)
	}

	return nil
}

func (s *Service) reload(config *Config) error {
	s.configLock.Lock()
	s.config = config
	s.configLock.Unlock()
	return s.addJobs()
}

func (s *Service) start() error {
	if err := s.addJobs(); err != nil {
		return err
	}

	log.Println("starting scheduler")
	defer log.Println("scheduler has stopped")

	s.scheduler.Start()
	defer s.scheduler.Stop()

	// TODO: add channels to handle shutdown
	select {}

	return nil
}
