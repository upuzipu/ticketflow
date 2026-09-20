CREATE TABLE holds (
                       id          uuid PRIMARY KEY,
                       user_id     uuid NOT NULL,
                       event_id    uuid NOT NULL REFERENCES events(id) ON DELETE CASCADE,
                       category_id uuid NOT NULL REFERENCES ticket_categories(id) ON DELETE CASCADE,
                       ticket_ids  uuid[] NOT NULL,
                       status      text NOT NULL,
                       expires_at  timestamptz NOT NULL,
                       created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_holds_status_expires ON holds (expires_at) WHERE status = 'active';