DROP TABLE IF EXISTS `tokens`;

CREATE TABLE `tokens` (
  `hashedValue` binary(64) NOT NULL,
  `expires` datetime NOT NULL,
  `type` varchar(7) NOT NULL,
  `userName` varchar(30) NOT NULL,
  PRIMARY KEY (`hashedValue`)
);
