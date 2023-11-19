// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/bnixon67/webapp/webapp"
	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/weblog"
	"github.com/bnixon67/webapp/weblogin"
	"github.com/bnixon67/webapp/webserver"
	"github.com/bnixon67/webapp/webutil"
	_ "github.com/go-sql-driver/mysql"
)

type BidApp struct {
	*weblogin.LoginApp
	*BidDB
	AuctionStart, AuctionEnd time.Time
}

const (
	ExitUsage    = iota + 1 // ExitUsage indicates a usage error.
	ExitConfig              // ExitConfig indicates a config error.
	ExitLog                 // ExitLog indicates a log error.
	ExitTemplate            // ExitTemplate indicates a template error.
	ExitDB                  // ExitConfig indicates a database error.
	ExitApp                 // ExitHandler indicates an app error.
	ExitServer              // ExitServer indicates a server error.
)

func toTimeZone(t time.Time, name string) time.Time {
	loc, err := time.LoadLocation(name)
	if err != nil {
		slog.Error("cannot load location", "name", name, "err", err)
		return t
	}
	return t.In(loc)
}

func main() {
	// Check for command line argument with config file.
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s [config file]\n", os.Args[0])
		os.Exit(ExitUsage)
	}

	// Read config.
	cfg, err := weblogin.GetConfigFromFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to get config:", err)
		os.Exit(ExitConfig)
	}

	// Initialize logging.
	err = weblog.Init(weblog.WithFilename(cfg.Log.Filename),
		weblog.WithLogType(cfg.Log.Type),
		weblog.WithLevel(cfg.Log.Level),
		weblog.WithSource(cfg.Log.WithSource))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error initializing logger:", err)
		os.Exit(ExitLog)
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
		fmt.Fprintln(os.Stderr, "failed to initialize database:", err)
		os.Exit(ExitDB)
	}

	// Create the web login app.
	app, err := weblogin.New(webapp.WithAppName(cfg.Name), webapp.WithTemplate(tmpl),
		weblogin.WithConfig(cfg), weblogin.WithDB(db))
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to create new weblogin:", err)
		os.Exit(ExitApp)
	}

	// Embed web login app into BidApp
	bidApp := BidApp{LoginApp: app, BidDB: &BidDB{}}
	bidApp.BidDB.sqlDB = app.DB

	err = bidApp.ConfigAuction()
	if err != nil {
		slog.Error("failed to ConfigAuction", "err", err)
		return
	}

	slog.Info("create app", "bidApp", bidApp)

	// Create a new ServeMux to handle HTTP requests.
	mux := http.NewServeMux()

	// Register handlers
	mux.Handle("/", http.RedirectHandler("/gallery", http.StatusMovedPermanently))
	mux.HandleFunc("/w3.css", webutil.ServeFileHandler("html/w3.css"))
	mux.HandleFunc("/favicon.ico", webutil.ServeFileHandler("html/favicon.ico"))
	mux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images"))))
	mux.HandleFunc("/login", app.LoginHandler)
	mux.HandleFunc("/register", app.RegisterHandler)
	mux.HandleFunc("/logout", app.LogoutHandler)
	mux.HandleFunc("/forgot", app.ForgotHandler)
	mux.HandleFunc("/reset", app.ResetHandler)
	mux.HandleFunc("/users", app.UsersHandler)
	mux.HandleFunc("/gallery", bidApp.GalleryHandler)
	mux.HandleFunc("/items", bidApp.ItemsHandler)
	mux.HandleFunc("/item/", bidApp.ItemHandler)
	mux.HandleFunc("/edit/", bidApp.ItemEditHandler)
	mux.HandleFunc("/winners", bidApp.WinnerHandler)
	mux.HandleFunc("/bids", bidApp.BidsHandler)

	// Create the web server.
	srv, err := webserver.New(
		webserver.WithAddr(cfg.Server.Host+":"+cfg.Server.Port),
		webserver.WithHandler(webhandler.AddRequestID(webhandler.AddRequestLogger(webhandler.LogRequest(mux)))),
		webserver.WithTLS(cfg.Server.CertFile, cfg.Server.KeyFile),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating server:", err)
		os.Exit(ExitServer)
	}

	// Create a new context.
	ctx := context.Background()

	// Start the web server.
	err = srv.Start(ctx)
	if err != nil {
		slog.Error("error running server", "err", err)
		fmt.Fprintln(os.Stderr, "Error running server:", err)
		os.Exit(ExitServer)
	}
}
