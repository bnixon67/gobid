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
	"log/slog"
	"testing"

	weblogin "github.com/bnixon67/go-weblogin"
	_ "github.com/go-sql-driver/mysql"
)

const TestLogFile = "test.log"

// global to provide a singleton app.
var bidApp *BidApp //nolint

// AppForTest is a helper function that returns an App used for testing.
func AppForTest(t *testing.T) *BidApp {
	if bidApp == nil {
		var err error
		weblogin.InitLog(TestLogFile, slog.LevelDebug, true)
		app, err := weblogin.NewApp("test_config.json")
		if err != nil {
			app = nil

			t.Fatalf("cannot create NewApp, %v", err)
		}

		bidApp = &BidApp{App: app, BidDB: &BidDB{}}
		bidApp.BidDB.sqlDB = app.DB
	}

	return bidApp
}
