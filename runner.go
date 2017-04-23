package main

import (
	"github.com/robfig/cron"
	"os/exec"
	"os/user"
	"syscall"
	"strconv"
)

type Runner struct {
	cron *cron.Cron
    jobCount int
}

func NewRunner() *Runner {
	r := &Runner{
		cron: cron.New(),
	}
	return r
}

func (r *Runner) Add(spec string, cmd string) error {
	err := r.cron.AddFunc(spec, r.cmdFunc(cmd, func(execCmd *exec.Cmd) bool {
        LoggerInfo.Printf("cronjob: spec:%v cmd:%v", spec, cmd)
        return true
    }))

    if err != nil {
        LoggerError.Printf("Failed add cron job spec:%v cmd:%v err:%v", spec, cmd, err)
    } else {
        r.jobCount++
        LoggerInfo.Printf("Add cron job spec:%v cmd:%v", spec, cmd)
    }

	return err
}

func (r *Runner) AddWithUser(spec string, username string, cmd string) error {
	err := r.cron.AddFunc(spec, r.cmdFunc(cmd, func(execCmd *exec.Cmd) bool {
        LoggerInfo.Printf("cronjob: spec:%v usr:%v cmd:%v", spec, username, cmd)

        u, err := user.Lookup(username)
        if err != nil {
            LoggerError.Printf("user lookup failed: %v", err)
            return false
        }

        userId, err := strconv.ParseUint(u.Uid, 10, 32)
        if err != nil {
            LoggerError.Printf("Cannot convert user to id:%v", err)
            return false
        }

        groupId, err := strconv.ParseUint(u.Gid, 10, 32)
        if err != nil {
            LoggerError.Printf("Cannot convert group to id:%v", err)
            return false
        }

        execCmd.SysProcAttr = &syscall.SysProcAttr{}
        execCmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(userId), Gid: uint32(groupId)}
        return true
    }))

    if err != nil {
        LoggerError.Printf("Failed add cron job spec:%v cmd:%v usr:%v err:%v", spec, cmd, username, err)
    } else {
        r.jobCount++
        LoggerInfo.Printf("Add cron job spec:%v cmd:%v usr:%v", spec, cmd, username)
    }

	return err
}

func (r *Runner) Len() int {
	return len(r.cron.Entries())
}

func (r *Runner) Start() {
	LoggerInfo.Printf("Start runner with %d jobs\n", r.jobCount)
	r.cron.Start()
}

func (r *Runner) Stop() {
	r.cron.Stop()
	LoggerInfo.Println("Stop runner")
}

func (r *Runner) cmdFunc(cmd string, cmdCallback func(*exec.Cmd) (bool) ) func() {
	cmdFunc := func() {
        execCmd := exec.Command("sh", "-c", cmd)
        if cmdCallback(execCmd) {
            out, err := execCmd.CombinedOutput()

            if err != nil {
                LoggerError.Printf("failed cronjob: cmd:%v out:%v err:%v", cmd, string(out), err)
            }
        }
	}
	return cmdFunc
}
