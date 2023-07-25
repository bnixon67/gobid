CREATE TABLE `events` (
  `userName` varchar(30) NOT NULL,
  `created` timestamp(6) NOT NULL DEFAULT current_timestamp,
  `action` varchar(10) NOT NULL,
  `result` boolean NOT NULL,
  `message` varchar(255) NOT NULL DEFAULT "",
  PRIMARY KEY (`userName`,`created`)
);
