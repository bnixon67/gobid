TRUNCATE TABLE items;

INSERT INTO items(id, title, created, description, openingBid, minBidIncr, artist, imageFileName)
VALUES
(1,"Bid Test","2022-12-30 01:00","Item to test PlaceBid",10,2,"ARTIST","FILENAME"),
(2,"Item Test","2022-12-30 02:00","Item to test GetItem",10,2,"ARTIST","FILENAME"),
(3,"Item Test with Bid","2022-12-30 03:00","Item to test GetItem with Bid",5,1,"Art","File"),
(4,"Item Test Display Only","2022-12-30 04:00","Item to test Display Only",0,0,"Art4","File4"),
(5,"UpdateItem Test","2022-12-30 05:00","Item to test UpdateItem",1,1,"Art5","File5");

TRUNCATE TABLE config;

INSERT INTO config(name, value, value_type)
VALUES("cname", "cvalue", "ctype");

TRUNCATE TABLE bids;

INSERT INTO bids(id, created, bidder, amount)
VALUES
(3,"2022-12-31","test",15);

TRUNCATE TABLE users;

INSERT INTO users(userName, fullName, email, hashedPassword)
VALUES ("test", "Test User", "test@user", "password");
