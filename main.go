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
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	weblogin "github.com/bnixon67/go-weblogin"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/exp/slog"
)

type BidApp struct {
	*weblogin.App
	*BidDB
	AuctionStart, AuctionEnd time.Time
}

func main() {
	// config file must be passed as argument and not empty
	if len(os.Args) != 2 || os.Args[1] == "" {
		fmt.Printf("%s [CONFIG FILE]\n", os.Args[0])
		return
	}

	// TODO: allow logfile to specified in config file
	configFileName := os.Args[1]
	logFileName := ""
	app, err := weblogin.NewApp(configFileName, logFileName)
	if err != nil {
		slog.Error("failed to create app", "err", err)
		return
	}
	slog.Info("created app", "config", configFileName, "log", logFileName)

	bidApp := BidApp{App: app, BidDB: &BidDB{}}
	bidApp.BidDB.sqlDB = app.DB

	err = bidApp.ConfigAuction()
	if err != nil {
		slog.Error("failed to ConfigAuction", "err", err)
		return
	}
	layout := "Monday January _2, 2006 3:04 PM"
	slog.Info("Auction",
		"Start", bidApp.AuctionStart.Format(layout),
		"IsAuctionStarted", bidApp.IsAuctionStarted(),
		"AuctionEnd", bidApp.AuctionEnd.Format(layout),
		"IsAuctionEnded", bidApp.IsAuctionEnded(),
		"IsAuctionOpen", bidApp.IsAuctionOpen(),
	)

	mux := http.NewServeMux()

	// define HTTP server
	// TODO: add values to config file
	srv := &http.Server{
		Addr:              ":" + app.Config.ServerPort,
		Handler:           &weblogin.LogRequestHandler{Next: mux},
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 30 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	// register handlers
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
	// TODO: define base html directory in config
	mux.HandleFunc("/w3.css", weblogin.ServeFileHandler("html/w3.css"))
	mux.HandleFunc("/favicon.ico", weblogin.ServeFileHandler("html/favicon.ico"))
	mux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images"))))
	mux.Handle("/", http.RedirectHandler("/gallery", http.StatusMovedPermanently))

	// create a channel to receive signals
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// start the server in a goroutine
	go func() {
		slog.Info("server", "addr", srv.Addr)
		// TODO: move cert locations to config file
		err = srv.ListenAndServeTLS("cert/cert.pem", "cert/key.pem")
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", "err", err)
			os.Exit(1)
		}
	}()

	// wait for an interrupt signal
	<-interrupt

	// create a context with a timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// initiate the shutdown process
	err = srv.Shutdown(ctx)
	if err != nil {
		slog.Error("server shutdown error", "err", err)
	}

	slog.Info("server closed")
}
