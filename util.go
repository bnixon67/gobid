package main

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func RandomFileName(ext string) string {
	charset := []rune("abcdefghijklmnopqrstuvwxyz1234567890")

	s := make([]rune, 8)
	for i := range s {
		s[i] = charset[rand.Intn(len(charset))]
	}
	return string(s) + "." + ext
}
