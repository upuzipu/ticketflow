CREATE TABLE ticket_categories (
                                   id          uuid PRIMARY KEY,
                                   event_id    uuid NOT NULL REFERENCES events(id),
                                   name        text NOT NULL,
                                   price_minor bigint NOT NULL CHECK (price_minor >= 0),
                                   currency    text NOT NULL,
                                   total_qty   int  NOT NULL CHECK (total_qty > 0),
                                   UNIQUE (event_id, name)
);

CREATE TABLE tickets (
                         id          uuid PRIMARY KEY,
                         event_id    uuid NOT NULL REFERENCES events(id) ON DELETE CASCADE,
                         category_id uuid NOT NULL REFERENCES ticket_categories(id) ON DELETE CASCADE,
                         status      text NOT NULL DEFAULT 'available',
                         hold_id     uuid,
                         code        text UNIQUE,
                         version     int  NOT NULL DEFAULT 0,
                         CONSTRAINT tickets_status_check CHECK (status IN ('available', 'held', 'sold'))
);

CREATE INDEX idx_tickets_category_available ON tickets (category_id) WHERE status = 'available';