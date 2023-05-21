create database grapevine;
create user grapevine with encrypted password 'grapevine';
grant all privileges on database grapevine to grapevine;

\c grapevine

DROP TABLE user_contacts;
DROP TABLE lists;
DROP TABLE users;
