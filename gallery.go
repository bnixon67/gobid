// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package main

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/bnixon67/webapp/weblogin"
	"github.com/bnixon67/webapp/webutil"
)

// GalleryPageData contains data passed to the HTML template.
type GalleryPageData struct {
	Title   string
	Message string
	User    weblogin.User
	Items   []Item
}

// GalleryHandler displays a gallery of items.
func (app *BidApp) GalleryHandler(w http.ResponseWriter, r *http.Request) {
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

	layout := "Mon Jan 2, 2006 3:04 PM"
	message := fmt.Sprintf("Auction is open from %s through %s",
		app.AuctionStart.Format(layout),
		app.AuctionEnd.Format(layout),
	)

	slog.Info("GalleryHandler",
		"message", message,
		"user", user,
		"len(items)", len(items),
	)

	err = webutil.RenderTemplate(app.Tmpl, w, "gallery.html",
		GalleryPageData{
			Title:   app.Cfg.Name,
			Message: message,
			User:    user,
			Items:   items,
		})
	if err != nil {
		slog.Error("unable to render template", "err", err)
		HttpError(w, http.StatusInternalServerError)
		return
	}
}
