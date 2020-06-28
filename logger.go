package main

import (
	log "github.com/sirupsen/logrus"
)

func LogCronjobToFields(cronjob CrontabEntry) log.Fields {
	return log.Fields{
		"spec": cronjob.Spec,
		"user": cronjob.User,
		"command": cronjob.Command,
	}
}
