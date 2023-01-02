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
	"errors"
	"reflect"
	"testing"
	"time"
)

var (
	ct = time.Date(2022, time.December, 30, 0, 0, 0, 0, time.UTC)
	mt = time.Date(2022, time.December, 31, 0, 0, 0, 0, time.UTC)
	mb = "test"

	testItem2 = Item{
		ID:            2,
		Title:         "Item Test",
		Created:       ct.Add(time.Hour * 2),
		Description:   "Item to test GetItem",
		OpeningBid:    10.0,
		MinBidIncr:    2.0,
		Artist:        "ARTIST",
		ImageFileName: "FILENAME",
		MinBid:        10.0,
	}

	testItem3 = Item{
		ID:            3,
		Title:         "Item Test with Bid",
		Created:       ct.Add(time.Hour * 3),
		Modified:      &mt,
		Bidder:        mb,
		Description:   "Item to test GetItem with Bid",
		OpeningBid:    5.0,
		MinBidIncr:    1.0,
		CurrentBid:    15.0,
		Artist:        "Art",
		ImageFileName: "File",
		MinBid:        16.0,
	}
)

func TestGetItem(t *testing.T) {
	cases := []struct {
		id   int
		want Item
		err  error
	}{
		{
			id:   0,
			want: Item{},
			err:  ErrNotFound,
		},
		{
			id:   999,
			want: Item{},
			err:  ErrNotFound,
		},
		{
			id:   2,
			want: testItem2,
			err:  nil,
		},
		{
			id:   3,
			want: testItem3,
			err:  nil,
		},
	}

	app := AppForTest(t)
	if app == nil {
		t.Fatalf("cannot create AppForTest")
	}

	for _, tc := range cases {
		got, err := app.BidDB.GetItem(tc.id)
		if !errors.Is(err, tc.err) {
			t.Errorf("got err %q want %q for GetItem(%d)",
				err, tc.err, tc.id)
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("got\n%s\nwant\n%s\nfor GetItem(%d)",
				AsJson(got), AsJson(tc.want), tc.id)
		}
	}

	sqlDB := app.BidDB.sqlDB
	app.BidDB.sqlDB = nil
	_, err := app.BidDB.GetItem(0)
	if err != ErrInvalidDB {
		t.Errorf("got err %q want %q", err, ErrInvalidDB)
	}
	app.BidDB.sqlDB = sqlDB
}

func TestGetConfigItem(t *testing.T) {
	cases := []struct {
		cname string
		want  ConfigItem
		err   error
	}{
		{
			cname: "",
			want:  ConfigItem{},
			err:   ErrNotFound,
		},
		{
			cname: "foo",
			want:  ConfigItem{},
			err:   ErrNotFound,
		},
		{
			cname: "cname",
			want:  ConfigItem{Name: "cname", Value: "cvalue", ValueType: "ctype"},
			err:   nil,
		},
	}

	app := AppForTest(t)
	if app == nil {
		t.Fatalf("cannot create AppForTest")
	}

	for _, tc := range cases {
		got, err := app.BidDB.GetConfigItem(tc.cname)
		if !errors.Is(err, tc.err) {
			t.Errorf("got err '%v' want '%v' for GetConfigItem(%q)",
				err, tc.err, tc.cname)
		}
		if got != tc.want {
			t.Errorf("got %+v want %+v for GetConfigItem(%q)",
				got, tc.want, tc.cname)
		}
	}

	sqlDB := app.BidDB.sqlDB
	app.BidDB.sqlDB = nil
	_, err := app.BidDB.GetConfigItem("")
	if err != ErrInvalidDB {
		t.Errorf("got err %q want %q", err, ErrInvalidDB)
	}
	app.BidDB.sqlDB = sqlDB
}

