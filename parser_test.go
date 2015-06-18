package main

import (
	"testing"
)

func TestParser(t *testing.T) {

	path := "crontab.example"
	p, err := NewParser(path)

	if err != nil {
		t.Error(err)
	}

	if p == nil {
		t.Error("parser is nil!")
	}

}
