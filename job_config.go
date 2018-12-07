package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/sosedoff/cron"
)

const (
	// Run modes
	nativeMode = "native"
	dockerMode = "docker"

	// Notification modes
	notifyError = "error"
	notifyAll   = "all"
)

// JobConfig represents a single job in the configuration file
type JobConfig struct {
	ID            string            `hcl:"-"`       // Internal entry ID
	Name          string            `hcl:"name"`    // Command name
	Spec          string            `hcl:"spec"`    // Cron expression
	Timezone      string            `hcl:"tz"`      // Time zone
	Command       string            `hcl:"command"` // Run command
	User          string            `hcl:"user"`    // Run as user
	Dir           string            `hcl:"dir"`     // Working dir
	Environment   map[string]string `hcl:"env"`     // Env vars
	Log           string            `hcl:"log"`     // Path to log file
	BashMode      bool              `hcl:"bash"`    // Run in bash wrapper
	TimeoutString string            `hcl:"timeout"` // Max execution time
	Docker        *DockerConfig     `hcl:"docker"`  // Docker options
	Notify        *NotifyConfig     `hcl:"notify"`  // Notification options

	// Computed fields
	RunMode string        `hcl:"-"`
	Timeout time.Duration `hcl:"-"`
}

// NotifyConfig represents job notification settings
type NotifyConfig struct {
	Mode string `hcl:"send"` // Mode could be one of "errors", "all"

	Webhook *struct {
		URL string `hcl:"url"`
	} `hcl:"webhook"`

	Slack *struct {
		URL     string `hcl:"url"`
		User    string `hcl:"username"`
		Channel string `hcl:"channel"`
	} `hcl:"slack"`
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

	// Notify on errors only by default
	if j.Notify != nil && j.Notify.Mode == "" {
		j.Notify.Mode = notifyError
	}

	return nil
}
