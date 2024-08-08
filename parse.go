package main

import (
	"fmt"
	"path"
	"strconv"
)

// Note that the current implementation
// (as well as the previous one, using regexps)
// doesn't check against the full path.
// So, both forms are allowed:
//   proto://host/issues/:num
//   proto://host/foo/issues/:num
// Actually, the fact that second one is invalid doesn't matter much,
// as we want to simply parse urls coming from glw to get the revelant data.

func parseIssueNum(url string) (int, error) {
	// match /issues/:num

	if hasLastSlash(url) {
		return 0, errNoMatch(url)
	}

	d, s := path.Split(url)
	h, d := path.Split(cutLastSlash(d))

	if d != "issues" || !hasLastSlash(h) {
		return 0, errNoMatch(url)
	}

	return nonNegativeNum(s)
}

func parseReflinkID(url string) (int, error) {
	// match /link/:id/web

	if hasLastSlash(url) {
		return 0, errNoMatch(url)
	}

	d, s := path.Split(url)
	if s != "web" {
		return 0, errNoMatch(url)
	}

	d, s = path.Split(cutLastSlash(d))
	h, d := path.Split(cutLastSlash(d))

	if d != "link" || !hasLastSlash(h) {
		return 0, errNoMatch(url)
	}

	return nonNegativeNum(s)
}

func nonNegativeNum(s string) (int, error) {
	if len(s) > 0 && s[0] == '-' {
		return 0, errNoMatch(s)
	}
	x, err := strconv.Atoi(s)
	if err != nil {
		return 0, errNoMatch(s)
	}
	return x, nil
}

func hasLastSlash(s string) bool {
	return len(s) > 0 && s[len(s)-1] == '/'
}
func cutLastSlash(s string) string {
	if hasLastSlash(s) {
		return s[:len(s)-1]
	}
	return s
}

func errNoMatch(s string) error {
	return fmt.Errorf("no match: %q", s)
}
