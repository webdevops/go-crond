package main

import (
	"fmt"
	"log"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
    "path/filepath"
    "os/user"
    "strings"
    flags "github.com/jessevdk/go-flags"
)

const (
    Author  = "webdevops.io"
    Version = "0.2.1"
)

const (
    CRONTAB_WITH_USERNAME = ""
)

var opts struct {
    Processes                 int       `           long:"processes"            description:"Number of parallel executions" default:"1"`
    DefaultUser               string    `           long:"default-user"         description:"Default user"                  default:"root"`
    IncludeCronD              []string  `           long:"include"              description:"Include files in directory as system crontabs (with user)"`
    UseSystemDefaults         bool      `           long:"system-defaults"      description:"Include standard paths for distribution"`
    RunParts                  []string  `           long:"run-parts"            description:"Execute files in directory with custom spec (like run-parts; spec-units:ns,us,s,m,h; format:time-spec:path; eg:10s,1m,1h30m)"`
    RunParts1m                []string  `           long:"run-parts-1min"       description:"Execute files in directory every beginning minute (like run-parts)"`
    RunParts15m               []string  `           long:"run-parts-15min"      description:"Execute files in directory every beginning 15 minutes (like run-parts)"`
    RunPartsHourly            []string  `           long:"run-parts-hourly"     description:"Execute files in directory every beginning hour (like run-parts)"`
    RunPartsDaily             []string  `           long:"run-parts-daily"      description:"Execute files in directory every beginning day (like run-parts)"`
    RunPartsWeekly            []string  `           long:"run-parts-weekly"     description:"Execute files in directory every beginning week (like run-parts)"`
    RunPartsMonthly           []string  `           long:"run-parts-monthly"    description:"Execute files in directory every beginning month (like run-parts)"`
    Verbose                   bool      `short:"v"  long:"verbose"              description:"verbose mode"`
    ShowVersion               bool      `short:"V"  long:"version"              description:"show version and exit"`
    ShowHelp                  bool      `short:"h"  long:"help"                 description:"show this help message"`
}

var argparser *flags.Parser
var args []string
func initArgParser() ([]string) {
    var err error
    argparser = flags.NewParser(&opts, flags.PassDoubleDash)
    args, err = argparser.Parse()

    // check if there is an parse error
    if err != nil {
        logFatalErrorAndExit(err, 1)
    }

    // --version
    if (opts.ShowVersion) {
        fmt.Println(fmt.Sprintf("go-crond version %s", Version))
        fmt.Println(fmt.Sprintf("Copyright (C) 2017 %s", Author))
        os.Exit(0)
    }

    // --help
    if (opts.ShowHelp) {
        argparser.WriteHelp(os.Stdout)
        os.Exit(1)
    }

    return args
}

var LoggerInfo *log.Logger
var LoggerVerbose *log.Logger
var LoggerError *log.Logger
func initLogger() {
    LoggerInfo = log.New(os.Stdout, "go-crond: ", 0)
    LoggerError = log.New(os.Stderr, "go-crond: ", 0)

    if opts.Verbose {
        LoggerVerbose = log.New(os.Stdout, "go-crond: ", 0)
    } else {
        LoggerVerbose = log.New(ioutil.Discard, "go-crond: ", 0)
    }
}

func findFilesInPaths(pathlist []string, callback func(os.FileInfo, string)) {
    for i := range pathlist {
        filepath.Walk(pathlist[i], func(path string, f os.FileInfo, err error) error {
            path, _ = filepath.Abs(path)

            if f.IsDir() {
                return nil
            }

            if checkIfFileIsValid(f, path) {
                callback(f, path)
            }

            return nil
        })
    }
}

