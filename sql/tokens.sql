CREATE TABLE `tokens` (
  `hashedValue` binary(64) NOT NULL,
  `expires` datetime NOT NULL,
  `type` varchar(7) NOT NULL,
  `userName` varchar(30) NOT NULL,
  `created` timestamp NOT NULL DEFAULT current_timestamp(),
  PRIMARY KEY (`hashedValue`)
);
