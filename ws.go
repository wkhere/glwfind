package main

import "regexp"

var rxTailWS = regexp.MustCompile(`(\s)+$`)

var tailWS func(string) bool = rxTailWS.MatchString