func TestGetItems(t *testing.T) {
	app := AppForTest(t)
	if app == nil {
		t.Fatalf("cannot create AppForTest")
	}

	got, err := app.BidDB.GetItems()
	if err != nil {
		t.Errorf("got err %q want nil", err)
	}

	if !reflect.DeepEqual(got[1], testItem2) {
		t.Errorf("got\n%s\nwant\n%s", AsJson(got[1]), AsJson(testItem2))
	}

	if !reflect.DeepEqual(got[2], testItem3) {
		t.Errorf("got\n%s\nwant\n%s", AsJson(got[2]), AsJson(testItem3))
	}

	sqlDB := app.BidDB.sqlDB
	app.BidDB.sqlDB = nil
	_, err = app.BidDB.GetItems()
	if err != ErrInvalidDB {
		t.Errorf("got err %q want %q", err, ErrInvalidDB)
	}
	app.BidDB.sqlDB = sqlDB
}

func TestGetWinners(t *testing.T) {
	app := AppForTest(t)
	if app == nil {
		t.Fatalf("cannot create AppForTest")
	}

	want := 2
	got, err := app.BidDB.GetWinners()
	if err != nil {
		t.Errorf("got err %q want nil", err)
	}

	if len(got) != want {
		t.Errorf("got %d want %d for len(GetWinners())",
			len(got), want)
	}

	modified := time.Date(2022, time.December, 31, 0, 0, 0, 0, time.UTC)
	winner := Winner{
		ID:         3,
		Title:      "Item Test with Bid",
		Artist:     "Art",
		CurrentBid: 15.0,
		Modified:   modified,
		ModifiedBy: "test",
		Email:      "test@user",
		FullName:   "Test User",
	}

	if !reflect.DeepEqual(got[1], winner) {
		t.Errorf("got\n%s\nwant\n%s\nfor GetWinners()[1]",
			AsJson(got[1]), AsJson(winner))
	}

	sqlDB := app.BidDB.sqlDB
	app.BidDB.sqlDB = nil
	_, err = app.BidDB.GetWinners()
	if err != ErrInvalidDB {
		t.Errorf("got err %q want %q", err, ErrInvalidDB)
	}
	app.BidDB.sqlDB = sqlDB
}

func TestPlaceBidValid(t *testing.T) {
	app := AppForTest(t)
	if app == nil {
		t.Fatalf("cannot create AppForTest")
	}

	item, err := app.BidDB.GetItem(1)
	if err != nil {
		t.Fatalf("GetItem failed: %v", err)
	}

	bidPlaced, msg, _, err := app.BidDB.PlaceBid(item.ID, item.MinBid, "test")
	if err != nil {
		t.Errorf("PlaceBid failed: %v", err)
	}

	if !bidPlaced {
		t.Errorf("got bidPlaced = %v, want %v", bidPlaced, true)
	}

	wantMsg := "Bid placed"
	if msg != wantMsg {
		t.Errorf("got msg = %v, want %v", msg, wantMsg)
	}

	sqlDB := app.BidDB.sqlDB
	app.BidDB.sqlDB = nil
	_, _, _, err = app.BidDB.PlaceBid(item.ID, item.MinBid, "test")
	if err != ErrInvalidDB {
		t.Errorf("got err %q want %q", err, ErrInvalidDB)
	}
	app.BidDB.sqlDB = sqlDB
}

func TestPlaceBidTooLow(t *testing.T) {
	app := AppForTest(t)
	if app == nil {
		t.Fatalf("cannot create AppForTest")
	}

	item, err := app.BidDB.GetItem(1)
	if err != nil {
		t.Fatalf("GetItem failed: %v", err)
	}

	bidPlaced, msg, _, err := app.BidDB.PlaceBid(item.ID, item.MinBid-1, "test")
	if err != nil {
		t.Errorf("PlaceBid failed: %v", err)
	}

	if bidPlaced {
		t.Errorf("got bidPlaced = %v, want %v", bidPlaced, true)
	}

	wantMsg := "Bid too low"
	if msg != wantMsg {
		t.Errorf("got msg = %v, want %v", msg, wantMsg)
	}
}

