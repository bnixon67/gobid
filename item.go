// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/bnixon67/webapp/webauth"
	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

// ItemPageData contains data passed to the HTML template.
type ItemPageData struct {
	Title         string
	Message       string
	User          webauth.User
	Item          Item
	IsAuctionOpen bool
	Bids          []Bid
}

// ItemHandler display an item.
func (app *BidApp) ItemHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.RequestLoggerWithFuncName(r)

	// Check if the HTTP method is valid.
	if !webutil.CheckAllowedMethods(w, r, http.MethodGet, http.MethodPost) {
		logger.Error("invalid method")
		return
	}

	if r.URL.Path == "/item" {
		logger.Warn("bad request")
		webutil.RespondWithError(w, http.StatusBadRequest)
		return
	}

	// get idString from URL path
	idString := strings.TrimPrefix(r.URL.Path, "/item/")
	if idString == "" {
		logger.Warn("missing id")
		webutil.RespondWithError(w, http.StatusBadRequest)
		return
	}

	// convert idString to int
	id, err := strconv.Atoi(idString)
	if err != nil {
		logger.Error("unable to convert id", "idString", idString, "err", err)
		webutil.RespondWithError(w, http.StatusBadRequest)
		return
	}

	currentUser, err := app.DB.UserFromRequest(w, r)
	if err != nil {
		logger.Error("failed to GetUser", "err", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case http.MethodGet:
		app.itemGetHandler(w, r, id, currentUser)
	case http.MethodPost:
		app.itemPostHandler(w, r, id, currentUser)
	}
}

func (app *BidApp) itemGetHandler(w http.ResponseWriter, r *http.Request, id int, user webauth.User) {
	// Get logger with request info and function name.
	logger := webhandler.RequestLoggerWithFuncName(r)

	// get item from database
	item, err := app.BidDB.GetItem(id)
	if err != nil {
		logger.Error("unable to GetItem", "id", id, "err", err)
		webutil.RespondWithError(w, http.StatusNotFound)
		return
	}

	// get bids for item from database
	bids, err := app.BidDB.GetBidsForItem(id)
	if err != nil {
		logger.Error("unable to GetBidsForItem", "id", id, "err", err)
		// TODO: what to display to user if this fails
	}

	// display page
	err = webutil.RenderTemplateOrError(app.Tmpl, w, "item.html",
		ItemPageData{
			Title:         app.Cfg.App.Name,
			Message:       "",
			User:          user,
			Item:          item,
			IsAuctionOpen: app.IsAuctionOpen(),
			Bids:          bids,
		})
	if err != nil {
		logger.Error("unable to RenderTemplate", "err", err)
		return
	}

	logger.Info("displayed item", "username", user.Username, "item", item,
		"auction open", app.IsAuctionOpen(), "bids", len(bids))
}

func (app *BidApp) itemPostHandler(w http.ResponseWriter, r *http.Request, id int, user webauth.User) {
	// Get logger with request info and function name.
	logger := webhandler.RequestLoggerWithFuncName(r)

	var msg string
	var err error

	// get bidAmount
	bidAmountStr := r.PostFormValue("bidAmount")
	if bidAmountStr == "" {
		logger.Warn("no bidAmount")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	bidAmount, err := strconv.ParseFloat(bidAmountStr, 64)
	if err != nil {
		msg = "Invalid bid amount."
		logger.Error("unable to parse",
			"bidAmountStr", bidAmountStr,
			"err", err)
	}

	// negative bid
	if bidAmount <= 0 {
		msg = "Invalid bid amount."
		logger.Error("zero or negative bid",
			"bidAmount", bidAmount)
	}

	// invalid user
	if user == (webauth.User{}) {
		msg = "Invalid user."
		logger.Error("invalid user")
	}

	// submit bid if we have a valid user and bidAmount and open Auction
	if user != (webauth.User{}) && bidAmount > 0 && app.IsAuctionOpen() {
		bidResult, err := app.BidDB.PlaceBid(id, bidAmount, user.Username)
		if err != nil {
			logger.Error("unable to PlaceBid",
				"id", id, "bidAmount", bidAmount, "user", user,
				"err", err)
			msg = bidResult.Message
		} else {
			logger.Info("PlaceBid",
				"id", id,
				"bidAmount", bidAmount,
				"user", user,
				"bidResult", bidResult,
			)
			msg = bidResult.Message

			if bidResult.BidPlaced && bidResult.PriorBidder != "" && bidResult.PriorBidder != user.Username {
				user, err := app.DB.UserForName(bidResult.PriorBidder)
				if err != nil {
					logger.Error("unable to GetUserForName",
						"PriorBidder", bidResult.PriorBidder,
						"err", err)
				}

				// get item from database
				// TODO: eliminate extra GetItem call
				item, err := app.BidDB.GetItem(id)
				if err != nil {
					logger.Error("unable to GetItem", "id", id, "err", err)
				}

				emailText := fmt.Sprintf(
					"You have been outbid on %q. Visit %s/item/%d to rebid.",
					item.Title, app.Cfg.Auth.BaseURL, id)

				err = app.Cfg.SMTP.SendMessage(app.Cfg.EmailFrom, []string{user.Email}, app.Cfg.App.Name, emailText)
				if err != nil {
					logger.Error("unable to send email",
						"to", user.Email,
						"err", err)
				}
			}
		}
	} else if !app.IsAuctionOpen() {
		msg = "Auction is not open"
	}

	// get item from database
	item, err := app.BidDB.GetItem(id)
	if err != nil {
		logger.Error("unable to get item", "id", id, "err", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// get bids for item from database
	bids, err := app.BidDB.GetBidsForItem(id)
	if err != nil {
		logger.Error("unable to get bids for item", "id", id, "err", err)
		// TODO: what to display to user if this fails
	}

	// display page
	err = webutil.RenderTemplateOrError(app.Tmpl, w, "item.html",
		ItemPageData{
			Title:         app.Cfg.App.Name,
			Message:       msg,
			User:          user,
			Item:          item,
			IsAuctionOpen: app.IsAuctionOpen(),
			Bids:          bids,
		})
	if err != nil {
		logger.Error("unable to RenderTemplate", "err", err)
		return
	}

	logger.Info("post bid",
		"message", msg,
		"username", user.Username,
		"item", item,
		"auction open", app.IsAuctionOpen(),
		"bids", len(bids),
	)
}
