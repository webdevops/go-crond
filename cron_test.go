package main

import (
	"testing"
)

func TestReadCronFile(t *testing.T) {
	path := "./crontab"

	c, err := NewCron(path)

	if err != nil {
		t.Error(err)
	}

	if c == nil {
		t.Error("cron is empty")
	}

}
