package main

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
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
	Spec    string
	User    string
	Command string
	Env     []string
	Shell   string
}

type Parser struct {
	cronLineRegex   *regexp.Regexp
	reader          io.Reader
	cronjobUsername string
}

// Create new crontab parser (user crontab without user specification)
func NewCronjobUserParser(reader io.Reader, username string) (*Parser, error) {
	p := &Parser{
		cronLineRegex:   cronjobUserRegex,
		reader:          reader,
		cronjobUsername: username,
	}

	return p, nil
}

// Create new crontab parser (crontab with user specification)
func NewCronjobSystemParser(reader io.Reader) (*Parser, error) {
	p := &Parser{
		cronLineRegex:   cronjobSystemRegex,
		reader:          reader,
		cronjobUsername: CRONTAB_TYPE_SYSTEM,
	}

	return p, nil
}

// Parse crontab
func (p *Parser) Parse() []CrontabEntry {
	entries := p.parseLines()

	return entries
}

// Parse lines from crontab
func (p *Parser) parseLines() []CrontabEntry {
	var entries []CrontabEntry
	var crontabSpec string
	var crontabUser string
	var crontabCommand string
	var environment []string

	shell := DEFAULT_SHELL

	specCleanupRegexp := regexp.MustCompile(`\s+`)

	scanner := bufio.NewScanner(p.reader)
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

			entries = append(entries, CrontabEntry{Spec: crontabSpec, User: crontabUser, Command: crontabCommand, Env: environment, Shell: shell})
		}
	}

	return entries
}
