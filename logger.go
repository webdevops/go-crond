package main

import (
	"fmt"
	"log"
	"strings"
	"os"
	"time"
)

var (
	LoggerInfo CronLogger
	LoggerError CronLogger
)

func initLogger() {
	LoggerInfo = CronLogger{log.New(os.Stdout, LogPrefix, 0)}
	LoggerError = CronLogger{log.New(os.Stderr, LogPrefix, 0)}
}

type CronLogger struct {
	*log.Logger
}

func (CronLogger CronLogger) Verbose(message string) {
	if opts.Verbose {
		CronLogger.Println(message)
	}
}

func (CronLogger CronLogger) CronjobToString(cronjob CrontabEntry) string {
	parts := []string{}

	parts = append(parts, fmt.Sprintf("spec:'%v'", cronjob.Spec))
	parts = append(parts, fmt.Sprintf("usr:%v", cronjob.User))
	parts = append(parts, fmt.Sprintf("cmd:'%v'", cronjob.Command))

	if len(cronjob.Env) >= 1 {
		parts = append(parts, fmt.Sprintf("env:'%v'", cronjob.Env))
	}

	return strings.Join(parts, " ")
}

func (CronLogger CronLogger) CronjobAdd(cronjob CrontabEntry) {
	CronLogger.Printf("add: %v\n", CronLogger.CronjobToString(cronjob))
}

func (CronLogger CronLogger) CronjobExec(cronjob CrontabEntry) {
	if opts.Verbose {
		CronLogger.Printf("exec: %v\n", CronLogger.CronjobToString(cronjob))
	}
}

func (CronLogger CronLogger) CronjobExecFailed(cronjob CrontabEntry, output string, err error, elapsed time.Duration) {
	CronLogger.Printf("failed cronjob: cmd:%v out:%v err:%v time:%s\n", cronjob.Command, output, err, elapsed)
}

func (CronLogger CronLogger) CronjobExecSuccess(cronjob CrontabEntry, output string, err error, elapsed time.Duration) {
	if opts.Verbose {
		CronLogger.Printf("ok: cronjob: cmd:%v out:%v err:%v time:%s\n", cronjob.Command, output, err, elapsed)
	}
}
