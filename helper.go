package main

import (
	"crypto/sha256"
	"fmt"
)

func Min(a int64, b int64) int64 {
	if a < b {
		return a
	} else {
		return b
	}
}

func CalculateSha256Of(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}
