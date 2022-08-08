-- +goose Up
CREATE TABLE revisions
(
    id         integer primary key autoincrement,
    page_id    integer,
    user_id    integer,
    created_at datetime,
    title      text,
    body       text
);

-- +goose Down
DROP TABLE revisions;
