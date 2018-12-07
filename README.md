# cron2

Cron2 is a job scheduling service similar to Cron

## Features

- Written in Go
- Configuration is defined with HCL
- Timezone support
- Custom logging per job
- Environment variables
- Docker support
- Test runs
- Notifications

## Motivation

Cron2 is created because i don't like standard Cron service that ships with most
Linux boxes.

## Configuration

Configuration file is based on HCL syntax. See example:

```hcl
// Simple job.
// Runs every minute.
job "test1" {
  spec = "* * * * *"
  command = "ping -c 1 google.com"
}

// Periodic job with time zone. Default timezone is UTC.
// Run at the beginning of hour, every 3 hours.
job "test2" {
  spec = "0 */3 * * *"
  tz = "America/Chicago"
  command = "rake db:backup"
}

// More job options
job "test3" {
  spec = "0 9 * * *"
  command = "rake reports:generate"
  user = "deploy"
  dir = "/home/deploy/app/current"

  // Add extra environment variables to the job
  env {
    RAILS_ENV = "production"
    DEBUG = "true"
  }
}
```

Full configuration example:

```hcl
job "demo" {
  spec = "0 * * * *"
  command = "curl https://google.com | jq"
  bash = true
  tz = "America/Chicago"
  user = "foo"
  dir = "/home/foo"
  timeout = "5s"
  log = "/tmp/log/demo.log"
  env {
    FOO = "bar"
  }
}
```

Docker configuration:

```hcl
// Test command that will print out directory structure
job "docker test" {
  spec = "* * * * *"
  command = "ls -al"
  docker {
    image = "alpine:3.6"
  }
}
```