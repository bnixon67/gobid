// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package main

import (
	"net/http"

	"github.com/bnixon67/webapp/webauth"
	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

// BidsPageData contains data passed to the HTML template.
type BidsPageData struct {
	Title   string
	Message string
	User    webauth.User
	Items   []ItemWithBids
}

// BidsHandler displays all of the bids.
func (app *BidApp) BidsHandler(w http.ResponseWriter, r *http.Request) {
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
		webutil.RespondWithError(w, http.StatusInternalServerError)
		return
	}

	if app.BidDB == nil {
		logger.Error("BidDB is nil")
		webutil.RespondWithError(w, http.StatusInternalServerError)
		return
	}

	itemsWithBids, err := app.BidDB.GetItemsWithBids()
	if err != nil {
		logger.Error("failed to get items with bids", "err", err)
		webutil.RespondWithError(w, http.StatusInternalServerError)
		return

	}

	err = webutil.RenderTemplateOrError(app.Tmpl, w, "bids.html",
		BidsPageData{
			Title:   app.Cfg.App.Name,
			Message: "",
			User:    user,
			Items:   itemsWithBids,
		})
	if err != nil {
		logger.Error("unable to RenderTemplate", "err", err)
		webutil.RespondWithError(w, http.StatusInternalServerError)
		return
	}

	logger.Info("showed bids", "username", user.Username,
		"len(itemsWithBids)", len(itemsWithBids))
}
