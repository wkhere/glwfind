package main

import (
	"fmt"
	"log"
	"os"
)

type config struct {
	offline bool
	terms   []string

	help func()
}

func init() {
	log.SetFlags(0)
}

func run(c *config) (err error) {
	db, err := setupDB()
	if err != nil {
		return err
	}
	defer db.Close()

	if !c.offline {
		err = feedAll(db)
		if err != nil && !isNetError(err) {
			return err
		}
	}

	err = find(os.Stdout, db, c.terms)
	if err != nil {
		return fmt.Errorf("find: %w", err)
	}

	return nil
}

func main() {
	c, err := parseArgs(os.Args[1:])
	if err != nil {
		die(2, err)
	}
	if c.help != nil {
		c.help()
		os.Exit(0)
	}

	err = run(&c)
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
