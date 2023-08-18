CREATE TABLE `events` (
  `name` varchar(10) NOT NULL,
  `result` boolean NOT NULL,
  `userName` varchar(30) NOT NULL,
  `message` varchar(255) NOT NULL DEFAULT "",
  `created` timestamp(6) NOT NULL DEFAULT current_timestamp,
  PRIMARY KEY (`created`,`name`,`userName`)
);
