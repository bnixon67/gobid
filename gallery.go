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

	weblogin "github.com/bnixon67/go-weblogin"
)

// GalleryPageData contains data passed to the HTML template.
type GalleryPageData struct {
	Message string
	User    weblogin.User
	Items   []Item
}

// GalleryHandler prints a simple hello message.
func (app *BidApp) GalleryHandler(w http.ResponseWriter, r *http.Request) {
	if !weblogin.ValidMethod(w, r, []string{http.MethodGet}) {
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

	items, err := app.BidDB.GetItems()
	if err != nil {
		log.Printf("GetItems failed: %v", err)
	}

	// display page
	err = weblogin.RenderTemplate(app.Tmpls, w, "gallery.html",
		GalleryPageData{Message: "", User: currentUser, Items: items})
	if err != nil {
		log.Printf("error executing template: %v", err)
		return
	}
}
