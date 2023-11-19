// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package main

import (
	"net/http"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/weblogin"
	"github.com/bnixon67/webapp/webutil"
)

// ItemsPageData contains data passed to the HTML template.
type ItemsPageData struct {
	Title   string
	Message string
	User    weblogin.User
	Items   []Item
}

// ItemsHandler displays all the items in a table.
func (app *BidApp) ItemsHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.GetRequestLoggerWithFunc(r)

	// Check if the HTTP method is valid.
	if !webutil.ValidMethod(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	user, err := app.DB.GetUserFromRequest(w, r)
	if err != nil {
		logger.Error("failed to get user", "err", err)
		HttpError(w, http.StatusInternalServerError)
		return
	}

	if app.BidDB == nil {
		logger.Error("database is nil")
		HttpError(w, http.StatusInternalServerError)
		return
	}

	items, err := app.BidDB.GetItems()
	if err != nil {
		logger.Error("failed to get items", "err", err)
		HttpError(w, http.StatusInternalServerError)
		return
	}

	err = webutil.RenderTemplate(app.Tmpl, w, "items.html",
		ItemsPageData{
			Title:   app.Cfg.Name,
			Message: "",
			User:    user,
			Items:   items,
		})
	if err != nil {
		logger.Error("unable to render template", "err", err)
		HttpError(w, http.StatusInternalServerError)
		return
	}

	logger.Info("displayed items",
		"user", user,
		"len(items)", len(items),
	)
}
