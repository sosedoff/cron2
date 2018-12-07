package main

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
)

// configKeys is a list of allowed keys in the config
var configKeys = []string{
	"job",
}

// jobKeys lists all allowed keys inside "job" block
var jobKeys = []string{
	"name",
	"disabled",
	"spec",
	"command",
	"shell",
	"env",
	"tz",
	"dir",
	"user",
	"log",
	"docker",
	"timeout",
	"notify",
}

// Config represents a service configuration
type Config struct {
	Jobs []*JobConfig `hcl:"job"`
}

// readConfig reads and returns a new configuration
func readConfig(path string) (*Config, error) {
	// Read the contents of the config file
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Parse text into HCL
	root, err := hcl.ParseBytes(data)
	if err != nil {
		return nil, err
	}

	// Get the top level elements
	list, ok := root.Node.(*ast.ObjectList)
	if !ok {
		return nil, errors.New("file does not containe a root object")
	}

	// Validate top level keys
	if err := checkHCLKeys(list, configKeys); err != nil {
		return nil, err
	}

	config := &Config{}
	jobNames := map[string]bool{}

	// Load all job definitions
	if o := list.Filter("job"); len(o.Items) > 0 {
		for _, item := range o.Items {
			node := item.Val

			// Validate all keys in the "job" block
			if err := checkHCLKeys(node, jobKeys); err != nil {
				return nil, err
			}

			// Parse the job block into config
			job := new(JobConfig)
			if err := hcl.DecodeObject(job, node); err != nil {
				return nil, err
			}

			// Try to find the job name from the block definition
			if job.Name == "" && len(item.Keys) > 0 {
				job.Name = item.Keys[0].Token.Value().(string)
			}

			// Validate the job config
			if err := job.validate(); err != nil {
				return nil, fmt.Errorf("error at %s: %s", node.Pos().String(), err.Error())
			}

			// Check for job duplicates
			if jobNames[job.Name] {
				return nil, fmt.Errorf("duplicate job: %s", job.Name)
			}
			jobNames[job.Name] = true

			config.Jobs = append(config.Jobs, job)
		}
	}

	return config, nil
}

// findJob returns a job config that matches given name
func (c *Config) findJob(name string) *JobConfig {
	for _, j := range c.Jobs {
		if j.Name == name {
			return j
		}
	}
	return nil
}
