CREATE TABLE events (
                        id           uuid PRIMARY KEY,
                        organizer_id uuid NOT NULL,
                        title        text        NOT NULL,
                        description  text        NOT NULL DEFAULT '',
                        starts_at    timestamptz NOT NULL,
                        status       text        NOT NULL,
                        created_at   timestamptz NOT NULL DEFAULT now(),
                        updated_at   timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_events_status_starts ON events (status, starts_at);