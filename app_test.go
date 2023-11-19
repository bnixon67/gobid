// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package main

import (
	"fmt"
	"html/template"
	"os"
	"testing"

	"github.com/bnixon67/webapp/webapp"
	"github.com/bnixon67/webapp/weblog"
	"github.com/bnixon67/webapp/weblogin"
	"github.com/bnixon67/webapp/webutil"
	_ "github.com/go-sql-driver/mysql"
)

const TestLogFile = "test.log"

// global to provide a singleton app.
var bidApp *BidApp //nolint

// AppForTest is a helper function that returns an App used for testing.
func AppForTest(t *testing.T) *BidApp {
	if bidApp == nil {
		var err error

		// Read config.
		cfg, err := weblogin.GetConfigFromFile("testdata/config.json")
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to get config:", err)
			os.Exit(ExitConfig)
		}

		// Initialize logging.
		err = weblog.Init(weblog.WithFilename(TestLogFile),
			weblog.WithLevel("debug"),
			weblog.WithSource(true))
		if err != nil {
			t.Fatalf("cannot init logging: %v", err)
		}

		// Define the custom function
		funcMap := template.FuncMap{
			"toTimeZone": toTimeZone,
		}

		// Initialize templates
		tmpl, err := webutil.InitTemplatesWithFuncMap(cfg.ParseGlobPattern, funcMap)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error initializing templates:", err)
			os.Exit(ExitTemplate)
		}

		// Initialize db
		db, err := weblogin.InitDB(cfg.SQL.DriverName, cfg.SQL.DataSourceName)
		if err != nil {
			t.Fatalf("cannot init db: %v", err)
		}

		// Create the web login app.
		app, err := weblogin.New(webapp.WithAppName(cfg.Name), webapp.WithTemplate(tmpl), weblogin.WithConfig(cfg), weblogin.WithDB(db))
		if err != nil {
			t.Fatalf("cannot create app: %v", err)
		}

		// Embed web login app into BidApp
		bidApp = &BidApp{LoginApp: app, BidDB: &BidDB{}}
		bidApp.BidDB.sqlDB = app.DB
	}

	return bidApp
}
