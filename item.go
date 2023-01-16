/*
Copyright 2022 Bill Nixon

Licensed under the Apache License, Version 2.0 (the "License"); you may not use
this file except in compliance with the License.  You may obtain a copy of the
License at http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed
under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
CONDITIONS OF ANY KIND, either express or implied.  See the License for the
specific language governing permissions and limitations under the License.
*/
package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	weblogin "github.com/bnixon67/go-weblogin"
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
	validMethods := []string{http.MethodGet, http.MethodPost}
	if !weblogin.ValidMethod(w, r, validMethods) {
		log.Println("invalid method", r.Method)
		return
	}

	currentUser, err := weblogin.GetUser(w, r, app.DB)
	if err != nil {
		log.Printf("error getting user: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// get idString from URL path
	idString := strings.TrimPrefix(r.URL.Path, "/item/")
	if idString == "" {
		log.Print("id string is empty")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// convert idString to int
	id, err := strconv.Atoi(idString)
	if err != nil {
		log.Printf("unable to convert id string %q to int, %v", idString, err)
		w.WriteHeader(http.StatusBadRequest)
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
	// get item from database
	item, err := app.BidDB.GetItem(id)
	if err != nil {
		log.Printf("unable to GetItem(%d), %v", id, err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// get bids for item from database
	bids, err := app.BidDB.GetBidsForItem(id)
	if err != nil {
		log.Printf("unable to GetBidsForItem(%d), %v", id, err)
		// TODO: what to display to user if this fails
	}

	// display page
	err = weblogin.RenderTemplate(app.Tmpls, w, "item.html",
		ItemPageData{
			Title:         app.Config.Title,
			Message:       "",
			User:          user,
			Item:          item,
			IsAuctionOpen: app.IsAuctionOpen(),
			Bids:          bids,
		})
	if err != nil {
		log.Printf("error executing template: %v", err)
		return
	}
}

func (app *BidApp) postItemHandler(w http.ResponseWriter, r *http.Request, id int, user weblogin.User) {
	var msg string
	var err error

	// get bidAmount
	bidAmountStr := r.PostFormValue("bidAmount")
	if bidAmountStr == "" {
		log.Print("no bidAmount")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	bidAmount, err := strconv.ParseFloat(bidAmountStr, 64)
	if err != nil {
		msg = "Invalid bid amount."
		log.Printf("unable to convert bidAmount to float64")
	}

	// submit bid if we have a valid user and bidAmount and open Auction
	if user != (weblogin.User{}) && bidAmount >= 0 && app.IsAuctionOpen() {
		log.Printf("PlaceBid(%d, %f, %q)", id, bidAmount, user.UserName)
		bidResult, err := app.BidDB.PlaceBid(id, bidAmount, user.UserName)
		if err != nil {
			log.Printf("unable to PlaceBid: %v", err)
			msg = bidResult.Message
		} else {
			log.Printf("PlaceBid(%d, %f, %q) result %+v", id, bidAmount, user.UserName, bidResult)

			if bidResult.BidPlaced && bidResult.PriorBidder != "" && bidResult.PriorBidder != user.UserName {
				user, err := weblogin.GetUserForName(app.DB, bidResult.PriorBidder)
				if err != nil {
					log.Printf("unable to GetUserForName(%q): %v",
						bidResult.PriorBidder, err)
				}

				// get item from database
				// TODO: eliminate extra GetItem call
				item, err := app.BidDB.GetItem(id)
				if err != nil {
					log.Printf("unable to GetItem(%d), %v", id, err)
				}

				emailText := fmt.Sprintf(
					"You have been outbid on %q. Visit %s/item/%d to rebid.",
					item.Title, app.Config.BaseURL, id)

				err = weblogin.SendEmail(app.Config.SMTPUser, app.Config.SMTPPassword,
					app.Config.SMTPHost, app.Config.SMTPPort, user.Email,
					app.Config.Title, emailText)
				if err != nil {
					log.Printf("unable to SendEmail: %v", err)
				}
			}
		}
	} else if !app.IsAuctionOpen() {
		msg = "Auction is not open"
	}

	// get item from database
	item, err := app.BidDB.GetItem(id)
	if err != nil {
		log.Printf("unable to GetItem(%d), %v", id, err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// get bids for item from database
	bids, err := app.BidDB.GetBidsForItem(id)
	if err != nil {
		log.Printf("unable to GetBidsForItem(%d), %v", id, err)
		// TODO: what to display to user if this fails
	}

	// display page
	err = weblogin.RenderTemplate(app.Tmpls, w, "item.html",
		ItemPageData{
			Title:         app.Config.Title,
			Message:       msg,
			User:          user,
			Item:          item,
			IsAuctionOpen: app.IsAuctionOpen(),
			Bids:          bids,
		})
	if err != nil {
		log.Printf("error executing template: %v", err)
		return
	}
}
