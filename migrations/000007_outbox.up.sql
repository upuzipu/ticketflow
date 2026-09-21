CREATE TABLE outbox (
                        id         uuid PRIMARY KEY,
                        name       text NOT NULL,
                        key        text NOT NULL,
                        payload    jsonb NOT NULL,
                        published  boolean NOT NULL DEFAULT false,
                        attempts   int NOT NULL DEFAULT 0,
                        created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_outbox_unpublished ON outbox (created_at) WHERE published = false;