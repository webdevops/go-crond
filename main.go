package main

import (
	"github.com/namsral/flag"

//	"log"
)

var (
	crontabPath string
)

func init() {
	flag.StringVar(&crontabPath, "file", "crontab", "crontab file path")
}

func main() {
	flag.Parse()

}
