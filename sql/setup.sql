CREATE DATABASE gobid;
CREATE USER gobid IDENTIFIED BY 'password';
GRANT ALL PRIVILEGES ON gobid.* TO gobid;

CREATE DATABASE gobid_test;
CREATE USER gobid_test IDENTIFIED BY 'password';
GRANT ALL PRIVILEGES ON gobid_test.* TO gobid_test;
