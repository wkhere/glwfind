package main

import (
	"log"
	"os"
)

type config struct {
	startURL string
	terms    []string
}

func defaultConfig() config {
	return config{
		startURL: "https://golangweekly.com/issues/latest",
	}
}

func init() {
	log.SetPrefix("glwfind: ")
	log.SetFlags(0)
}

func main() {
	c, err := parseArgs(os.Args[1:])
	if err != nil {
		die(2, err)
	}

	err = driver(os.Stdout, &c)
	if err != nil {
		die(1, err)
	}
}

func die(exitcode int, msgs ...any) {
	if len(msgs) > 0 {
		log.Println(msgs...)
	}
	os.Exit(exitcode)
}
