package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/webdevops/go-crond/config"

	flags "github.com/jessevdk/go-flags"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

const (
	Name                = "go-crond"
	Author              = "webdevops.io"
	CRONTAB_TYPE_SYSTEM = ""
)

var (
	opts      config.Opts
	argparser *flags.Parser

	// Git version information
	gitCommit = "<unknown>"
	gitTag    = "<unknown>"
)

func initArgParser() {
	argparser = flags.NewParser(&opts, flags.Default)
	_, err := argparser.Parse()

	// check if there is an parse error
	if err != nil {
		var flagsErr *flags.Error
		if ok := errors.As(err, &flagsErr); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			fmt.Println()
			argparser.WriteHelp(os.Stdout)
			os.Exit(1)
		}
	}

	// --dumpversion
	if opts.ShowOnlyVersion {
		fmt.Println(gitTag)
		os.Exit(0)
	}

	// --version
	if opts.ShowVersion {
		fmt.Printf("%s version %s (%s)\n", Name, gitTag, gitCommit)
		fmt.Printf("Copyright (C) 2022 %s\n", Author)
		os.Exit(0)
	}

	// --help
	if opts.ShowHelp {
		argparser.WriteHelp(os.Stdout)
		os.Exit(0)
	}

	// verbose level
	if opts.Log.Verbose {
		log.SetLevel(log.DebugLevel)
	}

	// verbose level
	if opts.Log.Json {
		log.SetFormatter(&log.JSONFormatter{})
	}
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

	user := opts.Cron.DefaultUser

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

	if username == CRONTAB_TYPE_SYSTEM {
		parser, err = NewCronjobSystemParser(path)
	} else {
		parser, err = NewCronjobUserParser(path, username)
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
	if opts.Cron.Auto {
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
	if len(opts.Cron.IncludeCronD) >= 1 {
		ret = append(ret, includePathsForCrontabs(opts.Cron.IncludeCronD, CRONTAB_TYPE_SYSTEM)...)
	}

	// --run-parts
	if len(opts.Cron.RunParts) >= 1 {
		for _, runPart := range opts.Cron.RunParts {
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
	if len(opts.Cron.RunParts1m) >= 1 {
		ret = append(ret, includeRunPartsDirectories("@every 1m", opts.Cron.RunParts1m)...)
	}

	// --run-parts-15min
	if len(opts.Cron.RunParts15m) >= 1 {
		ret = append(ret, includeRunPartsDirectories("*/15 * * * *", opts.Cron.RunParts15m)...)
	}

	// --run-parts-hourly
	if len(opts.Cron.RunPartsHourly) >= 1 {
		ret = append(ret, includeRunPartsDirectories("@hourly", opts.Cron.RunPartsHourly)...)
	}

	// --run-parts-daily
	if len(opts.Cron.RunPartsDaily) >= 1 {
		ret = append(ret, includeRunPartsDirectories("@daily", opts.Cron.RunPartsDaily)...)
	}

	// --run-parts-weekly
	if len(opts.Cron.RunPartsWeekly) >= 1 {
		ret = append(ret, includeRunPartsDirectories("@weekly", opts.Cron.RunPartsWeekly)...)
	}

	// --run-parts-monthly
	if len(opts.Cron.RunPartsMonthly) >= 1 {
		ret = append(ret, includeRunPartsDirectories("@monthly", opts.Cron.RunPartsMonthly)...)
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
			ret = append(ret, includePathForCrontabs("/etc/crontabs", opts.Cron.DefaultUser)...)
		}

		systemDetected = true
	}

	// ----------------------
	// RedHat
	// ----------------------
	if checkIfFileExistsAndOwnedByRoot("/etc/redhat-release") {
		log.Infof(" --> detected RedHat family, using distribution defaults")

		if checkIfFileExistsAndOwnedByRoot("/etc/crontabs") {
			ret = append(ret, includePathForCrontabs("/etc/crontabs", CRONTAB_TYPE_SYSTEM)...)
		}

		systemDetected = true
	}

	// ----------------------
	// SuSE
	// ----------------------
	if checkIfFileExistsAndOwnedByRoot("/etc/SuSE-release") {
		log.Infof(" --> detected SuSE family, using distribution defaults")

		if checkIfFileExistsAndOwnedByRoot("/etc/crontab") {
			ret = append(ret, parseCrontab("/etc/crontab", CRONTAB_TYPE_SYSTEM)...)
		}

		systemDetected = true
	}

	// ----------------------
	// Debian
	// ----------------------
	if checkIfFileExistsAndOwnedByRoot("/etc/debian_version") {
		log.Infof(" --> detected Debian family, using distribution defaults")

		if checkIfFileExistsAndOwnedByRoot("/etc/crontab") {
			ret = append(ret, parseCrontab("/etc/crontab", CRONTAB_TYPE_SYSTEM)...)
		}

		systemDetected = true
	}

	// ----------------------
	// General
	// ----------------------
	if !systemDetected {
		if checkIfFileExistsAndOwnedByRoot("/etc/crontab") {
			ret = append(ret, includePathForCrontabs("/etc/crontab", CRONTAB_TYPE_SYSTEM)...)
		}

		if checkIfFileExistsAndOwnedByRoot("/etc/crontabs") {
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
		if opts.Cron.EnableUserSwitching {
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
	initArgParser()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP)

	log.Infof("starting %s version %s (%s; %s) ", Name, gitTag, gitCommit, runtime.Version())
	log.Info(string(opts.GetJson()))

	// check if user switching is possible (have to be root)
	opts.Cron.EnableUserSwitching = true
	currentUser, err := user.Current()
	if err != nil || currentUser.Uid != "0" {
		if opts.Cron.AllowUnprivileged {
			log.Warnln("go-crond is NOT running as root, disabling user switching")
			opts.Cron.EnableUserSwitching = false
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
	if opts.Server.Bind != "" {
		log.Infof("starting http server on %s", opts.Server.Bind)
		startHttpServer()
	}

	// endless daemon-reload loop
	for {
		resetMetrics()

		// change to initial directory for fetching crontabs
		err = os.Chdir(confDir)
		if err != nil {
			log.Fatalf("cannot switch to path %s: %v", confDir, err)
		}

		// create new cron runner
		runner := createCronRunner(opts.Args.Crontabs)
		registerRunnerShutdown(runner)

		// chdir to root to prevent relative path errors
		err = os.Chdir(opts.Cron.WorkDir)
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
		if opts.Server.Bind != "" {
			mux := http.NewServeMux()

			// healthz
			mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
				if _, err := fmt.Fprint(w, "Ok"); err != nil {
					log.Error(err)
				}
			})

			// readyz
			mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
				if _, err := fmt.Fprint(w, "Ok"); err != nil {
					log.Error(err)
				}
			})

			if opts.Server.Metrics {
				mux.Handle("/metrics", promhttp.Handler())
			}

			srv := &http.Server{
				Addr:         opts.Server.Bind,
				Handler:      mux,
				ReadTimeout:  opts.Server.ReadTimeout,
				WriteTimeout: opts.Server.WriteTimeout,
			}
			log.Fatal(srv.ListenAndServe())
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
