// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/weblogin"
	"github.com/bnixon67/webapp/webutil"
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
	if !webutil.ValidMethod(w, r, validMethods...) {
		slog.Error("invalid HTTP method", "method", r.Method)
		return
	}

	user, err := app.DB.GetUserFromRequest(w, r)
	if err != nil {
		slog.Error("failed to get user", "err", err)
		HttpError(w, http.StatusInternalServerError)
		return
	}

	// only allowed by admin users
	if !user.IsAdmin {
		slog.Warn("non-admin user", "user", user)
		HttpError(w, http.StatusUnauthorized)
		return
	}

	// get idString from URL path
	idString := strings.TrimPrefix(r.URL.Path, "/edit/")
	if idString == "" {
		slog.Error("id string is empty")
		HttpError(w, http.StatusBadRequest)
		return
	}

	// convert idString to int
	id, err := strconv.Atoi(idString)
	if err != nil {
		slog.Warn("unable to convert id string to int",
			"idString", idString, "err", err)
		HttpError(w, http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		app.getItemEditHandler(w, r, id, user)
	case http.MethodPost:
		app.postItemEditHandler(w, r, id, user)
	}
}

func (app *BidApp) getItemEditHandler(w http.ResponseWriter, r *http.Request, id int, user weblogin.User) {
	var item Item
	var err error

	if id != 0 {
		item, err = app.BidDB.GetItem(id)
		if err != nil {
			slog.Error("unable to get item", "id", id, "err", err)
			HttpError(w, http.StatusNotFound)
			return
		}
	}

	err = webutil.RenderTemplate(app.Tmpl, w, "edit.html",
		ItemEditPageData{
			Title:   app.Cfg.Name,
			Message: "",
			User:    user,
			Item:    item,
		})
	if err != nil {
		slog.Error("unable to render template", "err", err)
		HttpError(w, http.StatusInternalServerError)
		return
	}
}

func (app *BidApp) postItemEditHandler(w http.ResponseWriter, r *http.Request, id int, user weblogin.User) {
	// Get logger with request info and function name.
	logger := webhandler.GetRequestLoggerWithFunc(r)

	logger.Debug("entered")

	var msg string
	var err error

	// get title
	title := r.PostFormValue("title")
	if title == "" {
		logger.Error("no title")
		HttpError(w, http.StatusBadRequest)
		return
	}

	// get description
	description := r.PostFormValue("description")
	if description == "" {
		logger.Warn("no description")
		HttpError(w, http.StatusBadRequest)
		return
	}

	// get openingBid
	openingBidStr := r.PostFormValue("openingBid")
	if openingBidStr == "" {
		logger.Warn("no openingBid")
		HttpError(w, http.StatusBadRequest)
		return
	}
	openingBid, err := strconv.ParseFloat(openingBidStr, 64)
	if err != nil {
		logger.Error("unable to convert openingBid to float64",
			"openingBidStr", openingBidStr,
			"err", err,
		)
		HttpError(w, http.StatusBadRequest)
		return
	}

	// get minBidIncr
	minBidIncrStr := r.PostFormValue("minBidIncr")
	if minBidIncrStr == "" {
		logger.Warn("no minBidIncr")
		HttpError(w, http.StatusBadRequest)
		return
	}
	minBidIncr, err := strconv.ParseFloat(minBidIncrStr, 64)
	if err != nil {
		logger.Error("unable to convert minBidIncr to float64",
			"minBidIncrStr", minBidIncrStr,
			"err", err,
		)
		HttpError(w, http.StatusBadRequest)
		return
	}

	// get artist
	artist := r.PostFormValue("artist")

	// get imageFileName
	imageFileName := r.PostFormValue("imageFileName")

	// get imageFile
	imageFile, fileHeader, err := r.FormFile("imageFile")
	if err != nil && err != http.ErrMissingFile {
		logger.Error("no imageFile", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	logger.Debug("got image file", "file", imageFile, "fileHeader", fileHeader)

	// new imageFile to upload
	if err != http.ErrMissingFile {
		defer imageFile.Close()

		imageFileName = SafeFileName(fileHeader.Filename, "jpg")

		var err error
		var name string

		name = filepath.Join("images", imageFileName)
		err = SaveScaledJPEG(imageFile, name, 1920, 0)
		if err != nil {
			logger.Error("unable to SaveScaledJPEG",
				"imageFile", imageFile,
				"name", name,
				"err", err)
			msg = err.Error()
		}

		name = filepath.Join("images", "thumbnails", imageFileName)
		err = SaveScaledJPEG(imageFile, name, 480, 0)
		if err != nil {
			logger.Error("unable to SaveScaledJPEG",
				"imageFile", imageFile,
				"name", name,
				"err", err)
			msg = err.Error()
		}
	}

	item := Item{
		ID:            id,
		Title:         title,
		Description:   description,
		OpeningBid:    openingBid,
		MinBidIncr:    minBidIncr,
		Artist:        artist,
		ImageFileName: imageFileName,
	}

	// only continue if msg is null, otherwise there was a prior error
	if msg == "" {
		if id == 0 { // create new item
			newId, err := app.BidDB.CreateItem(item)
			if err != nil {
				msg = "Could not create item"
				logger.Error("unable to CreateItem",
					"item", item, "err", err)
			} else {
				logger.Info("created item", "newId", newId)
				newUrl := fmt.Sprintf("/edit/%d", newId)
				http.Redirect(w, r, newUrl, http.StatusSeeOther)
				return
			}
		} else { // update existing item
			rows, err := app.BidDB.UpdateItem(item)
			if rows > 1 || err != nil {
				msg = "Could not update item"
				logger.Error("unable to UpdateItem",
					"item", item, "rows", rows, "err", err)
			} else {
				msg = "Updated item"
			}
		}
	}

	// get item from database
	item, err = app.BidDB.GetItem(id)
	if err != nil {
		logger.Error("unable to GetItem", "id", id, "err", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// display page
	err = webutil.RenderTemplate(app.Tmpl, w, "edit.html",
		ItemEditPageData{
			Title:   app.Cfg.Name,
			Message: msg,
			User:    user,
			Item:    item,
		})
	if err != nil {
		logger.Error("unable to RenderTemplate", "err", err)
		return
	}
}
