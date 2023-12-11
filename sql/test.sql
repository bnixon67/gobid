TRUNCATE TABLE items;

INSERT INTO items(id, title, created, description, openingBid, minBidIncr, artist, imageFileName)
VALUES
(1,"Bid Test","2022-12-30 01:00","Item to test PlaceBid",10,2,"ARTIST","FILENAME"),
(2,"Item Test","2022-12-30 02:00","Item to test GetItem",10,2,"ARTIST","FILENAME"),
(3,"Item Test with Bid","2022-12-30 03:00","Item to test GetItem with Bid",5,1,"Art","File"),
(4,"Item Test Display Only","2022-12-30 04:00","Item to test Display Only",0,0,"Art4","File4"),
(5,"UpdateItem Test","2022-12-30 05:00","Item to test UpdateItem",1,1,"Art5","File5"),
(6,"Item Test with 3 Bids","2022-12-30 06:00","Item to test GetItem with 3 Bids",3,2,"Art 3 Bid","File 3 Bid");

TRUNCATE TABLE config;

INSERT INTO config(name, value, value_type)
VALUES("cname", "cvalue", "ctype");

TRUNCATE TABLE bids;

INSERT INTO bids(id, created, bidder, amount)
VALUES
(3,"2022-12-31","test",15),
(6,"2022-12-31 01:00","test",3),
(6,"2022-12-31 02:00","test",5),
(6,"2022-12-31 03:00","test",7);

TRUNCATE TABLE users;

INSERT INTO users(userName, fullName, email, hashedPassword)
VALUES ("test", "Test User", "test@user", "$2a$10$2bLycFqUmc6m6iLkaeUgKOGwzekGd9IoAPMbXRNNuJ8Sv9ItgV29O");

INSERT INTO users(userName, fullName, email, hashedPassword, admin)
VALUES ("admin", "Admin User", "admin@user", "$2a$10$2bLycFqUmc6m6iLkaeUgKOGwzekGd9IoAPMbXRNNuJ8Sv9ItgV29O", 1);
