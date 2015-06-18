package main

import (
	"regexp"
	"strings"
	"testing"
)

func TestParser(t *testing.T) {

	s := `
# Test
*/10 * * * * echo "hello world"
* */1 * * * echo "hello world"
`
	strReader := strings.NewReader(s)
	p, err := NewParser(strReader)

	if err != nil {
		t.Error(err)
	}

	if p == nil {
		t.Error("parser is nil!")
	}

}

func TestParserRegex(t *testing.T) {
	rp := regexp.MustCompile(LINE_RE)
	s := `*/10 * * * * echo "hello world"`
	f := rp.MatchString(s)
	if f == false {
		t.Error("not match")
		return
	}

	r := rp.FindStringSubmatch(s)
	t.Logf("%#v", r)

}
