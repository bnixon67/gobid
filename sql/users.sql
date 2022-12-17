CREATE TABLE `users` (
  `userName` varchar(30) NOT NULL,
  `fullName` varchar(70) NOT NULL,
  `email` varchar(256) NOT NULL,
  `hashedPassword` binary(60) NOT NULL,
  `admin` boolean NOT NULL DEFAULT false,
  `created` timestamp NOT NULL DEFAULT current_timestamp(),
  PRIMARY KEY (`userName`),
  UNIQUE KEY `email` (`email`)
);
