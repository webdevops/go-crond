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
      --processes=       Number of parallel executions (default: 1)
      --default-user=    Default user (default: root)
      --include-crond=   Include files in directory as system crontabs (with user)
      --include-15min=   Include files in directory for 15 min execution
      --include-hourly=  Include files in directory for hourly execution
      --include-daily=   Include files in directory for daily execution
      --include-weekly=  Include files in directory for weekly execution
      --include-monthly= Include files in directory for monthly execution
  -V, --version          show version and exit
  -h, --help             show this help message
```

Crontab files can be added as arguments or automatic included by using eg. `--include-crond=path/`
