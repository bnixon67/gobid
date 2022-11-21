package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"
)

type BidDB struct {
	sqlDB *sql.DB
}

type BidResult struct {
	BidPlaced   bool
	ID          int
	Message     string
	PriorBidder string
	MinAmount   float64
	NewBidder   string
	NewAmount   float64
}

type Item struct {
	ID            int
	Title         string
	Created       time.Time
	Modified      *time.Time
	ModifiedBy    *string
	Description   string
	OpeningBid    float64
	MinBidIncr    float64
	CurrentBid    float64
	Bidder        string
	Artist        string
	ImageFileName string

	MinBid float64
}

type ConfigItem struct {
	Name      string
	Value     string
	ValueType string
}

func (db BidDB) GetItem(id int) (Item, error) {
	var item Item
	var err error

	if db.sqlDB == nil {
		return Item{}, errors.New("invalid db")
	}

	qry := "SELECT items.id, items.title, items.created, bids.created, bids.bidder, items.description, items.openingBid, items.minBidIncr, IFNULL(bids.amount,0), items.artist, items.imageFileName FROM items LEFT OUTER JOIN current_bids bids ON items.id = bids.id WHERE items.id = ?"

	row := db.sqlDB.QueryRow(qry, id)
	err = row.Scan(&item.ID, &item.Title, &item.Created, &item.Modified, &item.ModifiedBy, &item.Description,
		&item.OpeningBid, &item.MinBidIncr, &item.CurrentBid, &item.Artist, &item.ImageFileName)
	if err != nil {
		if err == sql.ErrNoRows {
			return Item{}, errors.New("no such item")
		}
		return Item{}, errors.New("query failed")
	}

	// TODO: make this a database field
	if item.CurrentBid == 0 {
		item.MinBid = item.OpeningBid
	} else {
		item.MinBid = item.CurrentBid + item.MinBidIncr
	}

	return item, nil
}

func (db BidDB) GetConfigItem(name string) (ConfigItem, error) {
	var config ConfigItem
	var err error

	if db.sqlDB == nil {
		return ConfigItem{}, errors.New("invalid db")
	}

	qry := "SELECT name, value, value_type FROM config WHERE name = ?"

	row := db.sqlDB.QueryRow(qry, name)
	err = row.Scan(&config.Name, &config.Value, &config.ValueType)
	if err != nil {
		if err == sql.ErrNoRows {
			return ConfigItem{}, fmt.Errorf("no such config: %s", name)
		}
		return ConfigItem{}, errors.New("query failed")
	}

	return config, nil
}

func (db BidDB) GetItems() ([]Item, error) {
	var items []Item
	var err error

	if db.sqlDB == nil {
		return items, errors.New("invalid db")
	}

	qry := "SELECT items.id, items.title, items.created, bids.created, bids.bidder, items.description, items.openingBid, items.minBidIncr, IFNULL(bids.amount,0), IFNULL(bids.bidder,'<none>'), items.artist, items.imageFileName FROM items LEFT OUTER JOIN current_bids bids ON items.id = bids.id"

	rows, err := db.sqlDB.Query(qry)
	if err != nil {
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		var item Item

		err = rows.Scan(&item.ID, &item.Title, &item.Created, &item.Modified, &item.ModifiedBy, &item.Description, &item.OpeningBid, &item.MinBidIncr, &item.CurrentBid, &item.Bidder, &item.Artist, &item.ImageFileName)
		if err != nil {
			return items, err
		} else {
			// TODO: make this a database field
			if item.CurrentBid == 0 {
				item.MinBid = item.OpeningBid
			} else {
				item.MinBid = item.CurrentBid + item.MinBidIncr
			}

			items = append(items, item)
		}
	}
	err = rows.Err()
	if err != nil {
	}

	return items, err
}

func (db BidDB) GetWinners() ([]Winner, error) {
	var winners []Winner
	var err error

	if db.sqlDB == nil {
		log.Print("db is nil")
		return winners, errors.New("invalid db")
	}

	qry := "SELECT items.id, title, artist, amount, fullName, email, bids.created FROM items LEFT OUTER JOIN current_bids bids ON items.id = bids.id LEFT JOIN users ON bids.bidder = users.userName WHERE bids.Amount <> 0 ORDER BY items.id"

	rows, err := db.sqlDB.Query(qry)
	if err != nil {
		log.Printf("query for winners failed, %v", err)
		return winners, err
	}
	defer rows.Close()

	for rows.Next() {
		var winner Winner

		err = rows.Scan(&winner.ID, &winner.Title, &winner.Artist, &winner.CurrentBid, &winner.FullName, &winner.Email, &winner.Modified)
		if err != nil {
			log.Printf("row.Scan failed, %v", err)
		}

		winners = append(winners, winner)
	}
	err = rows.Err()
	if err != nil {
		log.Printf("rows.Err failed, %v", err)
	}

	return winners, err
}

func (db BidDB) PlaceBid(id int, bidAmount float64, userName string) (bool, string, string, error) {
	var msg string
	var br BidResult

	row := db.sqlDB.QueryRow("CALL placeBid(?, ?, ?)", id, bidAmount, userName)
	err := row.Scan(&br.BidPlaced, &br.ID, &br.Message, &br.PriorBidder, &br.MinAmount,
		&br.NewBidder, &br.NewAmount)
	if err != nil {
		msg = "Unable to submit bid. Please try again."
		log.Printf("Unable to place bid of %v for item %v by %s",
			bidAmount, id, userName)
		log.Print(err)
	} else {
		msg = br.Message
		log.Printf("%s for item %v for amount %v by %s",
			msg, id, bidAmount, userName)
		log.Printf("%s was outbid", br.PriorBidder)
	}

	return br.BidPlaced, msg, br.PriorBidder, err
}

func (db BidDB) UpdateItem(item Item) (int64, error) {
	update := "UPDATE items SET title = ?, description = ?, openingBid = ?, minBidIncr = ?, artist = ?, imageFileName = ? WHERE id = ?"
	result, err := db.sqlDB.Exec(update, item.Title, item.Description, item.OpeningBid, item.MinBidIncr, item.Artist, item.ImageFileName, item.ID)
	if err != nil {
		log.Printf("UpdateItem failed for %d: %v", item.ID, err)
		return 0, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		log.Printf("result.RowsAffected failed: %v", err)
		return 0, err
	}

	return rows, err
}

func (db BidDB) CreateItem(item Item) (int64, int64, error) {
	insert := "INSERT INTO items(title, description, openingBid, minBidIncr, artist, imageFileName) VALUES (?, ?, ?, ?, ?, ?)"
	result, err := db.sqlDB.Exec(insert, item.Title, item.Description, item.OpeningBid, item.MinBidIncr, item.Artist, item.ImageFileName)
	if err != nil {
		log.Printf("CreateItem failed for %d: %v", item.ID, err)
		return 0, 0, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		log.Printf("result.RowsAffected failed: %v", err)
		return 0, 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("result.RowsAffected failed: %v", err)
		return 0, 0, err
	}

	return id, rows, err
}
