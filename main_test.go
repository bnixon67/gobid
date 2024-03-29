// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package main

import (
	"database/sql"
	"os"
	"testing"
)

const (
	driverName     = "mysql"
	dataSourceName = "gobid_test:password@/gobid_test?parseTime=true&multiStatements=true"
	file           = "sql/test.sql"
)

func setup() {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		panic(err)
	}

	script, err := os.ReadFile(file)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(string(script))
	if err != nil {
		panic(err)
	}
}

func teardown() {
}

func TestMain(m *testing.M) {
	setup()
	ret := m.Run()
	if ret == 0 {
		teardown()
	}
	os.Exit(ret)
}
