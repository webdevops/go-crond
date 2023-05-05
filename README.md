# go-crond

[![GitHub release](https://img.shields.io/github/release/webdevops/go-crond.svg)](https://github.com/webdevops/go-crond/releases)
[![license](https://img.shields.io/github/license/webdevops/go-crond.svg)](https://github.com/webdevops/go-crond/blob/master/LICENSE)
[![DockerHub](https://img.shields.io/badge/DockerHub-webdevops%2Fgo--crond-blue)](https://hub.docker.com/r/webdevops/go-crond/)
[![Quay.io](https://img.shields.io/badge/Quay.io-webdevops%2Fgo--crond-blue)](https://quay.io/repository/webdevops/go-crond)
[![Github All Releases](https://img.shields.io/github/downloads/webdevops/go-crond/total.svg)]()
[![Github Releases](https://img.shields.io/github/downloads/webdevops/go-crond/latest/total.svg)]()

A cron daemon written in golang

Inspired by https://github.com/anarcher/go-cron

Using https://godoc.org/github.com/robfig/cron

## Docker images

on [Docker hub](https://hub.docker.com/repository/docker/webdevops/go-crond/tags)

- `webdevops/go-crond:alpine` (based on `alpine`)
- `webdevops/go-crond:ubuntu` (based on `ubuntu:latest`)
- `webdevops/go-crond:debian` (based on `debian:stable-slim`)
- `webdevops/go-crond:{version}-alpine` (based on `alpine`)
- `webdevops/go-crond:{version}-ubuntu` (based on `ubuntu:latest`)
- `webdevops/go-crond:{version}-debian` (based on `debian:stable-slim`)

## Features

- system crontab (with username inside)
- user crontabs (without username inside)
- run-parts support
- Logging to STDOUT and STDERR (instead of sending mails)
- Keep current environment (eg. for usage in Docker containers)
- Supports Linux, MacOS, ARM/ARM64 (Rasbperry Pi and others)

## Usage

```
Usage:
  go-crond [OPTIONS] [Crontabs...]

Application Options:
  -V, --version               show version and exit
      --dumpversion           show only version number and exit
  -h, --help                  show this help message
      --default-user=         Default user (default: root)
      --include=              Include files in directory as system crontabs (with user)
      --auto                  Enable automatic system crontab detection
      --run-parts=            Execute files in directory with custom spec (like run-parts; spec-units:ns,us,s,m,h;
                              format:time-spec:path; eg:10s,1m,1h30m)
      --run-parts-1min=       Execute files in directory every beginning minute (like run-parts)
      --run-parts-15min=      Execute files in directory every beginning 15 minutes (like run-parts)
      --run-parts-hourly=     Execute files in directory every beginning hour (like run-parts)
      --run-parts-daily=      Execute files in directory every beginning day (like run-parts)
      --run-parts-weekly=     Execute files in directory every beginning week (like run-parts)
      --run-parts-monthly=    Execute files in directory every beginning month (like run-parts)
      --allow-unprivileged    Allow daemon to run as non root (unprivileged) user
      --working-directory=    Set the working directory for crontab commands (default: /)
  -v, --verbose               verbose mode [$VERBOSE]
      --log.json              Switch log output to json format [$LOG_JSON]
      --server.bind=          Server address, eg. ':8080' (/healthz and /metrics for prometheus) [$SERVER_BIND]
      --server.timeout.read=  Server read timeout (default: 5s) [$SERVER_TIMEOUT_READ]
      --server.timeout.write= Server write timeout (default: 10s) [$SERVER_TIMEOUT_WRITE]
      --server.metrics        Enable prometheus metrics (do not use senstive informations in commands -> use environment
                              variables or files for storing these informations) [$SERVER_METRICS]

Help Options:
  -h, --help                  Show this help message

Arguments:
  Crontabs:                   path to crontab files
```

Crontab files can be added as arguments or automatic included by using eg. `--include=crond-path/`

### Examples

Run crond with a system crontab:

    go-crond examples/crontab


Run crond with user crontabs (without user in it) under specific users:

    go-crond \
        root:examples/crontab-root \
        guest:examples/crontab-guest


Run crond with auto include of /etc/cron.d and script execution of hourly, weekly, daily and monthly:

    go-crond \
        --include=/etc/cron.d \
        --run-parts-hourly=/etc/cron.hourly \
        --run-parts-weekly=/etc/cron.weekly \
        --run-parts-daily=/etc/cron.daily \
        --run-parts-monthly=/etc/cron.monthly

Run crond with run-parts with custom time spec:

    go-crond \
        --run-parts=1m:/etc/cron.minute \
        --run-parts=15m:/etc/cron.15min

Run crond with run-parts with custom time spec and different user:

    go-crond \
        --run-parts=1m:application:/etc/cron.minute \
        --run-parts=15m:admin:/etc/cron.15min

## Installation

```bash
GOCROND_VERSION=22.9.1 \
GOCRON_OS=linux \
GOCRON_ARCH=amd64 \
&& wget -O /usr/local/bin/go-crond https://github.com/webdevops/go-crond/releases/download/${GOCROND_VERSION}/go-crond.${GOCRON_OS}.${GOCRON_ARCH} \
&& chmod +x /usr/local/bin/go-crond
```

## Docker images

| Image                       | Description                                    |
|:----------------------------|:-----------------------------------------------|
| `webdevops/go-crond:latest` | Latest release, binary only                    |
| `webdevops/go-crond:master` | Current development version in branch `master` |

## Metrics

go-crond exposes [Prometheus][] metrics on `:8080/metrics` if enabled.


| Metric                      | Description                                     |
|:----------------------------|:------------------------------------------------|
| `gocrond_task_info`         | List of all cronjobs                            |
| `gocrond_task_run_count`    | Counter for each executed task                  |
| `gocrond_task_run_result`   | Last status (0=failed, 1=success) for each task |
| `gocrond_task_run_time`     | Last exec time (unix timestamp) for each task   |
| `gocrond_task_run_duration` | Duration of last exec                           |

[Prometheus]: https://prometheus.io/

##jsonlog in ElasticSearch
##Logstash

An example of our production filter for logstash
```
logstashPipeline:
  logstash.conf: |
    input {
      beats {
        port => "5044"
      }
    }
    filter {
      json {
        skip_on_invalid_json => true
        source => "message"
        target => "json-data"
        add_tag => [ "_message_json_parsed" ]
      }
      if [kubernetes][container][name] =~ ".*-cron" {
        mutate {
          rename => { "[json-data][level]" => "[json-data][level_name]" }
          gsub => [ "[json-data][msg]", "\\\\n", "\n" ]
          remove_field => [ "message" ]
        }
        date {
          match => [ "[json-data][started_at]", "ISO8601" ]
        }
      }
      ...
```
##Grafana
![Image alt](https://github.com/promzeus/go-crond/master/grafana-dashboard/go-crond.png)