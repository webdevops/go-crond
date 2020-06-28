package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
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
		task            *prometheus.GaugeVec
		taskRunCount    *prometheus.CounterVec
		taskRunResult   *prometheus.GaugeVec
		taskRunTime     *prometheus.GaugeVec
		taskRunDuration *prometheus.GaugeVec
	}
}

func NewRunner() *Runner {
	r := &Runner{
		cron: cron.New(),
	}

	r.prometheus.task = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gocrond_task_info",
			Help: "gocrond task info",
		},
		[]string{"cronSpec", "cronUser", "cronCommand"},
	)
	prometheus.MustRegister(r.prometheus.task)

	r.prometheus.taskRunCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gocrond_task_run_count",
			Help: "gocrond task run count",
		},
		[]string{"cronSpec", "cronUser", "cronCommand", "result"},
	)
	prometheus.MustRegister(r.prometheus.taskRunCount)

	r.prometheus.taskRunResult = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gocrond_task_run_result",
			Help: "gocrond task run result",
		},
		[]string{"cronSpec", "cronUser", "cronCommand"},
	)
	prometheus.MustRegister(r.prometheus.taskRunResult)

	r.prometheus.taskRunTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gocrond_task_run_time",
			Help: "gocrond task last run time",
		},
		[]string{"cronSpec", "cronUser", "cronCommand"},
	)
	prometheus.MustRegister(r.prometheus.taskRunTime)

	r.prometheus.taskRunDuration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gocrond_task_run_duration",
			Help: "gocrond task last run duration",
		},
		[]string{"cronSpec", "cronUser", "cronCommand"},
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
		log.WithFields(LogCronjobToFields(cronjob)).Debugf("executing")
		return true
	}))

	if err != nil {
		r.prometheus.task.With(r.cronjobToPrometheusLabels(cronjob)).Set(0)
		log.WithFields(LogCronjobToFields(cronjob)).Errorf("cronjob failed adding:%v", err)
	} else {
		r.prometheus.task.With(r.cronjobToPrometheusLabels(cronjob)).Set(1)
		log.WithFields(LogCronjobToFields(cronjob)).Infof("cronjob added")
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
		log.WithFields(LogCronjobToFields(cronjob)).Debugf("executing")

		// lookup username
		u, err := user.Lookup(cronjob.User)
		if err != nil {
			log.WithFields(LogCronjobToFields(cronjob)).Errorf("user lookup failed: %v", err)
			return false
		}

		// convert userid to int
		userId, err := strconv.ParseUint(u.Uid, 10, 32)
		if err != nil {
			log.WithFields(LogCronjobToFields(cronjob)).Errorf("Cannot convert user to id:%v", err)
			return false
		}

		// convert groupid to int
		groupId, err := strconv.ParseUint(u.Gid, 10, 32)
		if err != nil {
			log.WithFields(LogCronjobToFields(cronjob)).Errorf("Cannot convert group to id:%v", err)
			return false
		}

		// add process credentials
		execCmd.SysProcAttr = &syscall.SysProcAttr{}
		execCmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(userId), Gid: uint32(groupId)}
		return true
	}))

	if err != nil {
		r.prometheus.task.With(r.cronjobToPrometheusLabels(cronjob)).Set(0)
		log.WithFields(LogCronjobToFields(cronjob)).Errorf("cronjob failed adding: %v", err)
	} else {
		r.prometheus.task.With(r.cronjobToPrometheusLabels(cronjob)).Set(1)
		log.WithFields(LogCronjobToFields(cronjob)).Infof("cronjob added")
	}

	return err
}

// Return number of jobs
func (r *Runner) Len() int {
	return len(r.cron.Entries())
}

// Start runner
func (r *Runner) Start() {
	log.Infof("Start runner with %d jobs\n", r.Len())
	r.cron.Start()
}

// Stop runner
func (r *Runner) Stop() {
	r.cron.Stop()
	log.Infof("Stop runner")
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
			cmdStdout, err := execCmd.CombinedOutput()

			elapsed := time.Since(start)

			r.prometheus.taskRunDuration.With(r.cronjobToPrometheusLabels(cronjob)).Set(elapsed.Seconds())
			r.prometheus.taskRunTime.With(r.cronjobToPrometheusLabels(cronjob)).SetToCurrentTime()

			logFields := LogCronjobToFields(cronjob)
			if err != nil {
				r.prometheus.taskRunCount.With(r.cronjobToPrometheusLabels(cronjob, prometheus.Labels{"result": "error"})).Inc()
				r.prometheus.taskRunResult.With(r.cronjobToPrometheusLabels(cronjob)).Set(0)
				logFields["result"] = "error"
				log.WithFields(logFields).Errorln(string(cmdStdout))
			} else {
				r.prometheus.taskRunCount.With(r.cronjobToPrometheusLabels(cronjob, prometheus.Labels{"result": "success"})).Inc()
				r.prometheus.taskRunResult.With(r.cronjobToPrometheusLabels(cronjob)).Set(1)
				logFields["result"] = "success"
				log.WithFields(logFields).Debugln(string(cmdStdout))
			}
		}
	}
	return cmdFunc
}

func (r *Runner) cronjobToPrometheusLabels(cronjob CrontabEntry, additionalLabels ...prometheus.Labels) (labels prometheus.Labels) {
	labels = prometheus.Labels{
		"cronSpec":    cronjob.Spec,
		"cronUser":    cronjob.User,
		"cronCommand": cronjob.Command,
	}
	for _, additionalLabelValue := range additionalLabels {
		for labelName, labelValue := range additionalLabelValue {
			labels[labelName] = labelValue
		}
	}
	return
}
