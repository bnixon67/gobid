// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package main

import (
	"net/http"
	"time"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/weblogin"
	"github.com/bnixon67/webapp/webutil"
)

// Winner represents current winners.
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

// WinnerPageData holds the data to be passed to the winners page template.
type WinnerPageData struct {
	Title   string
	User    weblogin.User
	Winners []Winner
}

// WinnerHandler handles requests for the winners page.
func (app *BidApp) WinnerHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.GetRequestLoggerWithFunc(r)

	// Check if the HTTP method is valid.
	if !webutil.ValidMethod(w, r, http.MethodGet) {
		// Method not allowed. Response w updated appropriately.
		logger.Error("invalid method")
		return
	}

	// Get the user from the request.
	user, err := app.DB.GetUserFromRequest(w, r)
	if err != nil {
		logger.Error("failed to GetUser", "err", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Retrieve the list of winners from the database.
	winners, err := app.BidDB.GetWinners()
	if err != nil {
		logger.Error("failed to GetWinners", "err", err)
	}

	// Render page.
	err = webutil.RenderTemplate(app.Tmpl, w, "winners.html",
		WinnerPageData{
			Title:   app.Cfg.Name,
			User:    user,
			Winners: winners,
		})
	if err != nil {
		logger.Error("unable to render page", "err", err)
		return
	}

	logger.Info("displayed winners", "winners", len(winners))
}

// WinnersCSVHandler provides list of the current users as a CSV file.
func (app *BidApp) WinnersCSVHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.GetRequestLoggerWithFunc(r)

	// Check if the HTTP method is valid.
	if !webutil.ValidMethod(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	user, err := app.DB.GetUserFromRequest(w, r)
	if err != nil {
		logger.Error("failed to GetUser", "err", err)
		webutil.HttpError(w, http.StatusInternalServerError)
		return
	}

	if !user.IsAdmin {
		logger.Error("user not authorized", "user", user)
		webutil.HttpError(w, http.StatusUnauthorized)
		return
	}

	winners, err := app.BidDB.GetWinners()
	if err != nil {
		logger.Error("failed to get winners", "err", err)
		webutil.HttpError(w, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment;filename=winners.csv")

	err = webutil.SliceOfStructsToCSV(w, winners)
	if err != nil {
		logger.Error("failed to convert struct to CSV",
			"err", err, "winners", winners)
		webutil.HttpError(w, http.StatusInternalServerError)
		return
	}
}
