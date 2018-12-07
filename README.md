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
Linux boxes. Also, to mess around with HCL.

## Install

```
go get -u github.com/sosedoff/cron2
```

## Configuration

Basic example:

```hcl
// Simple job that runs every minute.
job "simple" {
  spec = "* * * * *"
  command = "ping -c 1 google.com"
}
```

Customize timezone:

```hcl
// Run command every day at 9am CST.
// The default time zone is UTC.
job "demp" {
  spec = "0 9 * * *"
  command = "rake db:backup"
  tz = "America/Chicago"
}
```

More configuration options:

```hcl
job "demo" {
  spec = "0 9 * * *"
  command = "rake reports:generate"

  // Specify user for the job
  user = "deploy"

  // Change directory
  dir = "/home/deploy/app/current"

  // Custom log location
  log = "/var/log/myjob.log"

  // Configure environment variables
  env {
    RAILS_ENV = "production"
    DEBUG = "true"
  }

  // Configure max execution time
  timeout = "30min"

  // Setup notifications
  notify {
    // Configure delivery mode
    // Change to "all" to receive notifications for all runs
    on = "error"

    webhook {
      // Will send POST to this URL
      url = "https://mywebhook.com"
    }

    slack {
      // Set to incoming webhook URL
      url = "https://hooks.slack.com/services/..."

      // Set channel (optional)
      channel = "#ops"
      
      // Change username (optional)
      username = "cronbot"
    }
  }
}
```

Run with shell:

```hcl
job "test" {
  // Commands are not run in shell by default
  shell = "bash"

  spec = "* * * * *"
  command = "cat /my/file | grep foobar | xargs gzip --best"
}
```

Inline scripts:

```hcl
job "inline" {
  spec = "* * * * *"

  // You can write multi line bash scripts like this:
  command = <<END
    # Dump your bash script in here
    ls -al
    echo "line1"
    echo "line2"
  END
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

### Testing jobs

Let's look at the example config: the job is going to be executed at 9am every day.
We want to test the job without having to change the spec.

```hcl
job "test" {
  spec = "0 9 * * *"
  tz = "America/Chicago"
  command = "do something"
}

Start the service first:

```
cron2 -config=/config.hcl -socket=/var/run/cron2.sock
```

Once the service is running, trigger the job:

```
cron2 -socket=/var/run/cron2.sock -trigger=test
```

If the job exists you'll see success message:

```
ok: scheduled
```