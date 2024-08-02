// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package main

import (
	"fmt"
	"net/http"

	"github.com/bnixon67/webapp/webauth"
	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

// GalleryPageData contains data passed to the HTML template.
type GalleryPageData struct {
	Title   string
	Message string
	User    webauth.User
	Items   []Item
}

// GalleryHandler displays a gallery of items.
func (app *BidApp) GalleryHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.RequestLoggerWithFuncName(r)

	// Check if the HTTP method is valid.
	if !webutil.IsMethodOrError(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	user, err := app.DB.UserFromRequest(w, r)
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

	layout := "Mon Jan 2, 2006 3:04 PM MST"
	message := fmt.Sprintf("Auction is open from %s through %s",
		app.AuctionStart.Format(layout),
		app.AuctionEnd.Format(layout),
	)

	logger.Info("GalleryHandler", "username", user.Username)

	err = webutil.RenderTemplateOrError(app.Tmpl, w, "gallery.html",
		GalleryPageData{
			Title:   app.Cfg.App.Name,
			Message: message,
			User:    user,
			Items:   items,
		})
	if err != nil {
		logger.Error("unable to render template", "err", err)
		HttpError(w, http.StatusInternalServerError)
		return
	}

	logger.Info("success", "username", user.Username, "items", len(items))
}
