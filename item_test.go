package main

import (
	"bytes"
	"html/template"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/bnixon67/webapp/webauth"
	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

func itemBody(t *testing.T, data ItemPageData) string {
	tmplName := "item.html"

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

func TestItemHandler(t *testing.T) {
	app := AppForTest(t)

	// TODO: better way to define a test user
	token, err := app.LoginUser("test", "password")
	if err != nil {
		t.Fatalf("could not login user to get session token: %v", err)
	}
	user, err := app.DB.UserForLoginToken(token.Value)
	if err != nil {
		t.Errorf("could not get user: %v", err)
	}

	id := 1

	// get item from database
	item, err := app.BidDB.GetItem(id)
	if err != nil {
		t.Fatalf("unable to get item: %v", err)
	}

	// get bids for item from database
	bids, err := app.BidDB.GetBidsForItem(id)
	if err != nil {
		t.Fatalf("unable to get bids for item: %v", err)
	}

	tests := []webhandler.TestCase{
		{
			Name:          "Invalid Method",
			Target:        "/item",
			RequestMethod: http.MethodPatch,
			WantStatus:    http.StatusMethodNotAllowed,
			WantBody:      "PATCH Method Not Allowed\n",
		},
		{
			Name:          "Bad Request",
			Target:        "/item",
			RequestMethod: http.MethodGet,
			WantStatus:    http.StatusBadRequest,
			WantBody:      "Error: Bad Request\n",
		},
		{
			Name:          "Missing ID",
			Target:        "/item/",
			RequestMethod: http.MethodGet,
			WantStatus:    http.StatusBadRequest,
			WantBody:      "Error: Bad Request\n",
		},
		{
			Name:          "Invalid ID",
			Target:        "/item/a",
			RequestMethod: http.MethodGet,
			WantStatus:    http.StatusBadRequest,
			WantBody:      "Error: Bad Request\n",
		},
		{
			Name:          "Not Found ID",
			Target:        "/item/99",
			RequestMethod: http.MethodGet,
			WantStatus:    http.StatusNotFound,
			WantBody:      "Error: Not Found\n",
		},
		{
			Name:          "Valid ID without User",
			Target:        "/item/1",
			RequestMethod: http.MethodGet,
			WantStatus:    http.StatusOK,
			WantBody: itemBody(t, ItemPageData{
				Title: app.Cfg.App.Name,
				Item:  item,
				Bids:  bids,
			}),
		},
		{
			Name:          "Valid ID with User",
			Target:        "/item/1",
			RequestMethod: http.MethodGet,
			RequestCookies: []http.Cookie{
				{Name: webauth.LoginTokenCookieName, Value: token.Value},
			},
			WantStatus: http.StatusOK,
			WantBody: itemBody(t, ItemPageData{
				Title: app.Cfg.App.Name,
				User:  user,
				Item:  item,
				Bids:  bids,
			}),
		},
	}

	// Test the handler using the utility function.
	webhandler.TestHandler(t, app.ItemHandler, tests)
}
