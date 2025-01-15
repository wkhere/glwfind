package main

import (
	"crypto/sha256"
	"encoding/base64"
)

func minihash(s string) string {
	b := sha256.Sum224([]byte(s))
	return base64.RawStdEncoding.EncodeToString(b[:8])
}
