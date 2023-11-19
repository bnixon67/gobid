// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/weblogin"
	"github.com/bnixon67/webapp/webutil"
)

// ItemPageData contains data passed to the HTML template.
type ItemPageData struct {
	Title         string
	Message       string
	User          weblogin.User
	Item          Item
	IsAuctionOpen bool
	Bids          []Bid
}

// ItemHandler display an item.
func (app *BidApp) ItemHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.GetRequestLoggerWithFunc(r)

	// Check if the HTTP method is valid.
	if !webutil.ValidMethod(w, r, http.MethodGet, http.MethodPost) {
		logger.Error("invalid method")
		return
	}

	if r.URL.Path == "/item" {
		logger.Warn("bad request")
		HttpError(w, http.StatusBadRequest)
		return
	}

	// get idString from URL path
	idString := strings.TrimPrefix(r.URL.Path, "/item/")
	if idString == "" {
		logger.Warn("missing id")
		HttpError(w, http.StatusBadRequest)
		return
	}

	// convert idString to int
	id, err := strconv.Atoi(idString)
	if err != nil {
		logger.Error("unable to convert id", "idString", idString, "err", err)
		HttpError(w, http.StatusBadRequest)
		return
	}

	currentUser, err := app.DB.GetUserFromRequest(w, r)
	if err != nil {
		logger.Error("failed to GetUser", "err", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case http.MethodGet:
		app.getItemHandler(w, r, id, currentUser)
	case http.MethodPost:
		app.postItemHandler(w, r, id, currentUser)
	}
}

func (app *BidApp) getItemHandler(w http.ResponseWriter, r *http.Request, id int, user weblogin.User) {
	// Get logger with request info and function name.
	logger := webhandler.GetRequestLoggerWithFunc(r)

	// get item from database
	item, err := app.BidDB.GetItem(id)
	if err != nil {
		logger.Error("unable to GetItem", "id", id, "err", err)
		HttpError(w, http.StatusNotFound)
		return
	}

	// get bids for item from database
	bids, err := app.BidDB.GetBidsForItem(id)
	if err != nil {
		logger.Error("unable to GetBidsForItem", "id", id, "err", err)
		// TODO: what to display to user if this fails
	}

	// display page
	err = webutil.RenderTemplate(app.Tmpl, w, "item.html",
		ItemPageData{
			Title:         app.Cfg.Name,
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

	logger.Info("displayed item", "username", user.UserName, "item", item,
		"auction open", app.IsAuctionOpen(), "bids", len(bids))
}

func (app *BidApp) postItemHandler(w http.ResponseWriter, r *http.Request, id int, user weblogin.User) {
	// Get logger with request info and function name.
	logger := webhandler.GetRequestLoggerWithFunc(r)

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
		logger.Error("unable to parse", "bidAmountStr", bidAmountStr,
			"err", err)
	}

	// submit bid if we have a valid user and bidAmount and open Auction
	if user != (weblogin.User{}) && bidAmount > 0 && app.IsAuctionOpen() {
		bidResult, err := app.BidDB.PlaceBid(id, bidAmount, user.UserName)
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

			if bidResult.BidPlaced && bidResult.PriorBidder != "" && bidResult.PriorBidder != user.UserName {
				user, err := app.DB.GetUserForName(bidResult.PriorBidder)
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
					item.Title, app.Cfg.BaseURL, id)

				err = weblogin.SendEmail(app.Cfg.SMTP.User, app.Cfg.SMTP.Password, app.Cfg.SMTP.Host, app.Cfg.SMTP.Port, user.Email, app.Cfg.Name, emailText)
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
	err = webutil.RenderTemplate(app.Tmpl, w, "item.html",
		ItemPageData{
			Title:         app.Cfg.Name,
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
		"username", user.UserName,
		"item", item,
		"auction open", app.IsAuctionOpen(),
		"bids", len(bids),
	)
}
