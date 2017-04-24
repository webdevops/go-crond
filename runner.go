package main

import (
	"github.com/robfig/cron"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"syscall"
	"strconv"
	"strings"
)

type Runner struct {
	cron *cron.Cron
}

func NewRunner() *Runner {
	r := &Runner{
		cron: cron.New(),
	}
	return r
}

func (r *Runner) Add(cronjob CrontabEntry) error {
    cronSpec := cronjob.Spec
    if ! strings.HasPrefix(cronjob.Spec, "@") {
        cronSpec = fmt.Sprintf("0 %s", cronjob.Spec)
    }

	err := r.cron.AddFunc(cronSpec, r.cmdFunc(cronjob, func(execCmd *exec.Cmd) bool {
        LoggerInfo.Printf("cronjob: spec:%v cmd:%v env:%v", cronjob.Spec, cronjob.Command, cronjob.Env)
        return true
    }))

    if err != nil {
        LoggerError.Printf("Failed add cron job spec:%v cmd:%v err:%v", cronjob.Spec, cronjob.Command, err)
    } else {
        LoggerInfo.Printf("Add cron job spec:%v cmd:%v", cronjob.Spec, cronjob.Command)
    }

	return err
}

func (r *Runner) AddWithUser(cronjob CrontabEntry) error {

    cronSpec := cronjob.Spec
    if ! strings.HasPrefix(cronjob.Spec, "@") {
        cronSpec = fmt.Sprintf("0 %s", cronjob.Spec)
    }

	err := r.cron.AddFunc(cronSpec, r.cmdFunc(cronjob, func(execCmd *exec.Cmd) bool {
        LoggerInfo.Printf("cronjob: spec:%v usr:%v cmd:%v env:%v", cronjob.Spec, cronjob.User, cronjob.Command, cronjob.Env)

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
        LoggerError.Printf("Failed add cron job spec:%v cmd:%v usr:%v err:%v", cronjob.Spec, cronjob.Command, cronjob.User, err)
    } else {
        LoggerInfo.Printf("Add cron job spec:%v cmd:%v usr:%v", cronjob.Spec, cronjob.Command, cronjob.User)
    }

	return err
}

func (r *Runner) Len() int {
	return len(r.cron.Entries())
}

func (r *Runner) Start() {
	LoggerInfo.Printf("Start runner with %d jobs\n", r.Len())
	r.cron.Start()
}

func (r *Runner) Stop() {
	r.cron.Stop()
	LoggerInfo.Println("Stop runner")
}

func (r *Runner) cmdFunc(cronjob CrontabEntry, cmdCallback func(*exec.Cmd) (bool) ) func() {
	cmdFunc := func() {
        execCmd := exec.Command(cronjob.Shell, "-c", cronjob.Command)

        // add custom env to cronjob
        if len(cronjob.Env) >= 1 {
            execCmd.Env = append(os.Environ(), cronjob.Env...)
        }

        // exec custom callback
        if cmdCallback(execCmd) {

            // exec job
            out, err := execCmd.CombinedOutput()

            if err != nil {
                LoggerError.Printf("failed cronjob: cmd:%v out:%v err:%v", cronjob.Command, string(out), err)
            }
        }
	}
	return cmdFunc
}
