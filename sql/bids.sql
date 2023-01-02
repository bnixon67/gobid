CREATE TABLE `bids` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `created` timestamp NOT NULL DEFAULT current_timestamp(),
  `bidder` varchar(30) NOT NULL,
  `amount` decimal(13,2) NOT NULL,
  PRIMARY KEY (`id`,`created`)
);
