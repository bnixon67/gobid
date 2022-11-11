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
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	weblogin "github.com/bnixon67/go-weblogin"
)

// ItemEditPageData contains data passed to the HTML template.
type ItemEditPageData struct {
	Title   string
	Message string
	User    weblogin.User
	Item    Item
}

// ItemEditHandler display an item.
func (app *BidApp) ItemEditHandler(w http.ResponseWriter, r *http.Request) {
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

	// only allowed by admin users
	if !currentUser.Admin {
		log.Printf("non-admin user: %+v", currentUser)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// get idString from URL path
	idString := strings.TrimPrefix(r.URL.Path, "/edit/")
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
		app.getItemEditHandler(w, r, id, currentUser)
	case http.MethodPost:
		app.postItemEditHandler(w, r, id, currentUser)
	}
}

func (app *BidApp) getItemEditHandler(w http.ResponseWriter, r *http.Request, id int, user weblogin.User) {
	// get item from database
	item, err := app.BidDB.GetItem(id)
	if err != nil {
		log.Printf("unable to GetItem(%d), %v", id, err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// display page
	err = weblogin.RenderTemplate(app.Tmpls, w, "edit.html",
		ItemEditPageData{
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

func (app *BidApp) postItemEditHandler(w http.ResponseWriter, r *http.Request, id int, user weblogin.User) {
	var msg string
	var err error

	// get title
	title := r.PostFormValue("title")
	if title == "" {
		log.Print("no title")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// get description
	description := r.PostFormValue("description")
	if description == "" {
		log.Print("no description")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// get openingBid
	openingBidStr := r.PostFormValue("openingBid")
	if openingBidStr == "" {
		log.Print("no openingBid")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	openingBid, err := strconv.ParseFloat(openingBidStr, 64)
	if err != nil {
		log.Print("unable to convert openingBid to float64:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// get minBidIncr
	minBidIncrStr := r.PostFormValue("minBidIncr")
	if minBidIncrStr == "" {
		log.Print("no minBidIncr")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	minBidIncr, err := strconv.ParseFloat(minBidIncrStr, 64)
	if err != nil {
		log.Print("unable to convert minBidIncr to float64:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// get artist
	artist := r.PostFormValue("artist")
	if artist == "" {
		log.Print("no artist")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// get imageFileName
	imageFileName := r.PostFormValue("imageFileName")
	if imageFileName == "" {
		log.Print("no imageFileName")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// sanitize filename
	imageFileName = strings.ReplaceAll(imageFileName, "\\", "/")
	_, imageFileName = filepath.Split(imageFileName)

	// update item if we have a valid user
	item := Item{
		ID:            id,
		Title:         title,
		Description:   description,
		OpeningBid:    openingBid,
		MinBidIncr:    minBidIncr,
		Artist:        artist,
		ImageFileName: imageFileName,
	}
	rows, err := app.BidDB.UpdateItem(item)
	if rows > 1 || err != nil {
		msg = "Could not update item"
		log.Printf("unable to UpdateItem(%+v), %d, %q", item, rows, err)
	} else {
		msg = "Updated item"
	}

	// get item from database
	item, err = app.BidDB.GetItem(id)
	if err != nil {
		log.Printf("unable to GetItem(%d), %v", id, err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// display page
	err = weblogin.RenderTemplate(app.Tmpls, w, "edit.html",
		ItemEditPageData{
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
