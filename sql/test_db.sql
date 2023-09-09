DROP TABLE IF EXISTS users;
source users.sql;

INSERT INTO users(userName, fullName, email, hashedPassword)
VALUES ('test', 'Test User', 'test@email', '$2a$10$2bLycFqUmc6m6iLkaeUgKOGwzekGd9IoAPMbXRNNuJ8Sv9ItgV29O');

INSERT INTO users(userName, fullName, email, hashedPassword, admin)
VALUES ('admin', 'Admin User', 'admin@email', '$2a$10$2bLycFqUmc6m6iLkaeUgKOGwzekGd9IoAPMbXRNNuJ8Sv9ItgV29O', 1);

DROP TABLE IF EXISTS tokens;
source tokens.sql;

DROP TABLE IF EXISTS events;
source events.sql;

INSERT INTO events(userName, created, name, result)
VALUES
("test1", "2023-01-15 01:00:00", "login", true),

("test2", "2023-01-15 01:00:00", "login", true),
("test2", "2023-01-15 02:00:00", "login", true),

("test3", "2023-01-15 03:00:00", "login", true),
("test3", "2023-01-15 02:00:00", "login", true),
("test3", "2023-01-15 01:00:00", "login", true),

("test4", "2023-01-15 01:00:00", "login", true),
("test4", "2023-01-15 04:00:00", "login", true),
("test4", "2023-01-15 02:00:00", "login", true),
("test4", "2023-01-15 03:00:00", "login", true);
