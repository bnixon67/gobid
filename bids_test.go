package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBidsHandler(t *testing.T) {
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
			r := httptest.NewRequest(tc.method, "/bids", nil)
			w := httptest.NewRecorder()

			app := AppForTest(t)
			app.BidsHandler(w, r)

			if w.Code != tc.expectedStatus {
				t.Errorf("expected status %d but got %d", tc.expectedStatus, w.Code)
			}

			if !strings.Contains(w.Body.String(), tc.expectedInBody) {
				t.Errorf("expected %q in body but got %q", tc.expectedInBody, w.Body)
			}
		})
	}
}
