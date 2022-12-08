package main

import (
	"fmt"
	"io"
	"os"
)

var debugW io.Writer = os.Stderr

func debug(a ...any) {
	fmt.Fprint(debugW, a...)
}

func debugln(a ...any) {
	fmt.Fprintln(debugW, a...)
}

func debugf(format string, a ...any) {
	fmt.Fprintf(debugW, format, a...)
}
