package main

import (
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	cron "github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

type Runner struct {
	cron     *cron.Cron
	cronjobs map[cron.EntryID]*CrontabEntry
}

func NewRunner() *Runner {
	r := &Runner{
		cron: cron.New(
			cron.WithParser(
				cron.NewParser(
					cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
				),
			),
		),
		cronjobs: map[cron.EntryID]*CrontabEntry{},
	}
	return r
}

// Add crontab entry
func (r *Runner) Add(cronjob CrontabEntry) error {
	eid, err := r.cron.AddFunc(cronjob.Spec, r.cmdFunc(&cronjob, func(execCmd *exec.Cmd) bool {
		// before exec callback
		log.WithFields(LogCronjobToFields(cronjob)).Infof("executing")
		return true
	}))

	if err != nil {
		prometheusMetricTask.With(r.cronjobToPrometheusLabels(cronjob)).Set(0)
		log.WithFields(LogCronjobToFields(cronjob)).Errorf("cronjob failed adding:%v", err)
	} else {
		cronjob.SetEntryId(eid)
		r.cronjobs[eid] = &cronjob
		prometheusMetricTask.With(r.cronjobToPrometheusLabels(cronjob)).Set(1)
		log.WithFields(LogCronjobToFields(cronjob)).Infof("cronjob added")
	}

	return err
}

// Add crontab entry with user
func (r *Runner) AddWithUser(cronjob CrontabEntry) error {
	eid, err := r.cron.AddFunc(cronjob.Spec, r.cmdFunc(&cronjob, func(execCmd *exec.Cmd) bool {
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
		prometheusMetricTask.With(r.cronjobToPrometheusLabels(cronjob)).Set(0)
		log.WithFields(LogCronjobToFields(cronjob)).Errorf("cronjob failed adding: %v", err)
	} else {
		cronjob.SetEntryId(eid)
		r.cronjobs[eid] = &cronjob
		prometheusMetricTask.With(r.cronjobToPrometheusLabels(cronjob)).Set(1)
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
	log.Infof("start runner with %d jobs\n", r.Len())
	r.cron.Start()
	r.initAllCronEntryMetrics()
}

// Stop runner
func (r *Runner) Stop() {
	r.cron.Stop()
	log.Infof("stop runner")
}

// Execute crontab command
func (r *Runner) cmdFunc(cronjob *CrontabEntry, cmdCallback func(*exec.Cmd) bool) func() {
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

			cronjobMetricCommonLables := r.cronjobToPrometheusLabels(*cronjob)
			prometheusMetricTaskRunDuration.With(cronjobMetricCommonLables).Set(elapsed.Seconds())
			prometheusMetricTaskRunTime.With(cronjobMetricCommonLables).SetToCurrentTime()

			logFields := LogCronjobToFields(*cronjob)
			logFields["elapsed_s"] = elapsed.Seconds()
			if execCmd.ProcessState != nil {
				logFields["exitCode"] = execCmd.ProcessState.ExitCode()
			}

			if err != nil {
				prometheusMetricTaskRunCount.With(r.cronjobToPrometheusLabels(*cronjob, prometheus.Labels{"result": "error"})).Inc()
				prometheusMetricTaskRunResult.With(cronjobMetricCommonLables).Set(0)
				logFields["result"] = "error"
			} else {
				prometheusMetricTaskRunCount.With(r.cronjobToPrometheusLabels(*cronjob, prometheus.Labels{"result": "success"})).Inc()
				prometheusMetricTaskRunResult.With(cronjobMetricCommonLables).Set(1)
				logFields["result"] = "success"
			}

			r.updateCronEntryMetrics(cronjob)
			log.WithFields(logFields).Info("finished")
			if len(cmdStdout) > 0 {
				log.Debugln(string(cmdStdout))
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

func (r *Runner) updateCronEntryMetrics(cronjob *CrontabEntry) {
	cronjobMetricCommonLables := r.cronjobToPrometheusLabels(*cronjob)
	entry := r.cron.Entry(cronjob.EntryId)

	if entry.Next.IsZero() {
		prometheusMetricTaskRunNextTs.With(cronjobMetricCommonLables).Set(0)
	} else {
		prometheusMetricTaskRunNextTs.With(cronjobMetricCommonLables).Set(float64(entry.Next.Unix()))
	}

	if entry.Prev.IsZero() {
		prometheusMetricTaskRunPrevTs.With(cronjobMetricCommonLables).Set(0)
	} else {
		prometheusMetricTaskRunPrevTs.With(cronjobMetricCommonLables).Set(float64(entry.Prev.Unix()))
	}
}

func (r *Runner) initAllCronEntryMetrics() {
	for _, cronjob := range r.cronjobs {
		r.updateCronEntryMetrics(cronjob)
	}
}
