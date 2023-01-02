package main

import (
	"encoding/json"
	"math/rand"
	"path/filepath"
	"regexp"
	"strings"
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

func SafeFileName(name, ext string) string {
	if name == "" || ext == "" {
		return ""
	}

	if ext == "" {
		ext = "ext"
	}

	// all expect alphanumeric [0-9A-Za-z]
	bad := regexp.MustCompile(`[^[:alnum:]]`)

	_, name = filepath.Split(name)
	name = strings.TrimSuffix(name, filepath.Ext(name))
	name = strings.TrimSpace(name)
	name = strings.ToLower(name)
	name = bad.ReplaceAllLiteralString(name, "-")
	maxLen := 255 - len(ext) - 1
	if len(name) > maxLen {
		name = name[:maxLen]
	}

	return name + "." + ext
}

func AsJson(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		err.Error()
	}

	return string(b)
}
