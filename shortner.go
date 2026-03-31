package main

import (
	"crypto/sha256"
	"encoding/base64"
)

func hashing(long string) []byte {
	temp := sha256.Sum256([]byte(long))
	return temp[:6]
}

func encoder(small []byte) string {
	short := base64.RawURLEncoding.EncodeToString(small)
	return short
}
