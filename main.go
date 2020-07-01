package main

import (
	"fmt"
	flags "github.com/jessevdk/go-flags"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

const (
	Name                = "go-crond"
	Author              = "webdevops.io"
	CRONTAB_TYPE_SYSTEM = ""
)

var (
	// Git version information
	gitCommit = "<unknown>"
	gitTag    = "<unknown>"
)

var opts struct {
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
	EnableUserSwitching bool
	ShowVersion         bool `short:"V"  long:"version"       description:"show version and exit"`
	ShowOnlyVersion     bool `long:"dumpversion"              description:"show only version number and exit"`
	ShowHelp            bool `short:"h"  long:"help"          description:"show this help message"`

	// logger
	Verbose bool `short:"v"  long:"verbose"      env:"VERBOSE"  description:"verbose mode"`
	LogJson bool `           long:"log.json"     env:"LOG_JSON" description:"Switch log output to json format"`

	// server settings
	ServerBind    string `long:"server.bind"     env:"SERVER_BIND"     description:"Server address, eg. ':8080' (/healthz and /metrics for prometheus)" default:""`
	ServerMetrics bool   `long:"server.metrics"  env:"SERVER_METRICS"  description:"Enable prometheus metrics (do not use senstive informations in commands -> use environment variables or files for storing these informations)"`
}

var argparser *flags.Parser
var args []string

func initArgParser() []string {
	var err error
	argparser = flags.NewParser(&opts, flags.PassDoubleDash)
	args, err = argparser.Parse()

	// check if there is an parse error
	if err != nil {
		logFatalErrorAndExit(err, 1)
	}

	// --dumpversion
	if opts.ShowOnlyVersion {
		fmt.Println(gitTag)
		os.Exit(0)
	}

	// --version
	if opts.ShowVersion {
		fmt.Println(fmt.Sprintf("%s version %s (%s)", Name, gitTag, gitCommit))
		fmt.Println(fmt.Sprintf("Copyright (C) 2020 %s", Author))
		os.Exit(0)
	}

	// --help
	if opts.ShowHelp {
		argparser.WriteHelp(os.Stdout)
		os.Exit(1)
	}

	// verbose level
	if opts.Verbose {
		log.SetLevel(log.DebugLevel)
	}

	// verbose level
	if opts.LogJson {
		log.SetFormatter(&log.JSONFormatter{})
	}

	return args
}

// Log error object as message
func logFatalErrorAndExit(err error, exitCode int) {
	if err != nil {
		log.Errorln(err)
	} else {
		log.Errorf("Unknown fatal error")
	}
	os.Exit(exitCode)
}

func findFilesInPaths(pathlist []string, callback func(os.FileInfo, string)) {
	for _, path := range pathlist {
		if stat, err := os.Stat(path); err == nil && stat.IsDir() {
			err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
				path, _ = filepath.Abs(path)

				if f.IsDir() {
					return nil
				}

				if checkIfFileIsValid(f, path) {
					callback(f, path)
				}

				return nil
			})
			if err != nil {
				log.Fatal(err)
			}
		} else {
			log.Infof("path %s does not exists\n", path)
		}
	}
}

func findExecutabesInPathes(pathlist []string, callback func(os.FileInfo, string)) {
	findFilesInPaths(pathlist, func(f os.FileInfo, path string) {
		if f.Mode().IsRegular() && (f.Mode().Perm()&0100 != 0) {
			callback(f, path)
		} else {
			log.Infof("ignoring non exectuable file %s\n", path)
		}
	})
}

func includePathsForCrontabs(paths []string, username string) []CrontabEntry {
	var ret []CrontabEntry
	findFilesInPaths(paths, func(f os.FileInfo, path string) {
		entries := parseCrontab(path, username)
		ret = append(ret, entries...)
	})
	return ret
}

func includePathForCrontabs(path string, username string) []CrontabEntry {
	var ret []CrontabEntry
	var paths []string = []string{path}

	findFilesInPaths(paths, func(f os.FileInfo, path string) {
		entries := parseCrontab(path, username)
		ret = append(ret, entries...)
	})
	return ret
}

func includeRunPartsDirectories(spec string, paths []string) []CrontabEntry {
	var ret []CrontabEntry

	for _, path := range paths {
		ret = append(ret, includeRunPartsDirectory(spec, path)...)
	}

	return ret
}

