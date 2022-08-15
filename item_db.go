package main

import (
	"database/sql"
	"errors"
	"log"
	"time"

	weblogin "github.com/bnixon67/go-weblogin"
)

type BidDB struct {
	sqlDB *sql.DB
}

type Item struct {
	ID            int
	Title         string
	Created       time.Time
	Modified      *time.Time
	ModifiedBy    *string
	Description   *string
	OpeningBid    float64
	MinBidIncr    float64
	CurrentBid    float64
	Artist        string
	ImageFileName string

	MinBid float64
}

func (db BidDB) GetItem(id int) (Item, error) {
	var item Item
	var err error

	if db.sqlDB == nil {
		log.Print("db is nil")
		return Item{}, errors.New("invalid db")
	}

	qry := "SELECT id, title, created, modified, modifiedBy, description, openingBid, midBidIncr, currentBid, artist, imageFileName FROM items WHERE id = ?"

	row := db.sqlDB.QueryRow(qry, id)
	err = row.Scan(&item.ID, &item.Title, &item.Created, &item.Modified, &item.ModifiedBy, &item.Description,
		&item.OpeningBid, &item.MinBidIncr, &item.CurrentBid, &item.Artist, &item.ImageFileName)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("item %d does not exist", id)
			return Item{}, errors.New("no such item")
		}
		log.Print("query failed", err)
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

func (db BidDB) GetItems() ([]Item, error) {
	var items []Item
	var err error

	if db.sqlDB == nil {
		log.Print("db is nil")
		return items, errors.New("invalid db")
	}

	qry := "SELECT id, title, created, modified, modifiedBy, description, openingBid, midBidIncr, currentBid, artist, imageFileName FROM items ORDER BY title"

	rows, err := db.sqlDB.Query(qry)
	if err != nil {
		log.Printf("query for items failed, %v", err)
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		var item Item

		err = rows.Scan(&item.ID, &item.Title, &item.Created, &item.Modified, &item.ModifiedBy, &item.Description,
			&item.OpeningBid, &item.MinBidIncr, &item.CurrentBid, &item.Artist, &item.ImageFileName)
		if err != nil {
			log.Printf("row.Scan failed, %v", err)
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
		log.Printf("rows.Err failed, %v", err)
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

	qry := "SELECT id, title, artist, currentBid, fullName, email, modified FROM items LEFT JOIN users ON items.modifiedBy = users.userName WHERE currentBid <> 0 ORDER BY modified"

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

func (db BidDB) PlaceBid(id int, bidAmount float64, user weblogin.User) (string, error) {
	var msg string

	result, err := db.sqlDB.Exec(
		"UPDATE items SET currentBid = ?, modified = current_timestamp(), modifiedBy = ? WHERE id = ? AND IF(CurrentBid=0,OpeningBid,CurrentBid+MidBidIncr) <= ? AND OpeningBid <> 0",
		bidAmount, user.UserName, id, bidAmount)
	if err != nil {
		msg = "Unable to submit bid. Please try again."
		log.Printf("Unable to place bid of %v for item %v by %s", bidAmount, id, user.UserName)
		log.Print(err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		msg = "Unable to submit bid. Please try again."
		log.Printf("Unable to get RowsAffected()")
		log.Print(err)
	}
	if rowsAffected == 1 {
		msg = "Bid placed"
		log.Printf("bid placed on %v by %q for %v", id, user.UserName, bidAmount)
	} else {
		msg = "Your bid is too low or you were outbid by someone else. Please try again."
	}

	return msg, err
}
