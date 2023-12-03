CREATE TABLE IF NOT EXISTS bookmarks
(
    id SERIAL PRIMARY KEY,
    title VARCHAR(512) NOT NULL,
    link VARCHAR(2048) NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT now(),
    tag VARCHAR(255) DEFAULT NULL
);

ALTER TABLE bookmarks ADD COLUMN ts tsvector GENERATED ALWAYS AS (to_tsvector('english', title)) STORED;

CREATE INDEX ts_idx ON bookmarks USING GIN (ts);

DROP TABLE bookmarks;