// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/bnixon67/webapp/webauth"
)

// Config represents the overall application configuration.
type Config struct {
	webauth.Config // Inherit webapp.Config
	MailFrom       string
}

var (
	ErrConfigRead  = errors.New("failed to read config file")
	ErrConfigParse = errors.New("failed to parse config file")
)

// LoadConfigFromJSON loads configuration settings from a JSON file.
func LoadConfigFromJSON(filepath string) (*Config, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConfigRead, err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConfigParse, err)
	}

	return &config, nil
}

const timeLayout = "2006-01-02 15:04:05 MST"

func (app *BidApp) GetTimeConfig(name string) (time.Time, error) {
	var t time.Time

	ci, err := app.BidDB.GetConfigItem(name)
	if err != nil {
		return t, err
	}

	// TODO: define config variable for timezone
	loc, err := time.LoadLocation("America/Chicago")
	if err != nil {
		return t, err
	}

	t, err = time.ParseInLocation(timeLayout, ci.Value, loc)
	if err != nil {
		return t, err
	}

	return t, err
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
