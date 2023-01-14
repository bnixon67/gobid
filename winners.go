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
	"log"
	"net/http"
	"time"

	weblogin "github.com/bnixon67/go-weblogin"
)

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

// WinnerPageData contains data passed to the HTML template.
type WinnerPageData struct {
	Title   string
	Winners []Winner
}

// WinnerHandler prints a simple hello message.
func (app *BidApp) WinnerHandler(w http.ResponseWriter, r *http.Request) {
	if !weblogin.ValidMethod(w, r, []string{http.MethodGet}) {
		log.Println("invalid method", r.Method)
		return
	}

	currentUser, err := weblogin.GetUser(w, r, app.DB)
	if err != nil {
		log.Printf("error getting user: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if currentUser == (weblogin.User{}) {
		HttpError(w, http.StatusUnauthorized)
		return
	}

	winners, err := app.BidDB.GetWinners()
	if err != nil {
		log.Printf("GetWinners failed: %v", err)
	}

	// display page
	err = weblogin.RenderTemplate(app.Tmpls, w, "winners.html",
		WinnerPageData{
			Title:   app.Config.Title,
			Winners: winners,
		})
	if err != nil {
		log.Printf("error executing template: %v", err)
		return
	}
}
