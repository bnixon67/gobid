// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestItemsHandler(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		expectedStatus int
		expectedInBody string
	}{
		{
			name:           "ValidGET",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			expectedInBody: "Gallery",
		},
		{
			name:           "InvalidMethod",
			method:         http.MethodPost,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedInBody: "Method Not Allowed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "/items", nil)
			w := httptest.NewRecorder()

			app := AppForTest(t)
			app.ItemsHandler(w, r)

			if w.Code != tc.expectedStatus {
				t.Errorf("expected status %d but got %d", tc.expectedStatus, w.Code)
			}

			if !strings.Contains(w.Body.String(), tc.expectedInBody) {
				t.Errorf("expected %q in body but got %q", tc.expectedInBody, w.Body)
			}
		})
	}
}
