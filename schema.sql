grant all privileges on database grapevine to grapevine;
alter role grapevine superuser;

CREATE TABLE users ( id uuid PRIMARY KEY, username varchar(255) UNIQUE NOT NULL, password_hash varchar(255) NOT NULL);

CREATE TABLE lists ( id uuid PRIMARY KEY, list_type varchar(255) NOT NULL, owner_id uuid NOT NULL, entity_id uuid NOT NULL);

CREATE TABLE user_contacts ( id uuid PRIMARY KEY, ip inet NOT NULL, port bigint NOT NULL, timestamp timestamptz NOT NULL);
