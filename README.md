# go-crond

[![GitHub release](https://img.shields.io/github/release/webdevops/go-crond.svg)](https://github.com/webdevops/go-crond/releases)
[![license](https://img.shields.io/github/license/webdevops/go-crond.svg)](https://github.com/webdevops/go-crond/blob/master/LICENSE)
[![Build Status](https://travis-ci.org/webdevops/go-crond.svg?branch=master)](https://travis-ci.org/webdevops/go-crond)
[![Github All Releases](https://img.shields.io/github/downloads/webdevops/go-crond/total.svg)]()
[![Github Releases](https://img.shields.io/github/downloads/webdevops/go-crond/latest/total.svg)]()

Cron daemon implemented in Golang

Inspired by https://github.com/anarcher/go-cron

Using https://godoc.org/github.com/robfig/cron


## Features

- system crontab (with username inside)
- user crontabs (without user inside)
- run-parts support
- Logging to STDOUT and STDERR (instead of sending mails)
- Keep current environment (eg. for usage in Docker containers)

## Usage

```
Usage:
  go-crond

Application Options:
      --processes=         Number of parallel executions (default: 1)
      --default-user=      Default user (default: root)
      --include=           Include files in directory as system crontabs (with user)
      --system-defaults    Include standard paths for distribution
      --run-parts=         Execute files in directory with custom spec (like run-parts; spec-units:ns,us,s,m,h; format:time-spec:path; eg:10s,1m,1h30m)
      --run-parts-1min=    Execute files in directory every beginning minute (like run-parts)
      --run-parts-15min=   Execute files in directory every beginning 15 minutes (like run-parts)
      --run-parts-hourly=  Execute files in directory every beginning hour (like run-parts)
      --run-parts-daily=   Execute files in directory every beginning day (like run-parts)
      --run-parts-weekly=  Execute files in directory every beginning week (like run-parts)
      --run-parts-monthly= Execute files in directory every beginning month (like run-parts)
  -v, --verbose            verbose mode
  -V, --version            show version and exit
  -h, --help               show this help message
```

Crontab files can be added as arguments or automatic included by using eg. `--include-crond=path/`

### Examples

Run crond with a system crontab:

    go-crond examples/crontab


Run crond with user crontabs (without user in it) under specific users:

    go-crond root:examples/crontab-root guest:examples/crontab-guest


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

## Installation

```bash
GOCROND_VERSION=0.1.0 \
&& wget -O /usr/local/bin/go-crond https://github.com/webdevops/go-crond/releases/download/$GOREPLACE_VERSION/go-crond-64-linux \
&& chmod +x /usr/local/bin/go-crond
```
