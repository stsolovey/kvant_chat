-- noinspection SqlNoDataSourceInspectionForFiles
-- +migrate Up

CREATE TABLE users (
    user_id SERIAL PRIMARY KEY,
    username VARCHAR UNIQUE,
    hashed_password TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted BOOLEAN DEFAULT FALSE
);

-- +migrate Down

DROP TABLE IF EXISTS users;