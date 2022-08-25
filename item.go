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
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	weblogin "github.com/bnixon67/go-weblogin"
)

// ItemPageData contains data passed to the HTML template.
type ItemPageData struct {
	Title   string
	Message string
	User    weblogin.User
	Item    Item
}

// ItemHandler display an item.
func (app *BidApp) ItemHandler(w http.ResponseWriter, r *http.Request) {
	validMethods := []string{http.MethodGet, http.MethodPost}
	if !weblogin.ValidMethod(w, r, validMethods) {
		log.Println("invalid method", r.Method)
		return
	}

	// get sessionToken from cookie, if it exists
	var sessionToken string
	c, err := r.Cookie("sessionToken")
	if err != nil {
		if !errors.Is(err, http.ErrNoCookie) {
			log.Println("error getting cookie", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	} else {
		sessionToken = c.Value
	}

	// get user for sessionToken
	var currentUser weblogin.User
	if sessionToken != "" {
		currentUser, err = weblogin.GetUserForSessionToken(app.DB, sessionToken)
		if err != nil {
			log.Printf("failed to get user for session %q: %v", sessionToken, err)
			currentUser = weblogin.User{}
			// delete invalid sessionToken to prevent session fixation
			http.SetCookie(w, &http.Cookie{Name: "sessionToken", Value: "", MaxAge: -1})
		} else {
			log.Println("UserName =", currentUser.UserName)
		}
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

	// display page
	err = weblogin.RenderTemplate(app.Tmpls, w, "item.html",
		ItemPageData{
			Title:   app.Config.Title,
			Message: "",
			User:    user,
			Item:    item,
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

	// submit bid if we have a valid user and bidAmount
	if user != (weblogin.User{}) && bidAmount >= 0 {
		var bidPlaced bool
		var priorBidder string
		var err error

		bidPlaced, msg, priorBidder, err = app.BidDB.PlaceBid(id, bidAmount, user)
		if err != nil {
			log.Printf("unable to PlaceBid: %v", err)
		}

		log.Printf("PlaceBid(%d, %v, %q) = %v, %q, %q, %v",
			id, bidAmount, user, bidPlaced, msg, priorBidder, err)
		if bidPlaced && priorBidder != "" && priorBidder != user.UserName {
			user, err := weblogin.GetUserForName(app.DB, priorBidder)
			if err != nil {
				log.Printf("unable to GetUserForName(%q): %v",
					priorBidder, err)
			}

			// get item from database
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

	// get item from database
	item, err := app.BidDB.GetItem(id)
	if err != nil {
		log.Printf("unable to GetItem(%d), %v", id, err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// display page
	err = weblogin.RenderTemplate(app.Tmpls, w, "item.html",
		ItemPageData{
			Title:   app.Config.Title,
			Message: msg,
			User:    user,
			Item:    item,
		})
	if err != nil {
		log.Printf("error executing template: %v", err)
		return
	}
}
