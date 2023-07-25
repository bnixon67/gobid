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
	"net/http"

	weblogin "github.com/bnixon67/go-weblogin"
	"golang.org/x/exp/slog"
)

// GalleryPageData contains data passed to the HTML template.
type GalleryPageData struct {
	Title   string
	Message string
	User    weblogin.User
	Items   []Item
}

// GalleryHandler prints a simple hello message.
func (app *BidApp) GalleryHandler(w http.ResponseWriter, r *http.Request) {
	if !weblogin.ValidMethod(w, r, []string{http.MethodGet}) {
		slog.Warn("invalid", "method", r.Method)
		return
	}

	currentUser, err := weblogin.GetUser(w, r, app.DB)
	if err != nil {
		slog.Error("failed to GetUser", "err", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	items, err := app.BidDB.GetItems()
	if err != nil {
		slog.Error("failed to GetItems", "err", err)
	}

	layout := "Mon Jan 2, 2006 3:04 PM"
	message := fmt.Sprintf("Auction starts %s until %s", app.AuctionStart.Format(layout), app.AuctionEnd.Format(layout))

	// display page
	err = weblogin.RenderTemplate(app.Tmpls, w, "gallery.html",
		GalleryPageData{
			Title:   app.Config.Title,
			Message: message,
			User:    currentUser,
			Items:   items,
		})
	if err != nil {
		slog.Error("unable to RenderTemplate", "err", err)
		return
	}
}
