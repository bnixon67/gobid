DROP TABLE IF EXISTS `items`;

CREATE TABLE `items` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `title` varchar(40) NOT NULL,
  `created` timestamp NOT NULL DEFAULT current_timestamp(),
  `modified` timestamp NULL DEFAULT NULL,
  `modifiedBy` varchar(30) DEFAULT NULL,
  `description` varchar(255) DEFAULT NULL,
  `openingBid` decimal(13,2) NOT NULL,
  `midBidIncr` decimal(13,2) NOT NULL,
  `currentBid` decimal(13,2) NOT NULL,
  `artist` varchar(30) NOT NULL,
  `imageFileName` varchar(255) NOT NULL,
  PRIMARY KEY (`id`)
);
