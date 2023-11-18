// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package main

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/bnixon67/webapp/weblogin"
)

type BidDB struct {
	sqlDB *weblogin.LoginDB
}

type BidResult struct {
	BidPlaced   bool
	Message     string
	PriorBidder string
}

type Item struct {
	ID            int
	Title         string
	Created       time.Time
	Description   string
	OpeningBid    float64
	MinBidIncr    float64
	Artist        string
	ImageFileName string
	Bidder        string
	CurrentBid    float64
	Modified      *time.Time
	MinBid        float64
}

type ItemWithBids struct {
	ID            int
	Title         string
	Created       time.Time
	Description   string
	OpeningBid    float64
	MinBidIncr    float64
	Artist        string
	ImageFileName string
	Bids          []Bid
}

type ConfigItem struct {
	Name      string
	Value     string
	ValueType string
}

type Bid struct {
	ID       int
	Created  time.Time
	Bidder   string
	Amount   float64
	FullName string
	Email    string
}

var (
	ErrNotFound  = errors.New("not found")
	ErrInvalidDB = errors.New("invalid db")
)

func (db BidDB) GetItem(id int) (Item, error) {
	var item Item
	var err error

	if db.sqlDB == nil {
		return item, ErrInvalidDB
	}

	qry := "SELECT items.id, items.title, items.created, bids.created, IFNULL(bids.bidder,''), items.description, items.openingBid, items.minBidIncr, IFNULL(bids.amount,0), items.artist, items.imageFileName FROM items LEFT OUTER JOIN current_bids bids ON items.id = bids.id WHERE items.id = ?"

	row := db.sqlDB.QueryRow(qry, id)
	err = row.Scan(&item.ID, &item.Title, &item.Created, &item.Modified, &item.Bidder, &item.Description, &item.OpeningBid, &item.MinBidIncr, &item.CurrentBid, &item.Artist, &item.ImageFileName)
	if err != nil {
		if err == sql.ErrNoRows {
			return item, fmt.Errorf("item %d: %w", id, ErrNotFound)
		}
		return item, err
	}

	// TODO: make this a database field
	if item.CurrentBid == 0 {
		item.MinBid = item.OpeningBid
	} else {
		item.MinBid = item.CurrentBid + item.MinBidIncr
	}

	return item, err
}

func (db BidDB) GetConfigItem(name string) (ConfigItem, error) {
	var config ConfigItem
	var err error

	if db.sqlDB == nil {
		return config, ErrInvalidDB
	}

	qry := "SELECT name, value, value_type FROM config WHERE name = ?"

	row := db.sqlDB.QueryRow(qry, name)
	err = row.Scan(&config.Name, &config.Value, &config.ValueType)
	if err != nil {
		if err == sql.ErrNoRows {
			return ConfigItem{}, fmt.Errorf("name %q: %w",
				name, ErrNotFound)
		}
		return config, err
	}

	return config, err
}

func (db BidDB) GetItems() ([]Item, error) {
	var items []Item
	var err error

	if db.sqlDB == nil {
		return items, ErrInvalidDB
	}

	qry := "SELECT items.id, items.title, items.created, bids.created, items.description, items.openingBid, items.minBidIncr, IFNULL(bids.amount,0), IFNULL(bids.bidder,''), items.artist, items.imageFileName FROM items LEFT OUTER JOIN current_bids bids ON items.id = bids.id"

	rows, err := db.sqlDB.Query(qry)
	if err != nil {
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		var item Item

		err = rows.Scan(&item.ID, &item.Title, &item.Created, &item.Modified, &item.Description, &item.OpeningBid, &item.MinBidIncr, &item.CurrentBid, &item.Bidder, &item.Artist, &item.ImageFileName)
		if err != nil {
			return items, err
		}

		// TODO: make this a database field
		if item.CurrentBid == 0 {
			item.MinBid = item.OpeningBid
		} else {
			item.MinBid = item.CurrentBid + item.MinBidIncr
		}

		items = append(items, item)
	}
	err = rows.Err()
	if err != nil {
		return items, err
	}

	return items, err
}

func (db BidDB) GetWinners() ([]Winner, error) {
	var winners []Winner
	var err error

	if db.sqlDB == nil {
		return winners, ErrInvalidDB
	}

	qry := "SELECT items.id, items.title, items.artist, bids.amount, bids.created, bids.bidder, IFNULL(users.fullName,'<missing>'), IFNULL(users.email,'<missing>') FROM items LEFT OUTER JOIN current_bids bids ON items.id = bids.id LEFT JOIN users ON bids.bidder = users.userName WHERE bids.Amount <> 0 ORDER BY items.id"

	rows, err := db.sqlDB.Query(qry)
	if err != nil {
		return winners, err
	}
	defer rows.Close()

	for rows.Next() {
		var winner Winner

		err = rows.Scan(&winner.ID, &winner.Title, &winner.Artist, &winner.CurrentBid, &winner.Modified, &winner.ModifiedBy, &winner.FullName, &winner.Email)
		if err != nil {
			return winners, err
		}

		winners = append(winners, winner)
	}
	err = rows.Err()
	if err != nil {
		return winners, err
	}

	return winners, err
}

const PlaceBidError = "Unable to place bid. Try again."

