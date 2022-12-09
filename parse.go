package main

import (
	"fmt"
	"regexp"
	"strconv"
)

var (
	rxIssue   = regexp.MustCompile(`/issues/(\d+)$`)
	rxReflink = regexp.MustCompile(`/link/(\d+)/web$`)
)

func parseIssueNum(url string) (int, error) {
	return parsePathInt(url, rxIssue)
}

func parseReflinkID(url string) (int, error) {
	return parsePathInt(url, rxReflink)
}

func parsePathInt(url string, rx *regexp.Regexp) (int, error) {
	m := rx.FindStringSubmatch(url)
	if m == nil {
		return 0, fmt.Errorf("no match: %q", url)
	}
	x, err := strconv.Atoi(m[1])
	if err != nil {
		panic("should never happen")
	}
	return x, nil
}
