package main

import (
	"fmt"
)

func parseArgs(args []string) (c config, _ error) {

	const usage = `usage: glwfind [-x|--offline] term1 [termN...]`

	for len(args) > 0 {
		switch arg := args[0]; {

		case arg == "-x" || arg == "--offline":
			c.offline = true
			args = args[1:]

		case arg == "-h" || arg == "--help":
			c.help = func() { fmt.Println(usage) }
			return c, nil

		case len(arg) > 1 && arg[0] == '-':
			return c, fmt.Errorf("unknown flag %s", arg)

		default:
			c.terms = append(c.terms, arg)
			args = args[1:]
		}
	}

	if len(c.terms) == 0 {
		return c, fmt.Errorf("missing terms")
	}

	return c, nil
}
