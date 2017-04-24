package main

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

const (
    ENV_LINE =`^(\S+)=(\S+)\s*$`

    //                        ----spec------------------------------------    --user--  -cmd-
	CRONJOB_WITH_USER = `^\s*([^@\s]+\s+\S+\s+\S+\s+\S+\s+\S+|@every\s+\S+)\s+([^\s]+)\s+(.+)$`

    //                           ----spec------------------------------------    -cmd-
	CRONJOB_WITHOUT_USER = `^\s*([^@\s]+\s+\S+\s+\S+\s+\S+\s+\S+|@every\s+\S+)\s+(.+)$`

    DEFAULT_SHELL = "sh"
)

type EnvVar struct {
    Name  string
    Value string
}

type CrontabEntry struct {
    Spec    string
    User    string
    Command string
    Env     []string
    Shell   string
}

type Parser struct {
	cronLineRegex   *regexp.Regexp
	envRegex        *regexp.Regexp
	reader          io.Reader
    cronjobUsername string
}

func NewParser(reader io.Reader, username string) (*Parser, error) {
    var cronLineRegex *regexp.Regexp
    var envRegex *regexp.Regexp
    var err error

    if (username == CRONTAB_WITH_USERNAME) {
        cronLineRegex, err = regexp.Compile(CRONJOB_WITH_USER)
    } else {
        cronLineRegex, err = regexp.Compile(CRONJOB_WITHOUT_USER)
    }

	if err != nil {
		return nil, err
	}

    envRegex, err = regexp.Compile(ENV_LINE)
	if err != nil {
		return nil, err
	}

	p := &Parser{
		cronLineRegex: cronLineRegex,
		envRegex: envRegex,
		reader: reader,
        cronjobUsername: username,
	}

	return p, nil
}

func (p *Parser) Parse() ([]CrontabEntry) {
    entries := p.parseLines()

	return entries
}

func (p *Parser) parseLines() ([]CrontabEntry) {
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
        if p.envRegex.MatchString(line) == true {
            m := p.envRegex.FindStringSubmatch(line)
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
		if p.cronLineRegex.MatchString(line) == true {
			m := p.cronLineRegex.FindStringSubmatch(line)

            if p.cronjobUsername == CRONTAB_WITH_USERNAME {
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

            entries = append(entries, CrontabEntry{Spec:crontabSpec, User:crontabUser, Command:crontabCommand, Env:environment, Shell:shell})
		}
	}

    return entries
}
