package main

import (
	"fmt"
	"io"
	"net/url"
	"regexp"
)

func driver(w io.Writer, conf *config) (err error) {
	u, err := url.Parse(conf.startURL)
	if err != nil {
		return err
	}

	matchers := make([]*regexp.Regexp, len(conf.terms))
	for i, t := range conf.terms {
		matchers[i], err = regexp.Compile("(?i)" + t)
		if err != nil {
			return err
		}
	}

	for u != nil {
		var next *url.URL
		debug("\rprocessing ", u, "\t\t")
		next, err = find(w, u, matchers)
		if err != nil {
			debugln()
			return fmt.Errorf("%s: %v", u, err)
		}
		u = next
	}
	return nil
}
