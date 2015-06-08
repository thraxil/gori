CREATE TABLE pages (
    slug varchar(256) NOT NULL,
    title text,
    body text,
		created timestamp,
		modified timestamp
);

CREATE UNIQUE index title_idx on pages (title);