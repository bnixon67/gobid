CREATE TABLE `events` (
  `name` varchar(10) NOT NULL,
  `succeeded` boolean NOT NULL,
  `username` varchar(30) NOT NULL,
  `message` varchar(255) NOT NULL DEFAULT "",
  `created` timestamp(6) NOT NULL DEFAULT current_timestamp,
  PRIMARY KEY (`created`,`name`,`username`)
);
