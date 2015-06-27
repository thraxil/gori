CREATE TABLE pages (
    slug varchar(256) NOT NULL,
    title text,
    body text,
		created timestamp,
		modified timestamp
);

CREATE UNIQUE index slug_idx on pages (slug);
CREATE UNIQUE index title_idx on pages (title);

CREATE TABLE events (
    id uuid primary key,
    command text not null,
    aggregate_id text not null,
    created timestamp default current_timestamp,
    event_data text,
    event_context text
);

CREATE INDEX events_aggregate_id_idx on events (aggregate_id);