func findExecutabesInPathes(pathlist []string, callback func(os.FileInfo, string)) {
    findFilesInPaths(pathlist, func(f os.FileInfo, path string) {
        if f.Mode().IsRegular() && (f.Mode().Perm() & 0100 != 0) {
            callback(f, path)
        } else {
            LoggerInfo.Printf("Ignoring non exectuable file %s\n", path)
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

func includeRunPartsDirectories(spec string, user string, paths []string) []CrontabEntry {
    var ret []CrontabEntry
    findExecutabesInPathes(paths, func(f os.FileInfo, path string) {
        ret = append(ret, CrontabEntry{Spec: spec, User: user, Command: path})
    })
    return ret
}

func includeRunPartsDirectory(spec string, user string, path string) []CrontabEntry {
    var ret []CrontabEntry
    var paths []string = []string{path}

    findExecutabesInPathes(paths, func(f os.FileInfo, path string) {
        ret = append(ret, CrontabEntry{Spec: spec, User: user, Command: path})
    })
    return ret
}

func parseCrontab(path string, username string) []CrontabEntry {
	file, err := os.Open(path)
	if err != nil {
		LoggerError.Fatalf("crontab path: %v err:%v", path, err)
	}

	parser, err := NewParser(file, username)
	if err != nil {
		LoggerError.Fatalf("Parser read err: %v", err)
	}

    crontabEntries := parser.Parse()

    return crontabEntries
}

func collectCrontabs(args []string) []CrontabEntry {
    var ret []CrontabEntry

    // args: crontab files as normal arguments
    for i := range args {
        crontabPath := args[i]
        crontabUser := CRONTAB_WITH_USERNAME

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
        ret = append(ret, includePathsForCrontabs(opts.IncludeCronD, CRONTAB_WITH_USERNAME)...)
    }

    // --run-parts
    if len(opts.RunParts) >= 1 {
        for i := range opts.RunParts {
            runPart := opts.RunParts[i]

            if strings.Contains(runPart, ":") {
                split := strings.SplitN(runPart, ":", 2)
                cronSpec, cronPath := split[0], split[1]

                cronUser := opts.DefaultUser
                cronSpec = fmt.Sprintf("@every %s", cronSpec)

                // extract user from path
                if strings.Contains(cronPath, ":") {
                    split := strings.SplitN(cronPath, ":", 2)
                    cronUser, cronPath = split[0], split[1]
                }

                ret = append(ret, includeRunPartsDirectory(cronSpec, cronUser, cronPath)...)
            } else {
                LoggerError.Printf("Ignoring --run-parts because of missing time spec: %s\n", runPart)
            }
        }
    }

    // --run-parts-1min
    if len(opts.RunParts1m) >= 1 {
        ret = append(ret, includeRunPartsDirectories("@every 1m", opts.DefaultUser, opts.RunParts1m)...)
    }

    // --run-parts-15min
    if len(opts.RunParts15m) >= 1 {
        ret = append(ret, includeRunPartsDirectories("*/15 * * * *", opts.DefaultUser, opts.RunParts15m)...)
    }

    // --run-parts-hourly
    if len(opts.RunPartsHourly) >= 1 {
        ret = append(ret, includeRunPartsDirectories("@hourly", opts.DefaultUser, opts.RunPartsHourly)...)
    }

    // --run-parts-daily
    if len(opts.RunPartsDaily) >= 1 {
        ret = append(ret, includeRunPartsDirectories("@daily", opts.DefaultUser, opts.RunPartsDaily)...)
    }

    // --run-parts-weekly
    if len(opts.RunPartsWeekly) >= 1 {
        ret = append(ret, includeRunPartsDirectories("@weekly", opts.DefaultUser, opts.RunPartsWeekly)...)
    }

    // --run-parts-monthly
    if len(opts.RunPartsMonthly) >= 1 {
        ret = append(ret, includeRunPartsDirectories("@monthly", opts.DefaultUser, opts.RunPartsMonthly)...)
    }

    if opts.UseSystemDefaults {
        ret = append(ret, includeSystemDefaults()...)
    }

    return ret
}

func includeSystemDefaults() []CrontabEntry {
    var ret []CrontabEntry

    // ----------------------
    // Alpine
    // ----------------------
    if checkIfFileExistsAndOwnedByRoot("/etc/alpine-release") {
        LoggerInfo.Println(" --> detected Alpine family, using distribution defaults")

        if checkIfDirectoryExists("/etc/crontabs") {
            ret = append(ret, includePathForCrontabs("/etc/crontabs", opts.DefaultUser)...)
        }

        return ret
    }

    // ----------------------
    // RedHat
    // ----------------------
    if checkIfFileExistsAndOwnedByRoot("/etc/redhat-release") {
        LoggerInfo.Println(" --> detected RedHat family, using distribution defaults")

        if checkIfFileExists("/etc/crontabs") {
            ret = append(ret, includePathForCrontabs("/etc/crontabs", CRONTAB_WITH_USERNAME)...)
        }

        if checkIfDirectoryExists("/etc/cron.d") {
            ret = append(ret, includePathForCrontabs("/etc/cron.d", CRONTAB_WITH_USERNAME)...)
        }

        return ret
    }

    // ----------------------
    // SuSE
    // ----------------------
    if checkIfFileExistsAndOwnedByRoot("/etc/SuSE-release") {
        LoggerInfo.Println(" --> detected SuSE family, using distribution defaults")

        if checkIfFileExists("/etc/crontab") {
            ret = append(ret, includePathForCrontabs("/etc/crontab", CRONTAB_WITH_USERNAME)...)
        }

        return ret
    }

    // ----------------------
    // Debian
    // ----------------------
    if checkIfFileExistsAndOwnedByRoot("/etc/redhat-release") {
        LoggerInfo.Println(" --> detected Debian family, using distribution defaults")

        if checkIfFileExists("/etc/crontab") {
            ret = append(ret, includePathForCrontabs("/etc/crontab", CRONTAB_WITH_USERNAME)...)
        }

        return ret
    }

    return ret
}

func main() {
    args := initArgParser()
    initLogger()


    LoggerInfo.Printf("Starting version %s", Version)

    var wg sync.WaitGroup

    enableUserSwitch := true

    currentUser, _ := user.Current()
    if currentUser.Uid != "0" {
        LoggerError.Println("WARNING: go-crond is NOT running as root, disabling user switching")
        enableUserSwitch = false
    }

    crontabEntries := collectCrontabs(args)

	runtime.GOMAXPROCS(opts.Processes)
    runner := NewRunner()

    for i := range crontabEntries {
        crontabEntry := crontabEntries[i]

        if enableUserSwitch {
            runner.AddWithUser(crontabEntry)
        } else {
            runner.Add(crontabEntry)
        }
    }

    registerRunnerShutdown(runner, &wg)
    runner.Start()
    wg.Add(1)
	wg.Wait()

	LoggerInfo.Println("Terminated")
}

func registerRunnerShutdown(runner *Runner, wg *sync.WaitGroup) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		s := <-c
		LoggerInfo.Println("Got signal: ", s)
		runner.Stop()
		wg.Done()
	}()
}
