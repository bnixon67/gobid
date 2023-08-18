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
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	weblogin "github.com/bnixon67/go-weblogin"
	_ "github.com/go-sql-driver/mysql"
)

type BidApp struct {
	*weblogin.App
	*BidDB
	AuctionStart, AuctionEnd time.Time
}

// keys returns a slice of the keys in the map m.
func keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))

	for k := range m {
		keys = append(keys, k)
	}

	return keys
}

func main() {
	// map of log levels
	logLevels := map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}

	logLevelMsg := fmt.Sprintf("log level [%s]",
		strings.Join(keys(logLevels), "|"))

	// define command-line flags
	configFilename := flag.String("config", "", "config file")
	logFilename := flag.String("log", "", "log file")
	logLevel := flag.String("logLevel", "Info", logLevelMsg)
	logAddSource := flag.Bool("logAddSource", false, "add source code position to log")

	// define custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "The flags are:\n")
		flag.PrintDefaults()
	}

	// parse command-line flags
	flag.Parse()

	// get log level from map
	level, ok := logLevels[strings.ToLower(*logLevel)]
	if !ok {
		flag.Usage()
		fmt.Fprintf(os.Stderr, "logLevel %q is undefined.\n", *logLevel)
		os.Exit(2)
	}

	// configFilename is required
	if *configFilename == "" {
		flag.Usage()
		os.Exit(2)
	}

	// check for additional command-line arguments
	if flag.NArg() > 0 {
		flag.Usage()
		os.Exit(2)
	}

	weblogin.InitLog(*logFilename, level, *logAddSource)

	app, err := weblogin.NewApp(*configFilename)
	if err != nil {
		slog.Error("failed to create app", "err", err)
		return
	}

	bidApp := BidApp{App: app, BidDB: &BidDB{}}
	bidApp.BidDB.sqlDB = app.DB

	err = bidApp.ConfigAuction()
	if err != nil {
		slog.Error("failed to ConfigAuction", "err", err)
		return
	}
	// layout := "Monday January _2, 2006 3:04 PM"
	slog.Info("create app", "bidApp", bidApp)

	mux := http.NewServeMux()

	// define HTTP server
	// TODO: add values to config file
	srv := &http.Server{
		Addr:              ":" + app.Cfg.Server.Port,
		Handler:           weblogin.RequestIDHandler(weblogin.LogRequestHandler(mux)),
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
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// start the server in a goroutine
	go func() {
		slog.Info("starting server",
			slog.Group("srv",
				"Addr", srv.Addr,
				"ReadTimeout (s)", srv.ReadTimeout/time.Second,
				"WriteTimeout (s)", srv.WriteTimeout/time.Second,
				"IdleTimeout (s)", srv.IdleTimeout/time.Second,
				"ReadHeaderTimeout (s)", srv.ReadHeaderTimeout/time.Second,
				"MaxHeaderBytes (kb)", srv.MaxHeaderBytes/1024,
			),
		)
		// TODO: move cert locations to config file
		err = srv.ListenAndServeTLS("cert/cert.pem", "cert/key.pem")
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to start server", "err", err)
			os.Exit(1)
		}
	}()

	// wait for a signal
	<-sigChan

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
