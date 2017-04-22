package main

import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
    "path/filepath"
    "os/user"
    flags "github.com/jessevdk/go-flags"
)

var opts struct {
    Processes           int       `long:"processes"           description:"Number of parallel executions" default:"1"`
    DefaultUser         string    `long:"default-user"        description:"Default user"                  default:"root"`
    IncludeCronD        []string  `long:"include-crond"       description:"Include files in directory as system crontabs (with user)"`
    IncludeCron15Min    []string  `long:"include-15min"       description:"Include files in directory for 15 min execution"`
    IncludeCronHourly   []string  `long:"include-hourly"      description:"Include files in directory for hourly execution"`
    IncludeCronDaily    []string  `long:"include-daily"       description:"Include files in directory for daily execution"`
    IncludeCronWeekly   []string  `long:"include-weekly"      description:"Include files in directory for weekly execution"`
    IncludeCronMonthly  []string  `long:"include-monthly"     description:"Include files in directory for monthly execution"`
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

    return args
}

var LoggerInfo *log.Logger
var LoggerError *log.Logger
func initLogger() {
    LoggerInfo = log.New(os.Stdout, "go-crond: ", 0)
    LoggerError = log.New(os.Stderr, "go-crond: ", 0)
}

// Log error object as message
func logFatalErrorAndExit(err error, exitCode int) {
    LoggerError.Fatalf("ERROR: %s\n", err)
    os.Exit(exitCode)
}

func checkIfFileIsValid(f os.FileInfo, path string) bool {
    if ! f.Mode().IsRegular() {
        LoggerInfo.Printf("Ignoring non regular file %s\n", path)
        return false
    }

    if f.Mode().Perm() & 0022 != 0 {
        LoggerInfo.Printf("Ignoring file with wrong modes (not xx22) %s\n", path)
        return false
    }

    return true
}

func findFilesInPaths(pathlist []string, callback func(os.FileInfo, string)) {
    for i := range pathlist {
        filepath.Walk(pathlist[i], func(path string, f os.FileInfo, err error) error {
            path, _ = filepath.Abs(path)

            if ! checkIfFileIsValid(f, path) {
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
        }
    })
}

func parseCrontab(path string) []CrontabEntry {
	file, err := os.Open(path)
	if err != nil {
		LoggerError.Fatalf("crontab path: %v err:%v", path, err)
	}

	parser, err := NewParser(file)
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
        crontabFile, err := filepath.Abs(args[i])
        if err != nil {
            LoggerError.Fatalf("Invalid file: %v", err)
        }

        f, err := os.Lstat(crontabFile)
        if err != nil {
            LoggerError.Fatalf("File stats failed: %v", err)
        }
        if checkIfFileIsValid(f, crontabFile) {
            entries := parseCrontab(crontabFile)

            ret = append(ret, entries...)
        }
    }

    // --include-crond
    if len(opts.IncludeCronD) >= 1 {
        findFilesInPaths(opts.IncludeCronD, func(f os.FileInfo, path string) {
            entries := parseCrontab(path)
            ret = append(ret, entries...)
        })
    }

    // --include-15min
    if len(opts.IncludeCron15Min) >= 1 {
        findExecutabesInPathes(opts.IncludeCron15Min, func(f os.FileInfo, path string) {
            ret = append(ret, CrontabEntry{"@every 15m", opts.DefaultUser, path})
        })
    }

    // --include-hourly
    if len(opts.IncludeCronHourly) >= 1 {
        findExecutabesInPathes(opts.IncludeCronHourly, func(f os.FileInfo, path string) {
            ret = append(ret, CrontabEntry{"@hourly", opts.DefaultUser, path})
        })
    }

    // --include-daily
    if len(opts.IncludeCronDaily) >= 1 {
        findExecutabesInPathes(opts.IncludeCronDaily, func(f os.FileInfo, path string) {
            ret = append(ret, CrontabEntry{"@daily", opts.DefaultUser, path})
        })
    }

    // --include-weekly
    if len(opts.IncludeCronWeekly) >= 1 {
        findExecutabesInPathes(opts.IncludeCronWeekly, func(f os.FileInfo, path string) {
            ret = append(ret, CrontabEntry{"@weekly", opts.DefaultUser, path})
        })
    }

    // --include-monthly
    if len(opts.IncludeCronMonthly) >= 1 {
        findExecutabesInPathes(opts.IncludeCronMonthly, func(f os.FileInfo, path string) {
            ret = append(ret, CrontabEntry{"@monthly", opts.DefaultUser, path})
        })
    }

    return ret
}

func main() {
    initLogger()
    args := initArgParser()

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
            runner.AddWithUser(crontabEntry.Spec, crontabEntry.User, crontabEntry.Command)
        } else {
            runner.Add(crontabEntry.Spec, crontabEntry.Command)
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
