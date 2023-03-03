package main

import (
	"errors"
	"net"
)

func isNetError(err error) (ok bool) {
	var e net.Error
	return errors.As(err, &e)
}
