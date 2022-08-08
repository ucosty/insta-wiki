-- +goose Up
CREATE TABLE users
(
    id         integer primary key autoincrement,
    username   text,
    created_at datetime
);

-- +goose Down
DROP TABLE users;
