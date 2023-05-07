package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

const (
	ENV_LINE = `^(\S+)=(\S+)\s*$`

	//                     ----spec------------------------------------    --user--  -cmd-
	CRONJOB_SYSTEM = `^\s*([^@\s]+\s+\S+\s+\S+\s+\S+\s+\S+|@every\s+\S+)\s+([^\s]+)\s+(.+)$`

	//                  ----spec------------------------------------    -cmd-
	CRONJOB_USER = `^\s*([^@\s]+\s+\S+\s+\S+\s+\S+\s+\S+|@every\s+\S+)\s+(.+)$`

	DEFAULT_SHELL = "sh"
)

var (
	envLineRegex       = regexp.MustCompile(ENV_LINE)
	cronjobSystemRegex = regexp.MustCompile(CRONJOB_SYSTEM)
	cronjobUserRegex   = regexp.MustCompile(CRONJOB_USER)
)

type CrontabEntry struct {
	Spec        string
	User        string
	Command     string
	Env         []string
	Shell       string
	CrontabPath string
	EntryId     cron.EntryID
}

type Parser struct {
	cronLineRegex   *regexp.Regexp
	cronjobUsername string
	path            string
}

// Create new crontab parser (user crontab without user specification)
func NewCronjobUserParser(path string, username string) (*Parser, error) {
	p := &Parser{
		cronLineRegex:   cronjobUserRegex,
		path:            path,
		cronjobUsername: username,
	}

	return p, nil
}

// Create new crontab parser (crontab with user specification)
func NewCronjobSystemParser(path string) (*Parser, error) {
	p := &Parser{
		cronLineRegex:   cronjobSystemRegex,
		path:            path,
		cronjobUsername: CRONTAB_TYPE_SYSTEM,
	}

	return p, nil
}

func (e *CrontabEntry) SetEntryId(eid cron.EntryID) {
	(*e).EntryId = eid
}

// Parse crontab
func (p *Parser) Parse() []CrontabEntry {
	entries := p.parseLines()

	return entries
}

// Parse lines from crontab
func (p *Parser) parseLines() []CrontabEntry {
	var (
		entries        []CrontabEntry
		crontabSpec    string
		crontabUser    string
		crontabCommand string
		environment    []string
	)

	reader, err := os.Open(p.path)
	if err != nil {
		log.Fatalf("crontab path: %v err:%v", p.path, err)
	}
	defer reader.Close()

	shell := DEFAULT_SHELL

	specCleanupRegexp := regexp.MustCompile(`\s+`)

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// comment line
		if strings.HasPrefix(line, "#") {
			continue
		}

		// environment line
		if envLineRegex.MatchString(line) {
			m := envLineRegex.FindStringSubmatch(line)
			envName := strings.TrimSpace(m[1])
			envValue := strings.TrimSpace(m[2])

			if envName == "SHELL" {
				// custom shell for command
				shell = envValue
			} else {
				// normal environment variable
				environment = append(environment, fmt.Sprintf("%s=%s", envName, envValue))
			}
		}

		// cronjob line
		if p.cronLineRegex.MatchString(line) {
			m := p.cronLineRegex.FindStringSubmatch(line)

			if p.cronjobUsername == CRONTAB_TYPE_SYSTEM {
				crontabSpec = strings.TrimSpace(m[1])
				crontabUser = strings.TrimSpace(m[2])
				crontabCommand = strings.TrimSpace(m[3])
			} else {
				crontabSpec = strings.TrimSpace(m[1])
				crontabUser = p.cronjobUsername
				crontabCommand = strings.TrimSpace(m[2])
			}

			// shrink white spaces for better handling
			crontabSpec = specCleanupRegexp.ReplaceAllString(crontabSpec, " ")

			entries = append(
				entries,
				CrontabEntry{
					Spec:        crontabSpec,
					User:        crontabUser,
					Command:     crontabCommand,
					Env:         environment,
					Shell:       shell,
					CrontabPath: p.path,
				},
			)
		}
	}

	return entries
}
