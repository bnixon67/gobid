/*
Copyright 2022 Bill Nixon

Licensed under the Apache License, Version 2.0 (the "License"); you may not use
this file except in compliance with the License.  You may obtain a copy of the
License at http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed
under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
CONDITIONS OF ANY KIND, either express or implied.  See the License for the
specific language governing permissions and limitations under the License.
*/
package main

import (
	"strings"
	"testing"
)

func TestSafeFileName(t *testing.T) {
	cases := []struct {
		name, ext, expected string
	}{
		{"name.ext", "tst", "name.tst"},
		{"Name.Ext", "tst", "name.tst"},
		{"123.456", "tst", "123.tst"},
		{"", "", ""},
		{"/foo/bar/name.ext", "tst", "name.tst"},
		{"/foo/bar/$name.ext", "tst", "-name.tst"},
		{"Hello, 世界", "tst", "hello----.tst"},
		{strings.Repeat("1234567890", 26), "tst", strings.Repeat("1234567890", 25) + "1.tst"},
	}

	for _, tc := range cases {
		got := SafeFileName(tc.name, tc.ext)
		if got != tc.expected {
			t.Errorf("got %q expected %q for SafeFileName(%q, %q)",
				got, tc.expected, tc.name, tc.ext)
		}
	}
}

func TestAnyEmpty(t *testing.T) {
	cases := []struct {
		strings []string
		want    bool
	}{
		{[]string{}, false},
		{[]string{""}, true},
		{[]string{"one"}, false},
		{[]string{"", ""}, true},
		{[]string{"one", ""}, true},
		{[]string{"", "two"}, true},
		{[]string{"one", "two"}, false},
		{[]string{"", "", ""}, true},
		{[]string{"one", "", ""}, true},
		{[]string{"one", "two", ""}, true},
		{[]string{"one", "", "three"}, true},
		{[]string{"one", "two", "three"}, false},
	}

	for _, tc := range cases {
		got := AnyEmpty(tc.strings...)
		if got != tc.want {
			t.Errorf("AnyEmpty(%q): got %v want %v",
				tc.strings, got, tc.want)
		}
	}
}
