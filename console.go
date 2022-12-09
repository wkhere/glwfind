package main

import (
	"fmt"
	"io"
	"os"
)

var consoleW io.Writer = os.Stderr

func console(a ...any) {
	fmt.Fprint(consoleW, a...)
}

func consoleln(a ...any) {
	fmt.Fprintln(consoleW, a...)
}

func consolef(format string, a ...any) {
	fmt.Fprintf(consoleW, format, a...)
}
