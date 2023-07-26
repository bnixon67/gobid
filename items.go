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
	"net/http"

	weblogin "github.com/bnixon67/go-weblogin"
	"golang.org/x/exp/slog"
)

// ItemsPageData contains data passed to the HTML template.
type ItemsPageData struct {
	Title   string
	Message string
	User    weblogin.User
	Items   []Item
}

// ItemsHandler displays all the items in a table.
func (app *BidApp) ItemsHandler(w http.ResponseWriter, r *http.Request) {
	if !weblogin.ValidMethod(w, r, []string{http.MethodGet}) {
		slog.Error("invalid HTTP method", "method", r.Method)
		HttpError(w, http.StatusMethodNotAllowed)
		return
	}

	user, err := weblogin.GetUser(w, r, app.DB)
	if err != nil {
		slog.Error("failed to get user", "err", err)
		HttpError(w, http.StatusInternalServerError)
		return
	}

	if app.BidDB == nil {
		slog.Error("database is nil")
		HttpError(w, http.StatusInternalServerError)
		return
	}

	items, err := app.BidDB.GetItems()
	if err != nil {
		slog.Error("failed to get items", "err", err)
		HttpError(w, http.StatusInternalServerError)
		return
	}

	slog.Info("ItemsHandler",
		"user", user,
		"len(items)", len(items),
	)

	err = weblogin.RenderTemplate(app.Tmpls, w, "items.html",
		ItemsPageData{
			Title:   app.Config.Title,
			Message: "",
			User:    user,
			Items:   items,
		})
	if err != nil {
		slog.Error("unable to render template", "err", err)
		HttpError(w, http.StatusInternalServerError)
		return
	}
}
