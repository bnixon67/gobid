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
	"strings"
	"testing"
	"time"
)

var (
	ct = time.Date(2022, time.December, 30, 0, 0, 0, 0, time.UTC)
	mt = time.Date(2022, time.December, 31, 0, 0, 0, 0, time.UTC)
	bt = time.Date(2022, time.December, 31, 0, 0, 0, 0, time.UTC)
	mb = "test"

	testID2 = Item{
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

	testID3 = Item{
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
		{id: 0, want: Item{}, err: ErrNotFound},
		{id: 999, want: Item{}, err: ErrNotFound},
		{id: 2, want: testID2, err: nil},
		{id: 3, want: testID3, err: nil},
	}

	app := AppForTest(t)
	if app == nil {
		t.Fatalf("cannot create AppForTest")
	}

	for _, tc := range cases {
		got, err := app.BidDB.GetItem(tc.id)
		if !errors.Is(err, tc.err) {
			t.Errorf("GetItem(%d)\ngot err '%v' want '%v'",
				tc.id, err, tc.err)
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("GetItem(%d)\n got %s\nwant %s",
				tc.id, AsJson(got), AsJson(tc.want))
		}
	}

	// test for invalid DB
	sqlDB := app.BidDB.sqlDB
	app.BidDB.sqlDB = nil
	_, err := app.BidDB.GetItem(0)
	if err != ErrInvalidDB {
		t.Errorf("got err '%v' want '%v'", err, ErrInvalidDB)
	}
	app.BidDB.sqlDB = sqlDB
}

func TestGetConfigItem(t *testing.T) {
	testConfigItem := ConfigItem{
		Name: "cname", Value: "cvalue", ValueType: "ctype",
	}

	cases := []struct {
		cname string
		want  ConfigItem
		err   error
	}{
		{cname: "", want: ConfigItem{}, err: ErrNotFound},
		{cname: "nosuchitem", want: ConfigItem{}, err: ErrNotFound},
		{cname: "cname", want: testConfigItem, err: nil},
	}

	app := AppForTest(t)
	if app == nil {
		t.Fatalf("cannot create AppForTest")
	}

	for _, tc := range cases {
		got, err := app.BidDB.GetConfigItem(tc.cname)
		if !errors.Is(err, tc.err) {
			t.Errorf("GetConfigItem(%q)\ngot err '%v' want '%v'",
				tc.cname, err, tc.err)
		}
		if got != tc.want {
			t.Errorf("GetConfigItem(%q)\n got %+v\nwant %+v",
				tc.cname, got, tc.want)
		}
	}

	// test for invalid DB
	sqlDB := app.BidDB.sqlDB
	app.BidDB.sqlDB = nil
	_, err := app.BidDB.GetConfigItem("")
	if err != ErrInvalidDB {
		t.Errorf("got err '%v' want '%v'", err, ErrInvalidDB)
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
		t.Fatalf("got err '%v' want '%v'", err, nil)
	}

	// test for a few results in Items
	cases := []struct {
		idx  int
		want Item
	}{
		{idx: 1, want: testID2}, // arrary is 0 based, so idx-1 = id
		{idx: 2, want: testID3},
	}
	for _, tc := range cases {
		if !reflect.DeepEqual(got[tc.idx], tc.want) {
			t.Errorf("Item[%d]:\n got %s\nwant %s",
				tc.idx, AsJson(got[tc.idx]), AsJson(tc.want))
		}
	}

	// test for invalid DB
	sqlDB := app.BidDB.sqlDB
	app.BidDB.sqlDB = nil
	_, err = app.BidDB.GetItems()
	if err != ErrInvalidDB {
		t.Errorf("got err '%v' want '%v'", err, ErrInvalidDB)
	}
	app.BidDB.sqlDB = sqlDB
}

func TestGetWinners(t *testing.T) {
	app := AppForTest(t)
	if app == nil {
		t.Fatalf("cannot create AppForTest")
	}

	got, err := app.BidDB.GetWinners()
	if err != nil {
		t.Errorf("got err '%v' want '%v'", err, nil)
	}

	// test to see if there is a specific winner in the results
	modified := time.Date(2022, time.December, 31, 0, 0, 0, 0, time.UTC)
	tWinner := Winner{
		ID:         3,
		Title:      "Item Test with Bid",
		Artist:     "Art",
		CurrentBid: 15.0,
		Modified:   modified,
		ModifiedBy: "test",
		Email:      "test@user",
		FullName:   "Test User",
	}
	found := false
	for idx := range got {
		if got[idx].ID == tWinner.ID {
			found = true
			if !reflect.DeepEqual(got[idx], tWinner) {
				t.Errorf("GetWinners[%d]:\n got %s\nwant %s\n",
					idx, AsJson(got[idx]), AsJson(tWinner))
			}
		}
	}
	if !found {
		t.Errorf("did not find Winner with ID=%d", tWinner.ID)
	}

	// test for invalid DB
	sqlDB := app.BidDB.sqlDB
	app.BidDB.sqlDB = nil
	_, err = app.BidDB.GetWinners()
	if err != ErrInvalidDB {
		t.Errorf("got err '%v' want '%v'", err, ErrInvalidDB)
	}
	app.BidDB.sqlDB = sqlDB
}

func TestPlaceBid(t *testing.T) {
	app := AppForTest(t)
	if app == nil {
		t.Fatalf("cannot create AppForTest")
	}

	// get item to test
	id1, err := app.BidDB.GetItem(1)
	if err != nil {
		t.Fatalf("GetItem(1) failed: %v", err)
	}

	cases := []struct {
		id        int
		bidAmount float64
		bidder    string
		want      BidResult
		err       error
		sleep     bool
	}{
		{
			id: 0, bidAmount: 1.0, bidder: "test",
			want: BidResult{
				BidPlaced: false,
				Message:   "No such item",
			},
			err: nil,
		},
		{
			id: 4, bidAmount: 1.0, bidder: "test",
			want: BidResult{
				BidPlaced: false,
				Message:   "Display only item",
			},
			err: nil,
		},
		{
			id: 3, bidAmount: 1.0, bidder: "test",
			want: BidResult{
				BidPlaced:   false,
				Message:     "Bid too low",
				PriorBidder: "test",
			},
			err: nil,
		},
		{
			id: 1, bidAmount: id1.MinBid, bidder: "test",
			want: BidResult{
				BidPlaced:   true,
				Message:     "Bid placed",
				PriorBidder: "",
			},
			err: nil,
		},
		{
			id: 1, bidAmount: 100, bidder: "test",
			want: BidResult{
				BidPlaced:   true,
				Message:     "Bid placed",
				PriorBidder: "test",
			},
			err:   nil,
			sleep: true,
		},
	}

	for _, tc := range cases {
		if tc.sleep {
			time.Sleep(time.Second)
		}
		got, err := app.BidDB.PlaceBid(tc.id, tc.bidAmount, tc.bidder)
		if !errors.Is(err, tc.err) {
			t.Errorf("PlaceBid(%d, %f, %q)\ngot err '%v' want '%v'",
				tc.id, tc.bidAmount, tc.bidder, err, tc.err)
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("PlaceBid(%d, %f, %q)\n got %s\nwant %s",
				tc.id, tc.bidAmount, tc.bidder,
				AsJson(got), AsJson(tc.want))
		}
	}

	// test for invalid DB
	sqlDB := app.BidDB.sqlDB
	app.BidDB.sqlDB = nil
	_, err = app.BidDB.PlaceBid(0, 0, "test")
	if err != ErrInvalidDB {
		t.Errorf("got err '%v' want '%v'", err, ErrInvalidDB)
	}
	app.BidDB.sqlDB = sqlDB
}

func TestUpdateItem(t *testing.T) {
	app := AppForTest(t)
	if app == nil {
		t.Fatalf("cannot create AppForTest")
	}

	// get item to test
	testItem, err := app.BidDB.GetItem(5)
	if err != nil {
		t.Fatalf("GetItem(1) failed: %v", err)
	}
	testItem.MinBidIncr += 1

	cases := []struct {
		item Item
		want int64
		err  error
	}{
		{
			item: testItem,
			want: 1, err: nil,
		},
		{
			item: Item{},
			want: 0, err: ErrInvalidItem,
		},
		{
			item: Item{ID: 5},
			want: 0, err: ErrInvalidItem,
		},
		{
			item: Item{ID: 5, Title: "t"},
			want: 0, err: ErrInvalidItem,
		},
		{
			item: Item{ID: 5, Title: "t", Description: "d"},
			want: 0, err: ErrInvalidItem,
		},
		{
			item: Item{ID: 5, Title: "t", Description: "d", Artist: "a"},
			want: 0, err: ErrInvalidItem,
		},
		{
			item: Item{ID: 5, Title: "t", Description: "d", Artist: "a", ImageFileName: "i", Created: ct.Add(time.Hour * 5)},
			want: 1, err: nil,
		},
	}

	for _, tc := range cases {
		got, err := app.BidDB.UpdateItem(tc.item)
		if !errors.Is(err, tc.err) {
			t.Errorf("UpdateItem(%+v)\ngot err '%v' want '%v'",
				tc.item, err, tc.err)
		}
		if got != tc.want {
			t.Errorf("UpdateItem(%+v)\ngot %d want %d",
				tc.item, got, tc.want)
		}
		if got == 1 {
			item, err := app.BidDB.GetItem(tc.item.ID)
			if err != nil {
				t.Fatalf("GetItem(%d) failed: %v",
					tc.item.ID, err)
			}
			if !reflect.DeepEqual(item, tc.item) {
				t.Errorf("GetItem(%d)\n got %s\nwant %s",
					tc.item.ID,
					AsJson(item), AsJson(tc.item))
			}

		}
	}

	// test for invalid DB
	sqlDB := app.BidDB.sqlDB
	app.BidDB.sqlDB = nil
	_, err = app.BidDB.UpdateItem(Item{})
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

	testItem := Item{
		Title:         "Test CreateItem",
		Description:   "This is a test of CreateItem",
		OpeningBid:    42,
		MinBidIncr:    1,
		Artist:        "CreateItem Artist",
		ImageFileName: "CreateItem.jpg",
	}
	errorItem := testItem
	errorItem.Title = strings.Repeat("x", 100)

	cases := []struct {
		item Item
		err  error
	}{
		{item: testItem, err: nil},
		{item: Item{}, err: ErrInvalidItem},
		{item: errorItem, err: ErrCreateFailed},
	}

	for _, tc := range cases {
		newID, err := app.BidDB.CreateItem(tc.item)
		if !errors.Is(err, tc.err) {
			t.Errorf("UpdateItem(%+v)\ngot err '%v' want '%v'",
				tc.item, err, tc.err)
		}
		if err == nil {
			item, err := app.BidDB.GetItem(int(newID))
			if err != nil {
				t.Fatalf("GetItem(%d) failed: %v",
					newID, err)
			}
			tc.item.ID = int(newID)
			tc.item.Created = item.Created
			tc.item.MinBid = item.MinBid
			if !reflect.DeepEqual(item, tc.item) {
				t.Errorf("GetItem(%d)\n got %s\nwant %s",
					newID,
					AsJson(item), AsJson(tc.item))
			}
		}
	}

	// test for invalid DB
	sqlDB := app.BidDB.sqlDB
	app.BidDB.sqlDB = nil
	_, err := app.BidDB.CreateItem(Item{})
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

	var noBids []Bid

	testBidsID3 := []Bid{
		Bid{ID: 3, Created: bt, Bidder: "test", Amount: 15, FullName: "", Email: ""},
	}

	testBidsID6 := []Bid{
		Bid{ID: 6, Created: bt.Add(time.Hour * 3), Bidder: "test", Amount: 7, FullName: "", Email: ""},
		Bid{ID: 6, Created: bt.Add(time.Hour * 2), Bidder: "test", Amount: 5, FullName: "", Email: ""},
		Bid{ID: 6, Created: bt.Add(time.Hour * 1), Bidder: "test", Amount: 3, FullName: "", Email: ""},
	}

	cases := []struct {
		id   int
		want []Bid
		err  error
	}{
		{id: 0, want: noBids, err: nil},
		{id: 2, want: noBids, err: nil},
		{id: 4, want: noBids, err: nil},
		{id: 3, want: testBidsID3, err: nil},
		{id: 6, want: testBidsID6, err: nil},
	}

	for _, tc := range cases {
		got, err := app.BidDB.GetBidsForItem(tc.id)
		if !errors.Is(err, tc.err) {
			t.Errorf("GetBidsForItem(%d)\ngot err '%v' want '%v'",
				tc.id, err, tc.err)
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("GetBidsForItem(%d)\n got %s\nwant %s",
				tc.id,
				AsJson(got), AsJson(tc.want))
		}
	}

	// test for invalid DB
	sqlDB := app.BidDB.sqlDB
	app.BidDB.sqlDB = nil
	_, err := app.BidDB.CreateItem(Item{})
	if err != ErrInvalidDB {
		t.Errorf("got err %q want %q", err, ErrInvalidDB)
	}
	app.BidDB.sqlDB = sqlDB
}

func containsBid(bids []Bid, bid Bid) bool {
	for _, b := range bids {
		if reflect.DeepEqual(bid, b) {
			return true
		}
	}
	return false
}

func TestGetBids(t *testing.T) {
	app := AppForTest(t)
	if app == nil {
		t.Fatalf("cannot create AppForTest")
	}

	got, err := app.BidDB.GetBids()
	if err != nil {
		t.Errorf("GetBids failed: %v", err)
	}

	want := []Bid{
		Bid{ID: 3, Created: bt, Bidder: "test", Amount: 15, FullName: "Test User", Email: "test@user"},
		Bid{ID: 6, Created: bt.Add(time.Hour * 3), Bidder: "test", Amount: 7, FullName: "Test User", Email: "test@user"},
		Bid{ID: 6, Created: bt.Add(time.Hour * 2), Bidder: "test", Amount: 5, FullName: "Test User", Email: "test@user"},
		Bid{ID: 6, Created: bt.Add(time.Hour * 1), Bidder: "test", Amount: 3, FullName: "Test User", Email: "test@user"},
	}

	// check if want elements are in got
	for _, w := range want {
		if !containsBid(got, w) {
			t.Errorf("Did not find bid\n%s\nin\n%s",
				AsJson(w), AsJson(got))
		}
	}

	// test for invalid DB
	sqlDB := app.BidDB.sqlDB
	app.BidDB.sqlDB = nil
	_, err = app.BidDB.GetBids()
	if err != ErrInvalidDB {
		t.Errorf("got err %q want %q", err, ErrInvalidDB)
	}
	app.BidDB.sqlDB = sqlDB
}

func containsItemWithBids(items []ItemWithBids, item ItemWithBids) bool {
	for _, i := range items {
		if reflect.DeepEqual(item, i) {
			return true
		}
	}
	return false
}

func TestGetItemsWithBids(t *testing.T) {
	app := AppForTest(t)
	if app == nil {
		t.Fatalf("cannot create AppForTest")
	}

	got, err := app.BidDB.GetItemsWithBids()
	if err != nil {
		t.Errorf("GetBids failed: %v", err)
	}

	bidsID3 := []Bid{
		Bid{
			ID: 3, Created: bt, Bidder: "test", Amount: 15,
			FullName: "Test User", Email: "test@user",
		},
	}
	bidsID6 := []Bid{
		Bid{ID: 6, Created: bt.Add(time.Hour * 3), Bidder: "test", Amount: 7, FullName: "Test User", Email: "test@user"},
		Bid{ID: 6, Created: bt.Add(time.Hour * 2), Bidder: "test", Amount: 5, FullName: "Test User", Email: "test@user"},
		Bid{ID: 6, Created: bt.Add(time.Hour * 1), Bidder: "test", Amount: 3, FullName: "Test User", Email: "test@user"},
	}

	want := []ItemWithBids{
		ItemWithBids{
			ID: 3, Title: "Item Test with Bid",
			Created:     ct.Add(time.Hour * 3),
			Description: "Item to test GetItem with Bid",
			OpeningBid:  5, MinBidIncr: 1,
			Artist: "Art", ImageFileName: "File",
			Bids: bidsID3,
		},
		ItemWithBids{
			ID: 6, Title: "Item Test with 3 Bids",
			Created:     ct.Add(time.Hour * 6),
			Description: "Item to test GetItem with 3 Bids",
			OpeningBid:  3, MinBidIncr: 2,
			Artist: "Art 3 Bid", ImageFileName: "File 3 Bid",
			Bids: bidsID6,
		},
	}

	// check if want elements are in got
	for _, w := range want {
		if !containsItemWithBids(got, w) {
			t.Errorf("Did not find bid\n%s\nin\n%s",
				AsJson(w), AsJson(got))
		}
	}

	// test for invalid DB
	sqlDB := app.BidDB.sqlDB
	app.BidDB.sqlDB = nil
	_, err = app.BidDB.GetBids()
	if err != ErrInvalidDB {
		t.Errorf("got err %q want %q", err, ErrInvalidDB)
	}
	app.BidDB.sqlDB = sqlDB
}

/*
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
*/
