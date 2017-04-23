# go-crond

[![GitHub release](https://img.shields.io/github/release/webdevops/go-crond.svg)](https://github.com/webdevops/go-crond/releases)
[![license](https://img.shields.io/github/license/webdevops/go-crond.svg)](https://github.com/webdevops/go-crond/blob/master/LICENSE)
[![Build Status](https://travis-ci.org/webdevops/go-crond.svg?branch=master)](https://travis-ci.org/webdevops/go-crond)
[![Github All Releases](https://img.shields.io/github/downloads/webdevops/go-crond/total.svg)]()
[![Github Releases](https://img.shields.io/github/downloads/webdevops/go-crond/latest/total.svg)]()

Cron daemon implemented in Golang

Inspired by https://github.com/anarcher/go-cron

## Usage

```
Usage:
  go-crond

Application Options:
      --processes=         Number of parallel executions (default: 1)
      --default-user=      Default user (default: root)
      --include=           Include files in directory as system crontabs (with user)
      --run-parts=         Include files in directory with dynamic time execution (time-spec:path)
      --run-parts-1min=    Include files in directory every minute execution (run-part)
      --run-parts-hourly=  Include files in directory every hour execution (run-part)
      --run-parts-daily=   Include files in directory every day execution (run-part)
      --run-parts-weekly=  Include files in directory every week execution (run-part)
      --run-parts-monthly= Include files in directory every month execution (run-part)
  -V, --version            show version and exit
  -h, --help               show this help message
```

Crontab files can be added as arguments or automatic included by using eg. `--include-crond=path/`

### Examples

Run crond with 3 crontab files:

    go-crond crontab1 crontab2 crontab3


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
