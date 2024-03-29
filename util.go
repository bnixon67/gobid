// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package main

import (
	"encoding/json"
	"path/filepath"
	"regexp"
	"strings"
)

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

// AsJson returns v as a Json string. If Marshal fails, then the error string is returned instead of a Json string.
func AsJson(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return err.Error()
	}

	return string(b)
}

// AnyEmpty returns true if any of the strings are empty.
func AnyEmpty(strings ...string) bool {
	for _, s := range strings {
		if s == "" {
			return true
		}
	}
	return false
}
