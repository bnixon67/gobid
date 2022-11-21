package main

import (
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

	app.AuctionStart, err = app.GetTimeConfig("auction_start")
	if err != nil {
		return err
	}

	app.AuctionEnd, err = app.GetTimeConfig("auction_end")
	if err != nil {
		return err
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
