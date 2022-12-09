package main

import (
	"errors"
)

func parseArgs(args []string) (c config, _ error) {

	// for a start, trivial setup: all args are search term
	// todo: handle flags, at least -h, show usage only on -h,
	//		 show error on empty args.

	if len(args) == 0 {
		return c, errors.New(usage)
	}

	c.terms = args
	return c, nil
}

const usage = `usage: glwfind term1 [termN...]`
