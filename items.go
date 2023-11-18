// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package main

import (
	"log/slog"
	"net/http"

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
	if !webutil.ValidMethod(w, r, http.MethodGet) {
		slog.Error("invalid HTTP method", "method", r.Method)
		return
	}

	user, err := app.DB.GetUserFromRequest(w, r)
	if err != nil {
		slog.Error("failed to get user", "err", err)
		HttpError(w, http.StatusInternalServerError)
		return
	}

	if app.BidDB == nil {
		slog.Error("database is nil")
		HttpError(w, http.StatusInternalServerError)
		return
	}

	items, err := app.BidDB.GetItems()
	if err != nil {
		slog.Error("failed to get items", "err", err)
		HttpError(w, http.StatusInternalServerError)
		return
	}

	slog.Info("ItemsHandler",
		"user", user,
		"len(items)", len(items),
	)

	err = webutil.RenderTemplate(app.Tmpl, w, "items.html",
		ItemsPageData{
			Title:   app.Cfg.Name,
			Message: "",
			User:    user,
			Items:   items,
		})
	if err != nil {
		slog.Error("unable to render template", "err", err)
		HttpError(w, http.StatusInternalServerError)
		return
	}
}
