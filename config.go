// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package main

import (
	"fmt"
	"time"
)

const timeLayout = "2006-01-02 15:04:05"

func (app *BidApp) GetTimeConfig(name string) (time.Time, error) {
	var auction_start time.Time

	ci, err := app.BidDB.GetConfigItem(name)
	if err != nil {
		return auction_start, err
	}

	loc := time.Now().Location()
	auction_start, err = time.ParseInLocation(timeLayout, ci.Value, loc)
	if err != nil {
		return auction_start, err
	}

	return auction_start, err
}

func (app *BidApp) ConfigAuction() error {
	var err error

	s := "auction_start"
	app.AuctionStart, err = app.GetTimeConfig(s)
	if err != nil {
		return fmt.Errorf("%q %w", s, err)
	}

	s = "auction_end"
	app.AuctionEnd, err = app.GetTimeConfig(s)
	if err != nil {
		return fmt.Errorf("%q %w", s, err)
	}

	return err
}

func (app *BidApp) IsAuctionStarted() bool {
	return time.Now().After(app.AuctionStart)
}

func (app *BidApp) IsAuctionEnded() bool {
	return time.Now().After(app.AuctionEnd)
}

func (app *BidApp) IsAuctionOpen() bool {
	if app.IsAuctionStarted() && !app.IsAuctionEnded() {
		return true
	}
	return false
}
