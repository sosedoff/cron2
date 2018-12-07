package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Job represents the cron job
type Job struct {
	config     *JobConfig
	startedAt  time.Time
	success    bool
	exitStatus int
}

// Run executes the job
func (j Job) Run() {
	j.startedAt = time.Now()

	log.Printf("[%s] job started\n", j.config.Name)
	defer func() {
		log.Printf(
			"[%s] job finished with %d. success: %v, duration: %v,\n",
			j.config.Name,
			j.exitStatus,
			j.success,
			time.Since(j.startedAt),
		)
	}()

	switch j.config.RunMode {
	case nativeMode:
		runNative(&j)
	case dockerMode:
		runDocker(&j)
	}

}

// runNative executes the job on the host system
func runNative(j *Job) {
	var ctx context.Context
	var cancelFunc context.CancelFunc

	ctx = context.Background()
	if j.config.Timeout.Seconds() > 0 {
		ctx, cancelFunc = context.WithTimeout(ctx, j.config.Timeout)
		defer cancelFunc()
	}

	var cmd *exec.Cmd
	if j.config.BashMode {
		// When in bash mode we can write commands like this:
		// curl http://example.com | jq | foobar
		//
		cmd = exec.CommandContext(ctx, "bash")
		cmd.Stdin = strings.NewReader(strings.TrimSpace(j.config.Command) + "\n")
	} else {
		chunks := strings.Split(j.config.Command, " ")
		cmd = exec.CommandContext(ctx, chunks[0], chunks[1:]...)
	}

	// Run is custom directory
	if j.config.Dir != "" {
		cmd.Dir = j.config.Dir
	}

	// Add custom environment
	if len(j.config.Environment) > 0 {
		for k, v := range j.config.Environment {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// Run as a different user
	if j.config.User != "" {
		usr, err := user.Lookup(j.config.User)
		if err != nil {
			log.Printf("[%s] cant find user %q: %v\n", j.config.User, err)
			j.exitStatus = 1
			j.success = false
			return
		}
		uid, _ := strconv.Atoi(usr.Uid)
		gid, _ := strconv.Atoi(usr.Gid)

		cmd.SysProcAttr = &syscall.SysProcAttr{}
		cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}
	}

	// Default output
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Prepare log file for the run
	if j.config.Log != "" {
		f, err := os.OpenFile(j.config.Log, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Printf("[%s] cant open log file:", err)
		}
		logFile := bufio.NewWriter(f)

		cmd.Stdout = logFile
		cmd.Stderr = logFile

		defer func() {
			logFile.Flush()
			f.Close()
		}()
	}

	if err := cmd.Run(); err != nil {
		j.success = false
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				j.exitStatus = status.ExitStatus()
			}
		}
		log.Printf("[%s] execution error: %v\n", j.config.Name, err)
		return
	}
}

// runDocker executes the job in a docker container
func runDocker(j *Job) {
	args := []string{
		"docker",
		"run",
		"-i",   // interfactive mode
		"--rm", // remove container after execution
	}

	// Working directory inside container
	if j.config.Dir != "" {
		args = append(args, "--workir", j.config.Dir)
	}

	// Append env vars
	for k, v := range j.config.Environment {
		args = append(args, "-e", k+"="+v)
	}

	// Run command also needs splitting
	args = append(args, j.config.Docker.Image)
	args = append(args, strings.Split(j.config.Command, " ")...)

	log.Println("command:", strings.Join(args, " "))

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Printf("[%s] execution error: %v\n", j.config.Name, err)
		return
	}
}
