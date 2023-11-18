// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package main

import (
	"log/slog"
	"net/http"

	"github.com/bnixon67/webapp/weblogin"
	"github.com/bnixon67/webapp/webutil"
)

// BidsPageData contains data passed to the HTML template.
type BidsPageData struct {
	Title   string
	Message string
	User    weblogin.User
	Items   []ItemWithBids
}

// BidsHandler displays all of the bids.
func (app *BidApp) BidsHandler(w http.ResponseWriter, r *http.Request) {
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

	itemsWithBids, err := app.BidDB.GetItemsWithBids()
	if err != nil {
		slog.Error("failed to get items with bids", "err", err)
		HttpError(w, http.StatusInternalServerError)
		return

	}

	slog.Info("BidsHandler",
		"user", user,
		"len(itemsWithBids)", len(itemsWithBids),
	)

	err = webutil.RenderTemplate(app.Tmpl, w, "bids.html",
		BidsPageData{
			Title:   app.Cfg.Name,
			Message: "",
			User:    user,
			Items:   itemsWithBids,
		})
	if err != nil {
		slog.Error("unable to RenderTemplate", "err", err)
		HttpError(w, http.StatusInternalServerError)
		return
	}
}