func TestPlaceBidInvalidItem(t *testing.T) {
	app := AppForTest(t)
	if app == nil {
		t.Fatalf("cannot create AppForTest")
	}

	bidPlaced, msg, _, err := app.BidDB.PlaceBid(0, 0, "test")
	if err != nil {
		t.Errorf("PlaceBid failed: %v", err)
	}

	if bidPlaced {
		t.Errorf("got bidPlaced = %v, want %v", bidPlaced, true)
	}

	wantMsg := "No such item"
	if msg != wantMsg {
		t.Errorf("got msg = %v, want %v", msg, wantMsg)
	}
}

func TestUpdateItem(t *testing.T) {
	app := AppForTest(t)
	if app == nil {
		t.Fatalf("cannot create AppForTest")
	}

	item := Item{
		ID:            1,
		Title:         "Aquarium",
		Description:   "Picture of an Aquarium",
		OpeningBid:    10,
		MinBidIncr:    5,
		Artist:        "Microsoft",
		ImageFileName: "Aquarium.jpg",
	}

	rows, err := app.BidDB.UpdateItem(item)
	if rows > 1 || err != nil {
		t.Errorf("UpdateItem failed, rows = %d, err = %v", rows, err)
	}

	sqlDB := app.BidDB.sqlDB
	app.BidDB.sqlDB = nil
	_, err = app.BidDB.UpdateItem(item)
	if err != ErrInvalidDB {
		t.Errorf("got err %q want %q", err, ErrInvalidDB)
	}
	app.BidDB.sqlDB = sqlDB
}

func TestCreateItem(t *testing.T) {
	app := AppForTest(t)
	if app == nil {
		t.Fatalf("cannot create AppForTest")
	}

	item := Item{
		Title:         "Test CreateItem",
		Description:   "This is a test of CreateItem",
		OpeningBid:    42,
		MinBidIncr:    1,
		Artist:        "CreateItem Artist",
		ImageFileName: "CreateItem.jpg",
	}

	id, rows, err := app.BidDB.CreateItem(item)
	if rows > 1 || err != nil {
		t.Fatalf("CreateItem failed, id = %d, rows = %d, err = %v", id, rows, err)
	}

	got, err := app.BidDB.GetItem(int(id))
	if err != nil {
		t.Fatalf("GetItem(%d) failed: %v", id, err)
	}

	if got.Title != item.Title && got.Description != item.Description && got.OpeningBid != item.OpeningBid && got.MinBidIncr != item.MinBidIncr && got.Artist != item.Artist && got.ImageFileName != item.ImageFileName {
		t.Fatalf("got %v, want %v", got, item)
	}

	sqlDB := app.BidDB.sqlDB
	app.BidDB.sqlDB = nil
	_, err = app.BidDB.UpdateItem(item)
	if err != ErrInvalidDB {
		t.Errorf("got err %q want %q", err, ErrInvalidDB)
	}
	app.BidDB.sqlDB = sqlDB
}

func TestGetBidsForItem(t *testing.T) {
	app := AppForTest(t)
	if app == nil {
		t.Fatalf("cannot create AppForTest")
	}

	_, err := app.BidDB.GetBidsForItem(1)
	if err != nil {
		t.Errorf("GetBidsForItem failed: %v", err)
	}
}

func TestGetBids(t *testing.T) {
	app := AppForTest(t)
	if app == nil {
		t.Fatalf("cannot create AppForTest")
	}

	_, err := app.BidDB.GetBids()
	if err != nil {
		t.Errorf("GetBidsForItem failed: %v", err)
	}
}

func TestGetItemsWithBids(t *testing.T) {
	app := AppForTest(t)
	if app == nil {
		t.Fatalf("cannot create AppForTest")
	}

	_, err := app.BidDB.GetItemsWithBids()
	if err != nil {
		t.Errorf("GetItemsWithBids failed: %v", err)
	}
}
