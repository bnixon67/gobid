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
	"github.com/bnixon67/webapp/webauth"
	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/weblog"
	"github.com/bnixon67/webapp/webserver"
	"github.com/bnixon67/webapp/webutil"
	_ "github.com/go-sql-driver/mysql"
)

type BidApp struct {
	*webauth.AuthApp
	*BidDB
	AuctionStart, AuctionEnd time.Time
	MailFrom                 string
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

func main() {
	// Check for command line argument with config file.
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s [config file]\n", os.Args[0])
		os.Exit(ExitUsage)
	}

	// Read config.
	cfg, err := webauth.LoadConfigFromJSON(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to load config:", err)
		os.Exit(ExitConfig)
	}

	// Validate config.
	missingFields, err := cfg.MissingFields()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to validate config:", err)
		os.Exit(ExitConfig)
	}
	if len(missingFields) != 0 {
		fmt.Fprintln(os.Stderr, "Missing fields in config", missingFields)
		os.Exit(ExitConfig)
	}

	// Initialize logging.
	err = weblog.Init(cfg.Log)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to init logging:", err)
		os.Exit(ExitLog)
	}

	// Define the custom function
	funcMap := template.FuncMap{
		"ToTimeZone": webutil.ToTimeZone,
	}

	// Initialize templates with custom functions.
	tmpl, err := webutil.TemplatesWithFuncs(cfg.App.TmplPattern, funcMap)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to init templates:", err)
		os.Exit(ExitTemplate)
	}

	// Initialize db
	db, err := webauth.InitDB(cfg.SQL.DriverName, cfg.SQL.DataSourceName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to init db:", err)
		os.Exit(ExitDB)
	}

	// Create the web login app.
	app, err := webauth.NewApp(webapp.WithName(cfg.App.Name), webapp.WithTemplate(tmpl), webauth.WithConfig(*cfg), webauth.WithDB(db))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create new webauth:", err)
		os.Exit(ExitApp)
	}

	// Embed web login app into BidApp
	bidApp := BidApp{AuthApp: app, BidDB: &BidDB{}}
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
	mux.HandleFunc("/w3.css", webhandler.FileHandler("html/w3.css"))
	mux.HandleFunc("/favicon.ico", webhandler.FileHandler("html/favicon.ico"))
	mux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images"))))

	mux.HandleFunc("GET /login", app.LoginGetHandler)
	mux.HandleFunc("POST /login", app.LoginPostHandler)

	mux.HandleFunc("GET /build", app.BuildHandlerGet)

	mux.HandleFunc("/register", app.RegisterHandler)
	mux.HandleFunc("/logout", app.LogoutHandler)
	mux.HandleFunc("/forgot", app.ForgotHandler)
	mux.HandleFunc("/reset", app.ResetHandler)
	mux.HandleFunc("/users", app.UsersHandler)
	mux.HandleFunc("/userscsv", app.UsersCSVHandler)
	mux.HandleFunc("/gallery", bidApp.GalleryHandler)
	mux.HandleFunc("/items", bidApp.ItemsHandler)
	mux.HandleFunc("/item/", bidApp.ItemHandler)
	mux.HandleFunc("/edit/", bidApp.ItemEditHandler)
	mux.HandleFunc("/winners", bidApp.WinnerHandler)
	mux.HandleFunc("/winnerscsv", bidApp.WinnersCSVHandler)
	mux.HandleFunc("/bids", bidApp.BidsHandler)
	mux.HandleFunc("/events", app.EventsHandler)
	mux.HandleFunc("/eventscsv", app.EventsCSVHandler)

	// Create the web server.
	srv, err := webserver.New(
		webserver.WithAddr(cfg.Server.Host+":"+cfg.Server.Port),
		webserver.WithHandler(webhandler.NewRequestIDMiddleware(webhandler.MiddlewareLogger(webhandler.LogRequest(mux)))),
		webserver.WithTLS(cfg.Server.CertFile, cfg.Server.KeyFile),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating server:", err)
		os.Exit(ExitServer)
	}

	// Create a new context.
	ctx := context.Background()

	// Run the web server.
	err = srv.Run(ctx)
	if err != nil {
		slog.Error("error running server", "err", err)
		fmt.Fprintln(os.Stderr, "Error running server:", err)
		os.Exit(ExitServer)
	}
}
