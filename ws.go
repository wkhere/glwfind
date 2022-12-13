package main

import (
	"unicode"
	"unicode/utf8"
)

func tailWS(s string) bool {
	r, _ := utf8.DecodeLastRuneInString(s)
	switch r {
	case '\u00A0':
		// used in glw around a special section marker;
		// when dumping, we want extra space after it
		return false
	default:
		return unicode.IsSpace(r)
	}
}
