package main

import (
	"bufio"
	"io"
	"regexp"
	"strings"
)

const (
    //          ----spec-----------------------------------------    --user--  -cmd-
	LINE_RE = `^\s*([^@\s]+\s+\S+\s+\S+\s+\S+\s+\S+|@every\s+\S+)\s+([^\s]+)\s+(.+)$`
)

type CrontabEntry struct {
    Spec    string
    User    string
    Command string
}

type Parser struct {
	rp     *regexp.Regexp
	reader io.Reader
}

func NewParser(reader io.Reader) (*Parser, error) {
	rp, err := regexp.Compile(LINE_RE)
	if err != nil {
		return nil, err
	}

	p := &Parser{
		rp:     rp,
		reader: reader,
	}

	return p, nil
}

func (p *Parser) Parse() ([]CrontabEntry) {
    entries := p.parseLines()

	return entries
}

func (p *Parser) parseLines() ([]CrontabEntry) {
    var entries []CrontabEntry

	scanner := bufio.NewScanner(p.reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

        // comment line
        if strings.HasPrefix(line, "#") {
			continue
		}

		if p.rp.MatchString(line) == true {
			m := p.rp.FindStringSubmatch(line)

            crontabSpec := strings.TrimSpace(m[1])
            crontabUser := strings.TrimSpace(m[2])
            crontabCommand := strings.TrimSpace(m[3])

            entries = append(entries, CrontabEntry{crontabSpec, crontabUser, crontabCommand})
		}
	}

    return entries
}
