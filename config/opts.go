package config

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
)

type (
	Opts struct {
		ShowVersion     bool `short:"V"  long:"version"       description:"show version and exit"`
		ShowOnlyVersion bool `long:"dumpversion"              description:"show only version number and exit"`
		ShowHelp        bool `short:"h"  long:"help"          description:"show this help message"`

		Cron struct {
			DefaultUser         string   `long:"default-user"         description:"Default user"                  default:"root"`
			IncludeCronD        []string `long:"include"              description:"Include files in directory as system crontabs (with user)"`
			Auto                bool     `long:"auto"                 description:"Enable automatic system crontab detection"`
			RunParts            []string `long:"run-parts"            description:"Execute files in directory with custom spec (like run-parts; spec-units:ns,us,s,m,h; format:time-spec:path; eg:10s,1m,1h30m)"`
			RunParts1m          []string `long:"run-parts-1min"       description:"Execute files in directory every beginning minute (like run-parts)"`
			RunParts15m         []string `long:"run-parts-15min"      description:"Execute files in directory every beginning 15 minutes (like run-parts)"`
			RunPartsHourly      []string `long:"run-parts-hourly"     description:"Execute files in directory every beginning hour (like run-parts)"`
			RunPartsDaily       []string `long:"run-parts-daily"      description:"Execute files in directory every beginning day (like run-parts)"`
			RunPartsWeekly      []string `long:"run-parts-weekly"     description:"Execute files in directory every beginning week (like run-parts)"`
			RunPartsMonthly     []string `long:"run-parts-monthly"    description:"Execute files in directory every beginning month (like run-parts)"`
			AllowUnprivileged   bool     `long:"allow-unprivileged"   description:"Allow daemon to run as non root (unprivileged) user"`
			WorkDir             string   `long:"working-directory"    description:"Set the working directory for crontab commands" default:"/"`
			EnableUserSwitching bool
		}

		// logger
		Log struct {
			Verbose bool `short:"v"  long:"verbose"      env:"VERBOSE"  description:"verbose mode"`
			Json    bool `           long:"log.json"     env:"LOG_JSON" description:"Switch log output to json format"`
		}

		// server settings
		Server struct {
			Bind    string `long:"server.bind"     env:"SERVER_BIND"     description:"Server address, eg. ':8080' (/healthz and /metrics for prometheus)" default:""`
			Metrics bool   `long:"server.metrics"  env:"SERVER_METRICS"  description:"Enable prometheus metrics (do not use senstive informations in commands -> use environment variables or files for storing these informations)"`
		}
	}
)

func (o *Opts) GetJson() []byte {
	jsonBytes, err := json.Marshal(o)
	if err != nil {
		log.Panic(err)
	}
	return jsonBytes
}
