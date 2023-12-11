package main

import (
	"bytes"
	"html/template"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/weblogin"
	"github.com/bnixon67/webapp/webutil"
)

func winnersBody(t *testing.T, data WinnerPageData) string {
	tmplName := "winners.html"

	// Initialize FuncMap with the custom function.
	funcMap := template.FuncMap{"ToTimeZone": webutil.ToTimeZone}

	// Directly include the name of the template in New for clarity.
	tmpl := template.New(tmplName).Funcs(funcMap)

	// Construct the template file path.
	tmplFile := filepath.Join("html", tmplName)

	// Parse the template file, checking for errors.
	tmpl, err := tmpl.ParseFiles(tmplFile)
	if err != nil {
		t.Fatalf("could not parse template file '%s': %v", tmplFile, err)
	}

	// Create a buffer to store the rendered HTML.
	var body bytes.Buffer

	// Execute the template with the data and write the result to the buffer.
	err = tmpl.Execute(&body, data)
	if err != nil {
		t.Fatalf("could not execute template: %v", err)
	}

	return body.String()
}

func TestWinnerHandler(t *testing.T) {
	app := AppForTest(t)

	// TODO: better way to define a test user
	token, err := app.LoginUser("test", "password")
	if err != nil {
		t.Errorf("could not login user to get session token")
	}

	user, err := app.DB.GetUserForSessionToken(token.Value)
	if err != nil {
		t.Errorf("could not get user: %v", err)
	}

	// Retrieve the list of winners from the database.
	winners, err := app.BidDB.GetWinners()
	if err != nil {
		t.Fatalf("failed to GetWinners: %v", err)
	}

	tests := []webhandler.TestCase{
		{
			Name:          "Invalid Method",
			Target:        "/winners",
			RequestMethod: http.MethodPatch,
			WantStatus:    http.StatusMethodNotAllowed,
			WantBody:      "PATCH Method Not Allowed\n",
		},
		{
			Name:          "No User",
			Target:        "/winners",
			RequestMethod: http.MethodGet,
			WantStatus:    http.StatusOK,
			WantBody: winnersBody(t, WinnerPageData{
				Title: app.Cfg.Name}),
		},
		{
			Name:          "Valid User",
			Target:        "/winners",
			RequestMethod: http.MethodGet,
			RequestCookies: []http.Cookie{
				{Name: weblogin.SessionTokenCookieName, Value: token.Value},
			},
			WantStatus: http.StatusOK,
			WantBody: winnersBody(t, WinnerPageData{
				Title:   app.Cfg.Name,
				Winners: winners,
				User:    user}),
		},
	}

	// Test the handler using the utility function.
	webhandler.HandlerTestWithCases(t, app.WinnerHandler, tests)
}

func TestWinnersCSVHandler(t *testing.T) {
	app := AppForTest(t)

	// TODO: better way to define a test user
	userToken, err := app.LoginUser("test", "password")
	if err != nil {
		t.Fatalf("could not login user to get session token")
	}
	adminToken, err := app.LoginUser("admin", "password")
	if err != nil {
		t.Fatalf("could not login user to get session token")
	}

	winners, err := app.BidDB.GetWinners()
	if err != nil {
		t.Fatalf("failed to get winners: %v", err)
	}
	var body bytes.Buffer
	err = webutil.SliceOfStructsToCSV(&body, winners)
	if err != nil {
		t.Fatalf("failed SliceOfStructsToCSV: %v", err)
	}

	tests := []webhandler.TestCase{
		{
			Name:          "Invalid Method",
			Target:        "/events",
			RequestMethod: http.MethodPost,
			WantStatus:    http.StatusMethodNotAllowed,
			WantBody:      "POST Method Not Allowed\n",
		},
		{
			Name:          "Valid GET Request without Cookie",
			Target:        "/events",
			RequestMethod: http.MethodGet,
			WantStatus:    http.StatusUnauthorized,
			WantBody:      "Error: Unauthorized\n",
		},
		{
			Name:          "Valid GET Request with Bad Session Token",
			Target:        "/events",
			RequestMethod: http.MethodGet,
			RequestCookies: []http.Cookie{
				{Name: weblogin.SessionTokenCookieName, Value: "foo"},
			},
			WantStatus: http.StatusUnauthorized,
			WantBody:   "Error: Unauthorized\n",
		},
		{
			Name:          "Valid GET Request with Good Session Token - Non Admin",
			Target:        "/events",
			RequestMethod: http.MethodGet,
			RequestCookies: []http.Cookie{
				{Name: weblogin.SessionTokenCookieName, Value: userToken.Value},
			},
			WantStatus: http.StatusUnauthorized,
			WantBody:   "Error: Unauthorized\n",
		},
		{
			Name:          "Valid GET Request with Good Session Token - Admin",
			Target:        "/events",
			RequestMethod: http.MethodGet,
			RequestCookies: []http.Cookie{
				{Name: weblogin.SessionTokenCookieName, Value: adminToken.Value},
			},
			WantStatus: http.StatusOK,
			WantBody:   body.String(),
		},
	}

	// Test the handler using the utility function.
	webhandler.HandlerTestWithCases(t, app.WinnersCSVHandler, tests)
}