func (db BidDB) PlaceBid(id int, bidAmount float64, userName string) (BidResult, error) {
	var bidResult BidResult

	if db.sqlDB == nil {
		return bidResult, ErrInvalidDB
	}

	row := db.sqlDB.QueryRow("CALL placeBid(?, ?, ?)", id, bidAmount, userName)
	err := row.Scan(&bidResult.BidPlaced, &bidResult.Message, &bidResult.PriorBidder)
	if err != nil {
		bidResult.Message = PlaceBidError
		return bidResult, err
	}

	return bidResult, err
}

var ErrInvalidItem = errors.New("invalid item")

func (db BidDB) UpdateItem(item Item) (int64, error) {
	if db.sqlDB == nil {
		return 0, ErrInvalidDB
	}

	// require non-empty strings for some fields
	if AnyEmpty(
		item.Title,
		item.Description,
		item.Artist,
		item.ImageFileName,
	) {
		return 0, ErrInvalidItem
	}

	update := "UPDATE items SET title = ?, description = ?, openingBid = ?, minBidIncr = ?, artist = ?, imageFileName = ? WHERE id = ?"
	result, err := db.sqlDB.Exec(update, item.Title, item.Description, item.OpeningBid, item.MinBidIncr, item.Artist, item.ImageFileName, item.ID)
	if err != nil {
		return 0, err
	}

	cnt, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return cnt, err
}

var ErrCreateFailed = errors.New("create failed")

func (db BidDB) CreateItem(item Item) (int64, error) {
	if db.sqlDB == nil {
		return 0, ErrInvalidDB
	}

	// require non-empty strings for some fields
	if AnyEmpty(
		item.Title,
		item.Description,
		item.Artist,
	) {
		return 0, ErrInvalidItem
	}

	insert := "INSERT INTO items(title, description, openingBid, minBidIncr, artist, imageFileName) VALUES (?, ?, ?, ?, ?, ?)"
	result, err := db.sqlDB.Exec(insert, item.Title, item.Description, item.OpeningBid, item.MinBidIncr, item.Artist, item.ImageFileName)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrCreateFailed, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrCreateFailed, err)
	}
	if rows != 1 {
		return 0, fmt.Errorf("%w: %v", ErrCreateFailed, "multiple rows affected")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrCreateFailed, err)
	}

	return id, err
}

func (db BidDB) GetBidsForItem(id int) ([]Bid, error) {
	var bids []Bid
	var err error

	if db.sqlDB == nil {
		return bids, ErrInvalidDB
	}

	qry := "SELECT id, created, bidder, amount FROM bids WHERE id = ? ORDER BY created DESC"

	rows, err := db.sqlDB.Query(qry, id)
	if err != nil {
		return bids, err
	}
	defer rows.Close()

	for rows.Next() {
		var bid Bid

		err = rows.Scan(&bid.ID, &bid.Created, &bid.Bidder, &bid.Amount)
		if err != nil {
			return bids, err
		}

		bids = append(bids, bid)
	}
	err = rows.Err()
	if err != nil {
		return bids, err
	}

	return bids, err
}

func (db BidDB) GetBids() ([]Bid, error) {
	var bids []Bid
	var err error

	if db.sqlDB == nil {
		return bids, ErrInvalidDB
	}

	qry := "SELECT b.id, b.created, b.bidder, b.amount, u.fullName, u.email FROM bids b INNER JOIN users u ON b.bidder = u.userName ORDER BY b.id, b.created DESC"

	rows, err := db.sqlDB.Query(qry)
	if err != nil {
		return bids, err
	}
	defer rows.Close()

	for rows.Next() {
		var bid Bid

		err = rows.Scan(&bid.ID, &bid.Created, &bid.Bidder, &bid.Amount, &bid.FullName, &bid.Email)
		if err != nil {
			return bids, err
		}

		bids = append(bids, bid)
	}
	err = rows.Err()
	if err != nil {
		return bids, err
	}

	return bids, err
}

type ItemsWithBidsResult struct {
	ID            int
	Title         string
	ItemCreated   time.Time
	Description   string
	OpeningBid    float64
	MinBidIncr    float64
	Artist        string
	ImageFileName string
	BidCreated    time.Time
	Bidder        string
	Amount        float64
	FullName      string
	Email         string
}

func (db BidDB) GetItemsWithBids() ([]ItemWithBids, error) {
	var items []ItemWithBids
	var err error

	if db.sqlDB == nil {
		return items, ErrInvalidDB
	}

	qry := "SELECT items.id, items.title, items.created AS itemCreated, items.description, items.openingBid, items.minBidIncr, items.artist, items.imageFileName, bids.created AS bidCreated, bids.bidder, bids.amount, users.fullName, users.email FROM items INNER JOIN bids ON items.id = bids.id INNER JOIN users ON bids.bidder = users.UserName ORDER BY items.id, bids.created DESC"

	rows, err := db.sqlDB.Query(qry)
	if err != nil {
		return items, err
	}
	defer rows.Close()

	var priorID int

	for rows.Next() {
		var item ItemWithBids
		var bid Bid

		err = rows.Scan(&item.ID, &item.Title, &item.Created, &item.Description, &item.OpeningBid, &item.MinBidIncr, &item.Artist, &item.ImageFileName, &bid.Created, &bid.Bidder, &bid.Amount, &bid.FullName, &bid.Email)
		if err != nil {
			return items, err
		}

		bid.ID = item.ID

		if item.ID != priorID {
			items = append(items, item)
		}

		n := len(items) - 1
		items[n].Bids = append(items[n].Bids, bid)

		priorID = item.ID
	}
	err = rows.Err()
	if err != nil {
		return items, err
	}

	return items, err
}
