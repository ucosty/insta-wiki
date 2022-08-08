-- +goose Up
CREATE TABLE pages
(
    id          integer primary key autoincrement,
    user_id     integer,
    revision_id integer,
    path        text,
    title       text,
    created_at  datetime,
    updated_at  datetime
);

CREATE UNIQUE INDEX pages_path_idx on pages (path);

-- +goose Down
DROP TABLE pages;
DROP INDEX pages_path_idx;
