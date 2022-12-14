package main

import (
	"strings"
	"testing"
)

type tc struct {
	input string
	want  int
	err   string
}

type parseFunc func(string) (int, error)

var tabIssues = []tc{
	{"", 0, "no match"},
	{".", 0, "no match"},
	{"\n", 0, "no match"},
	{"\x00", 0, "no match"},
	{"\x00\x00\x00\x00", 0, "no match"},
	{"\x00\x00\x00\x00\x00", 0, "no match"},
	{"\xFF", 0, "no match"},
	{"\xFF\xFF\xFF\xFF", 0, "no match"},
	{"\xFF\xFF\xFF\xFF\xFF", 0, "no match"},
	{"issues", 0, "no match"},
	{"/issues", 0, "no match"},
	{"/issues/", 0, "no match"},
	{"/issues/aa", 0, "no match"},
	{"/issues/aa1", 0, "no match"},
	{"/issues//", 0, "no match"},
	{"/issues/-1", 0, "no match"},
	{"issues/1", 0, "no match"},
	{"/issues/0", 0, ""},
	{"/issues/1/", 0, "no match"},
	{"/issues/1", 1, ""},
	{"/issues/10", 10, ""},
	{"/issues/1234567890", 1234567890, ""},
	{"/issues/3.14", 0, "no match"},
	{"/issues/1/foo", 0, "no match"},
	{"https://foo.tld/issues/123", 123, ""},
	{"https://foo.tld/issues/123/", 0, "no match"},
}

var tabReflinks = []tc{
	{"", 0, "no match"},
	{".", 0, "no match"},
	{"\n", 0, "no match"},
	{"\x00", 0, "no match"},
	{"\x00\x00\x00\x00", 0, "no match"},
	{"\x00\x00\x00\x00\x00", 0, "no match"},
	{"\xFF", 0, "no match"},
	{"\xFF\xFF\xFF\xFF", 0, "no match"},
	{"\xFF\xFF\xFF\xFF\xFF", 0, "no match"},
	{"aaa", 0, "no match"},
	{"link", 0, "no match"},
	{"/link/aa", 0, "no match"},
	{"/link", 0, "no match"},
	{"/link/aa", 0, "no match"},
	{"/link//web", 0, "no match"},
	{"/link/aa/web", 0, "no match"},
	{"link/0/web", 0, "no match"},
	{"/link/-1/web", 0, "no match"},
	{"/link/1", 0, "no match"},
	{"/link/1/", 0, "no match"},
	{"/link/1/web", 1, ""},
	{"/link/1234567890/web", 1234567890, ""},
	{"/link/3.14/web", 0, "no match"},
	{"/link/1/web/", 0, "no match"},
	{"/link/1/web/foo", 0, "no match"},
	{"https://foo.tld/link/123/web", 123, ""},
	{"https://foo.tld/link/123/web/", 0, "no match"},
}

func testParse(t *testing.T, f parseFunc, tab []tc) {
	t.Helper()

	for i, tc := range tab {
		x, err := f(tc.input)

		switch {
		case tc.err == "" && err != nil:
			t.Errorf("tc#%d unexpected error: %v", i, err)
		case tc.err != "" && err == nil:
			t.Errorf("tc#%d no error, want one containing: %q", i, tc.err)
		case tc.err != "'" && err != nil:
			if !strings.Contains(err.Error(), tc.err) {
				t.Errorf(
					"tc#%d error mismatch\nhave: %v\nwant one containing: %v",
					i, err, tc.err,
				)
			}
		default:
			if x != tc.want {
				t.Errorf("tc#%d mismatch\nhave %v\nwant %v", i, x, tc.want)
			}
		}
	}
}

func TestParseIssueNum(t *testing.T) {
	testParse(t, parseIssueNum, tabIssues)
}

func TestParseReflinkID(t *testing.T) {
	testParse(t, parseReflinkID, tabReflinks)
}

func BenchmarkParseIssueNum(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, tc := range tabIssues {
			x, _ := parseIssueNum(tc.input)
			_ = x
		}
	}
}

func BenchmarkParseReflinkID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, tc := range tabReflinks {
			x, _ := parseReflinkID(tc.input)
			_ = x
		}
	}
}
