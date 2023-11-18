// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/bnixon67/webapp/weblogin"
	"github.com/bnixon67/webapp/webutil"
)

type Winner struct {
	ID         int
	Title      string
	Artist     string
	CurrentBid float64
	Modified   time.Time
	ModifiedBy string
	Email      string
	FullName   string
}

// WinnerPageData contains data passed to the HTML template.
type WinnerPageData struct {
	Title   string
	Winners []Winner
}

// WinnerHandler prints a simple hello message.
func (app *BidApp) WinnerHandler(w http.ResponseWriter, r *http.Request) {
	if !webutil.ValidMethod(w, r, http.MethodGet) {
		slog.Error("invalid HTTP method", "method", r.Method)
		return
	}

	currentUser, err := app.DB.GetUserFromRequest(w, r)
	if err != nil {
		slog.Error("failed to GetUser", "err", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if currentUser == (weblogin.User{}) {
		HttpError(w, http.StatusUnauthorized)
		return
	}

	winners, err := app.BidDB.GetWinners()
	if err != nil {
		slog.Error("failed to GetWinners", "err", err)
	}

	// display page
	err = webutil.RenderTemplate(app.Tmpl, w, "winners.html",
		WinnerPageData{
			Title:   app.Cfg.Name,
			Winners: winners,
		})
	if err != nil {
		slog.Error("unable to RenderTemplate", "err", err)
		return
	}
}
