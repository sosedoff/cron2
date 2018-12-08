package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sosedoff/cron"
)

const (
	// Default shell
	defaultShell = "bash"

	// Run modes
	nativeMode = "native"
	dockerMode = "docker"

	// Notification modes
	notifyError = "error"
	notifyAll   = "all"
)

// JobConfig represents a single job in the configuration file
type JobConfig struct {
	ID            int               `hcl:"-"`        // Internal entry ID
	Disabled      bool              `hcl:"disabled"` // Availability flag
	Name          string            `hcl:"name"`     // Command name
	Spec          string            `hcl:"spec"`     // Cron expression
	Timezone      string            `hcl:"tz"`       // Time zone
	Command       string            `hcl:"command"`  // Run command
	User          string            `hcl:"user"`     // Run as user
	Dir           string            `hcl:"dir"`      // Working dir
	Environment   map[string]string `hcl:"env"`      // Env vars
	Log           string            `hcl:"log"`      // Path to log file
	Shell         string            `hcl:"shell"`    // Shell to use for the run
	TimeoutString string            `hcl:"timeout"`  // Max execution time
	Docker        *DockerConfig     `hcl:"docker"`   // Docker options
	Notify        *NotifyConfig     `hcl:"notify"`   // Notification options

	// Computed fields
	RunMode string        `hcl:"-"`
	Timeout time.Duration `hcl:"-"`
}

// NotifyConfig represents job notification settings
type NotifyConfig struct {
	Mode string `hcl:"on"` // Mode could be one of "errors", "all"

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

// state returns current job state
func (j *JobConfig) state() string {
	if j.Disabled {
		return "inactive"
	}
	return "active"
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

	if j.Spec == "" {
		return errors.New("spec is required")
	}

	if j.Command == "" {
		return errors.New("command is required")
	}

	// Configure shell when multi-line scripts
	if j.Shell == "" && len(strings.Split(j.Command, "\n")) > 1 {
		j.Shell = defaultShell
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

	if j.Docker != nil {
		j.RunMode = dockerMode
	} else {
		j.RunMode = nativeMode
	}

	// Notify on errors only by default
	if j.Notify != nil && j.Notify.Mode == "" {
		j.Notify.Mode = notifyError
	}

	return nil
}
