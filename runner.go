package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Runner struct {
	cron *cron.Cron

	prometheus struct {
		taskRunSuccess *prometheus.GaugeVec
		taskRunTime *prometheus.GaugeVec
		taskRunDuration *prometheus.GaugeVec
	}
}

func NewRunner() *Runner {
	r := &Runner{
		cron: cron.New(),
	}

	r.prometheus.taskRunSuccess = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gocrond_task_run_success",
			Help: "gocrond task run successfull",
		},
		[]string{"cronSpec","cronUser","cronCommand"},
	)
	prometheus.MustRegister(r.prometheus.taskRunSuccess)

	r.prometheus.taskRunTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gocrond_task_run_time",
			Help: "gocrond task last run time",
		},
		[]string{"cronSpec","cronUser","cronCommand"},
	)
	prometheus.MustRegister(r.prometheus.taskRunTime)

	r.prometheus.taskRunDuration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gocrond_task_run_duration",
			Help: "gocrond task last run duration",
		},
		[]string{"cronSpec","cronUser","cronCommand"},
	)
	prometheus.MustRegister(r.prometheus.taskRunDuration)

	return r
}

// Add crontab entry
func (r *Runner) Add(cronjob CrontabEntry) error {
	cronSpec := cronjob.Spec
	if !strings.HasPrefix(cronjob.Spec, "@") {
		cronSpec = fmt.Sprintf("0 %s", cronjob.Spec)
	}

	err := r.cron.AddFunc(cronSpec, r.cmdFunc(cronjob, func(execCmd *exec.Cmd) bool {
		// before exec callback
		LoggerInfo.CronjobExec(cronjob)
		return true
	}))

	if err != nil {
		LoggerError.Printf("Failed add cron job spec:%v cmd:%v err:%v", cronjob.Spec, cronjob.Command, err)
	} else {
		LoggerInfo.CronjobAdd(cronjob)
	}

	return err
}

// Add crontab entry with user
func (r *Runner) AddWithUser(cronjob CrontabEntry) error {

	cronSpec := cronjob.Spec
	if !strings.HasPrefix(cronjob.Spec, "@") {
		cronSpec = fmt.Sprintf("0 %s", cronjob.Spec)
	}

	err := r.cron.AddFunc(cronSpec, r.cmdFunc(cronjob, func(execCmd *exec.Cmd) bool {
		// before exec callback
		LoggerInfo.CronjobExec(cronjob)

		// lookup username
		u, err := user.Lookup(cronjob.User)
		if err != nil {
			LoggerError.Printf("user lookup failed: %v", err)
			return false
		}

		// convert userid to int
		userId, err := strconv.ParseUint(u.Uid, 10, 32)
		if err != nil {
			LoggerError.Printf("Cannot convert user to id:%v", err)
			return false
		}

		// convert groupid to int
		groupId, err := strconv.ParseUint(u.Gid, 10, 32)
		if err != nil {
			LoggerError.Printf("Cannot convert group to id:%v", err)
			return false
		}

		// add process credentials
		execCmd.SysProcAttr = &syscall.SysProcAttr{}
		execCmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(userId), Gid: uint32(groupId)}
		return true
	}))

	if err != nil {
		LoggerError.Printf("Failed add cron job %v; Error:%v", LoggerError.CronjobToString(cronjob), err)
	} else {
		LoggerInfo.Printf("Add cron job %v", LoggerError.CronjobToString(cronjob))
	}

	return err
}

// Return number of jobs
func (r *Runner) Len() int {
	return len(r.cron.Entries())
}

// Start runner
func (r *Runner) Start() {
	LoggerInfo.Printf("Start runner with %d jobs\n", r.Len())
	r.cron.Start()
}

// Stop runner
func (r *Runner) Stop() {
	r.cron.Stop()
	LoggerInfo.Println("Stop runner")
}

// Execute crontab command
func (r *Runner) cmdFunc(cronjob CrontabEntry, cmdCallback func(*exec.Cmd) bool) func() {
	cmdFunc := func() {
		// fall back to normal shell if not specified
		taskShell := cronjob.Shell
		if taskShell == "" {
			taskShell = DEFAULT_SHELL
		}

		start := time.Now()

		// Init command
		execCmd := exec.Command(taskShell, "-c", cronjob.Command)

		// add custom env to cronjob
		if len(cronjob.Env) >= 1 {
			execCmd.Env = append(os.Environ(), cronjob.Env...)
		}

		// exec custom callback
		if cmdCallback(execCmd) {

			// exec job
			out, err := execCmd.CombinedOutput()

			elapsed := time.Since(start)

			r.prometheus.taskRunDuration.With(r.cronjobToPrometheusLabels(cronjob)).Set(elapsed.Seconds())
			r.prometheus.taskRunTime.With(r.cronjobToPrometheusLabels(cronjob)).SetToCurrentTime()

			if err != nil {
				r.prometheus.taskRunSuccess.With(r.cronjobToPrometheusLabels(cronjob)).Set(0)
				LoggerError.CronjobExecFailed(cronjob, string(out), err, elapsed)
			} else {
				r.prometheus.taskRunSuccess.With(r.cronjobToPrometheusLabels(cronjob)).Set(1)
				LoggerInfo.CronjobExecSuccess(cronjob, string(out), err, elapsed)
			}
		}
	}
	return cmdFunc
}

func (r *Runner) cronjobToPrometheusLabels(cronjob CrontabEntry) (labels prometheus.Labels) {
	labels = prometheus.Labels{
		"cronSpec": cronjob.Spec,
		"cronUser": cronjob.User,
		"cronCommand": cronjob.Command,
	}
	return
}
