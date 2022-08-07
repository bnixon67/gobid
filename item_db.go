package main

import (
	"database/sql"
	"errors"
	"log"
	"time"
)

type BidDB struct {
	sqlDB *sql.DB
}

type Item struct {
	ID            int
	Title         string
	Created       time.Time
	Modified      time.Time
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
