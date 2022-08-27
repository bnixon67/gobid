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
	"testing"
)

func TestPlaceBidValid(t *testing.T) {
	app := AppForTest(t)
	if app == nil {
		t.Fatalf("cannot create AppForTest")
	}

	item, err := app.BidDB.GetItem(1)
	if err != nil {
		t.Fatalf("GetItem failed: %v", err)
	}

	bidPlaced, msg, _, err := app.BidDB.PlaceBid(item.ID,item.MinBid, "test")
	if err != nil {
		t.Errorf("PlaceBid failed: %v", err)
	}

	if !bidPlaced {
		t.Errorf("Expected bidPlaced = %v, got %v", true, bidPlaced)
	}

	expectedMsg := "Bid placed"
	if msg != expectedMsg {
		t.Errorf("Expected msg = %v, got %v", expectedMsg, msg)
	}
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

	bidPlaced, msg, _, err := app.BidDB.PlaceBid(item.ID,item.MinBid-1, "test")
	if err != nil {
		t.Errorf("PlaceBid failed: %v", err)
	}

	if bidPlaced {
		t.Errorf("Expected bidPlaced = %v, got %v", true, bidPlaced)
	}

	expectedMsg := "Bid too low"
	if msg != expectedMsg {
		t.Errorf("Expected msg = %v, got %v", expectedMsg, msg)
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
		t.Errorf("Expected bidPlaced = %v, got %v", true, bidPlaced)
	}

	expectedMsg := "No such item"
	if msg != expectedMsg {
		t.Errorf("Expected msg = %v, got %v", expectedMsg, msg)
	}
}