func includeRunPartsDirectory(spec string, path string) []CrontabEntry {
	var ret []CrontabEntry

	user := opts.DefaultUser

	// extract user from path
	if strings.Contains(path, ":") {
		split := strings.SplitN(path, ":", 2)
		user, path = split[0], split[1]
	}

	var paths []string = []string{path}
	findExecutabesInPathes(paths, func(f os.FileInfo, path string) {
		ret = append(ret, CrontabEntry{Spec: spec, User: user, Command: path})
	})
	return ret
}

func parseCrontab(path string, username string) []CrontabEntry {
	var parser *Parser
	var err error

	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("crontab path: %v err:%v", path, err)
	}

	if username == CRONTAB_TYPE_SYSTEM {
		parser, err = NewCronjobSystemParser(file)
	} else {
		parser, err = NewCronjobUserParser(file, username)
	}

	if err != nil {
		log.Fatalf("parser read err: %v", err)
	}

	crontabEntries := parser.Parse()

	return crontabEntries
}

func collectCrontabs(args []string) []CrontabEntry {
	var ret []CrontabEntry

	// include system default crontab
	if opts.Auto {
		ret = append(ret, includeSystemDefaults()...)
	}

	// args: crontab files as normal arguments
	for _, crontabPath := range args {
		crontabUser := CRONTAB_TYPE_SYSTEM

		if strings.Contains(crontabPath, ":") {
			split := strings.SplitN(crontabPath, ":", 2)
			crontabUser, crontabPath = split[0], split[1]
		}

		crontabAbsPath, f := fileGetAbsolutePath(crontabPath)
		if checkIfFileIsValid(f, crontabAbsPath) {
			entries := parseCrontab(crontabAbsPath, crontabUser)
			ret = append(ret, entries...)
		}
	}

	// --include-crond
	if len(opts.IncludeCronD) >= 1 {
		ret = append(ret, includePathsForCrontabs(opts.IncludeCronD, CRONTAB_TYPE_SYSTEM)...)
	}

	// --run-parts
	if len(opts.RunParts) >= 1 {
		for _, runPart := range opts.RunParts {
			if strings.Contains(runPart, ":") {
				split := strings.SplitN(runPart, ":", 2)
				cronSpec, cronPath := split[0], split[1]
				cronSpec = fmt.Sprintf("@every %s", cronSpec)

				ret = append(ret, includeRunPartsDirectory(cronSpec, cronPath)...)
			} else {
				log.Infof("ignoring --run-parts because of missing time spec: %s\n", runPart)
			}
		}
	}

	// --run-parts-1min
	if len(opts.RunParts1m) >= 1 {
		ret = append(ret, includeRunPartsDirectories("@every 1m", opts.RunParts1m)...)
	}

	// --run-parts-15min
	if len(opts.RunParts15m) >= 1 {
		ret = append(ret, includeRunPartsDirectories("*/15 * * * *", opts.RunParts15m)...)
	}

	// --run-parts-hourly
	if len(opts.RunPartsHourly) >= 1 {
		ret = append(ret, includeRunPartsDirectories("@hourly", opts.RunPartsHourly)...)
	}

	// --run-parts-daily
	if len(opts.RunPartsDaily) >= 1 {
		ret = append(ret, includeRunPartsDirectories("@daily", opts.RunPartsDaily)...)
	}

	// --run-parts-weekly
	if len(opts.RunPartsWeekly) >= 1 {
		ret = append(ret, includeRunPartsDirectories("@weekly", opts.RunPartsWeekly)...)
	}

	// --run-parts-monthly
	if len(opts.RunPartsMonthly) >= 1 {
		ret = append(ret, includeRunPartsDirectories("@monthly", opts.RunPartsMonthly)...)
	}

	return ret
}

