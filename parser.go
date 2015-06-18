package main

import (
	"bufio"
	"errors"
	"io"
	"regexp"
	"strings"
)

const (
	LINE_RE = `^(\S+\s+\S+\s+\S+\s+\S+\s+\S+)\s+(.+)$`
)

type Parser struct {
	rp     *regexp.Regexp
	reader io.Reader
	runner *Runner
}

func NewParser(reader io.Reader) (*Parser, error) {
	rp, err := regexp.Compile(LINE_RE)
	if err != nil {
		return nil, err
	}

	p := &Parser{
		rp:     rp,
		reader: reader,
		runner: NewRunner(),
	}

	return p, nil
}

func (p *Parser) Parse() (*Runner, error) {
	if err := p.parseLines(); err != nil {
		return nil, err
	}
	return p.runner, nil
}

func (p *Parser) parseLines() error {
	scanner := bufio.NewScanner(p.reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") {
			continue
		}
		if p.rp.MatchString(line) == true {
			m := p.rp.FindStringSubmatch(line)
			p.runner.Add(m[1], m[2])
		}
	}

	if p.runner.Len() > 0 {
		return nil
	}
	return errors.New("No parse line")
}
