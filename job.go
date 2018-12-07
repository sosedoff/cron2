package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Job represents the cron job
type Job struct {
	config     *JobConfig
	startedAt  time.Time
	duration   time.Duration
	success    bool
	exitStatus int
}

// Run executes the job
func (j Job) Run() {
	j.startedAt = time.Now()
	log.Printf("[%s] job started\n", j.config.Name)

	switch j.config.RunMode {
	case nativeMode:
		runNative(&j)
	case dockerMode:
		runDocker(&j)
	}

	j.duration = time.Since(j.startedAt)
	log.Printf(
		"[%s] job finished with %d. success: %v, duration: %v,\n",
		j.config.Name,
		j.exitStatus,
		j.success,
		j.duration,
	)

	sendNotifications(&j)
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

	j.success = true

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

// sendNotifications sends alerts to all notification targets
func sendNotifications(j *Job) {
	notify := j.config.Notify
	if notify == nil {
		return
	}
	if notify.Mode == notifyError && j.success == true {
		return
	}

	var message string
	if j.success {
		message = fmt.Sprintf("Job %q has finished. Duration: %v", j.config.Name, j.duration)
	} else {
		message = fmt.Sprintf("Job %q has failed with status code: %v. Duration: %v", j.config.Name, j.exitStatus, j.duration)
	}

	log.Printf("[%s] sending notifications\n", j.config.Name)
	defer log.Printf("[%s] done sending notifications\n", j.config.Name)

	wg := &sync.WaitGroup{}

	if webhook := notify.Webhook; webhook != nil {
		wg.Add(1)

		go func() {
			defer wg.Done()

			form := url.Values{}
			form.Add("job_name", j.config.Name)
			form.Add("duration", fmt.Sprintf("%v", j.duration))
			form.Add("started_at", fmt.Sprintf("%v", j.startedAt))
			form.Add("success", fmt.Sprintf("%v", j.success))
			form.Add("exit_status", fmt.Sprintf("%v", j.exitStatus))
			form.Add("message", message)

			resp, err := http.PostForm(webhook.URL, form)
			if err != nil {
				log.Printf("[%s] failed to send webhook: %v\n", j.config.Name, err)
				return
			}
			resp.Body.Close()

			log.Printf("[%s] sent notification to webhook %v\n", j.config.Name, webhook.URL)
		}()
	}

	if slack := notify.Slack; slack != nil {
		wg.Add(1)

		go func() {
			defer wg.Done()

			payload := map[string]string{
				"text":     message,
				"username": slack.User,
				"channel":  slack.Channel,
			}
			body, err := json.Marshal(payload)
			if err != nil {
				log.Printf("[%s] json error: %v\n", j.config.Name, err)
				return
			}

			resp, err := http.Post(slack.URL, "application/json", bytes.NewReader(body))
			if err != nil {
				log.Printf("[%s] failed to send slack: %v\n", j.config.Name, err)
				return
			}
			resp.Body.Close()

			log.Printf("[%s] sent notification to slack %v\n", j.config.Name, slack.Channel)
		}()
	}

	wg.Wait()
}
