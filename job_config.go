package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/sosedoff/cron"
)

var (
	nativeMode = "native"
	dockerMode = "docker"
)

// JobConfig represents a single job in the configuration file
type JobConfig struct {
	ID            string            `hcl:"-"`
	Name          string            `hcl:"name"`
	Spec          string            `hcl:"spec"`
	Timezone      string            `hcl:"tz"`
	Command       string            `hcl:"command"`
	User          string            `hcl:"user"`
	Dir           string            `hcl:"dir"`
	Environment   map[string]string `hcl:"env"`
	Log           string            `hcl:"log"`
	BashMode      bool              `hcl:"bash"`
	TimeoutString string            `hcl:"timeout"`
	Docker        *DockerConfig     `hcl:"docker"`

	// Computed fields
	RunMode string        `hcl:"-"`
	Timeout time.Duration `hcl:"-"`
}

// DockerConfig represends config options for docker run
type DockerConfig struct {
	Image string `hcl:"image"`
}

// fullSpec returns cron spec with time zone
func (j *JobConfig) fullSpec() string {
	if j.Timezone != "" {
		return fmt.Sprintf("TZ=%s %s", j.Timezone, j.Spec)
	}
	return j.Spec
}

// validate performs validation on job attributes
func (j *JobConfig) validate() error {
	if j.Name == "" {
		return errors.New("job must have a name")
	}

	if j.Command == "" {
		return errors.New("command is required")
	}

	if j.Spec == "" {
		return errors.New("spec is required")
	}

	if j.Docker != nil {
		j.RunMode = dockerMode
	} else {
		j.RunMode = nativeMode
	}

	if _, err := cron.Parse(j.Spec); err != nil {
		return fmt.Errorf("invalid cron spec: %v", err)
	}

	if val := j.TimeoutString; val != "" {
		dur, err := time.ParseDuration(val)
		if err != nil {
			return fmt.Errorf("invalid timeout: %v", err)
		}
		j.Timeout = dur
	}

	return nil
}
