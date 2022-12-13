package main

import (
	"testing"
)

var tabTailWS = []struct {
	input string
	want  bool
}{
	{"", false},
	{".", false},
	{"a", false},
	{"aa", false},
	{" ", true},
	{"\t", true},
	{"\n", true},
	{"\n\n", true},
	{"\r", true},
	{"\u00A0", false}, // nbsp, don't report as ws even if it is
	{"\x00", false},
	{"\xFF", false},
	{"\x00\x00\x00\x00", false},
	{"\x00\x00\x00\x00\x00", false},
	{"\xFF\xFF\xFF\xFF", false},
	{"\xFF\xFF\xFF\xFF\xFF", false},
}

func TestTailWS(t *testing.T) {
	for i, tc := range tabTailWS {
		if res := tailWS(tc.input); res != tc.want {
			t.Errorf("tc#%d: input %q gives %v, want %v",
				i, tc.input, res, tc.want,
			)
		}
	}
}

func BenchmarkTailWS(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, tc := range tabTailWS {
			res := tailWS(tc.input)
			_ = res
		}
	}
}