func includeSystemDefaults() []CrontabEntry {
	var ret []CrontabEntry

	systemDetected := false

	// ----------------------
	// Alpine
	// ----------------------
	if checkIfFileExistsAndOwnedByRoot("/etc/alpine-release") {
		log.Infof(" --> detected Alpine family, using distribution defaults")

		if checkIfDirectoryExists("/etc/crontabs") {
			ret = append(ret, includePathForCrontabs("/etc/crontabs", opts.DefaultUser)...)
		}

		systemDetected = true
	}

	// ----------------------
	// RedHat
	// ----------------------
	if checkIfFileExistsAndOwnedByRoot("/etc/redhat-release") {
		log.Infof(" --> detected RedHat family, using distribution defaults")

		if checkIfFileExists("/etc/crontabs") {
			ret = append(ret, includePathForCrontabs("/etc/crontabs", CRONTAB_TYPE_SYSTEM)...)
		}

		systemDetected = true
	}

	// ----------------------
	// SuSE
	// ----------------------
	if checkIfFileExistsAndOwnedByRoot("/etc/SuSE-release") {
		log.Infof(" --> detected SuSE family, using distribution defaults")

		if checkIfFileExists("/etc/crontab") {
			ret = append(ret, includePathForCrontabs("/etc/crontab", CRONTAB_TYPE_SYSTEM)...)
		}

		systemDetected = true
	}

	// ----------------------
	// Debian
	// ----------------------
	if checkIfFileExistsAndOwnedByRoot("/etc/debian_version") {
		log.Infof(" --> detected Debian family, using distribution defaults")

		if checkIfFileExists("/etc/crontab") {
			ret = append(ret, includePathForCrontabs("/etc/crontab", CRONTAB_TYPE_SYSTEM)...)
		}

		systemDetected = true
	}

	// ----------------------
	// General
	// ----------------------
	if !systemDetected {
		if checkIfFileExists("/etc/crontab") {
			ret = append(ret, includePathForCrontabs("/etc/crontab", CRONTAB_TYPE_SYSTEM)...)
		}

		if checkIfFileExists("/etc/crontabs") {
			ret = append(ret, includePathForCrontabs("/etc/crontabs", CRONTAB_TYPE_SYSTEM)...)
		}
	}

	if checkIfDirectoryExists("/etc/cron.d") {
		ret = append(ret, includePathForCrontabs("/etc/cron.d", CRONTAB_TYPE_SYSTEM)...)
	}

	return ret
}

func createCronRunner(args []string) *Runner {
	crontabEntries := collectCrontabs(args)

	runner := NewRunner()

	for _, crontabEntry := range crontabEntries {
		if opts.EnableUserSwitching {
			if err := runner.AddWithUser(crontabEntry); err != nil {
				log.Fatal(err)
			}
		} else {
			if err := runner.Add(crontabEntry); err != nil {
				log.Fatal(err)
			}
		}
	}

	return runner
}

func main() {
	args := initArgParser()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP)

	log.Infof("starting %s version %s (%s; %s) ", Name, gitTag, gitCommit, runtime.Version())

	// check if user switching is possible (have to be root)
	opts.EnableUserSwitching = true
	currentUser, err := user.Current()
	if err != nil || currentUser.Uid != "0" {
		if opts.AllowUnprivileged {
			log.Warnln("go-crond is NOT running as root, disabling user switching")
			opts.EnableUserSwitching = false
		} else {
			log.Errorln("go-crond is NOT running as root, add option --allow-unprivileged if this is ok")
			os.Exit(1)
		}
	}

	// get current path
	confDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("could not get current path: %v", err)
	}

	// daemon mode
	initMetrics()
	log.Infof("starting http server on %s", opts.ServerBind)
	startHttpServer()

	// endless daemon-reload loop
	for {
		resetMetrics()

		// change to initial directory for fetching crontabs
		err = os.Chdir(confDir)
		if err != nil {
			log.Fatalf("cannot switch to path %s: %v", confDir, err)
		}

		// create new cron runner
		runner := createCronRunner(args)
		registerRunnerShutdown(runner)

		// chdir to root to prevent relative path errors
		err = os.Chdir("/")
		if err != nil {
			log.Fatalf("cannot switch to path %s: %v", confDir, err)
		}

		// start new cron runner
		runner.Start()

		// check if we received SIGHUP and start a new loop
		s := <-c
		log.Infof("Got signal: %v", s)
		runner.Stop()
		log.Infof("Reloading configuration")
	}
}

// start and handle prometheus handler
func startHttpServer() {
	go func() {
		if opts.ServerBind != "" {
			// healthz
			http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
				if _, err := fmt.Fprint(w, "Ok"); err != nil {
					log.Errorf(err.Error())
				}
			})

			if opts.ServerMetrics {
				http.Handle("/metrics", promhttp.Handler())
			}
			if err := http.ListenAndServe(opts.ServerBind, nil); err != nil {
				panic(err)
			}
		}
	}()
}

func registerRunnerShutdown(runner *Runner) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		s := <-c
		log.Infof("got signal: %v", s)
		runner.Stop()

		log.Infof("terminated")
		os.Exit(1)
	}()
}
