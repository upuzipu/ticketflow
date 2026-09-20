CREATE TABLE orders (
                        id              uuid PRIMARY KEY,
                        user_id         uuid NOT NULL,
                        event_id        uuid NOT NULL,
                        hold_id         uuid NOT NULL REFERENCES holds(id),
                        status          text NOT NULL,
                        total_minor     bigint NOT NULL,
                        currency        text NOT NULL,
                        idempotency_key uuid NOT NULL,
                        version         int NOT NULL DEFAULT 0,
                        created_at      timestamptz NOT NULL DEFAULT now(),
                        updated_at      timestamptz NOT NULL DEFAULT now(),

                        CONSTRAINT orders_status_check CHECK (status IN ('pending', 'paid', 'failed', 'expired', 'confirmed', 'refunded')),
                        UNIQUE (user_id, idempotency_key)
);

CREATE TABLE payments (
                          id          uuid PRIMARY KEY,
                          order_id    uuid NOT NULL REFERENCES orders(id),
                          status      text NOT NULL,
                          amount_minor bigint NOT NULL,
                          currency    text NOT NULL,
                          gateway_ref text,
                          created_at  timestamptz NOT NULL DEFAULT now(),

                          CONSTRAINT payments_status_check CHECK (status IN ('pending', 'authorized', 'failed'))
);

CREATE INDEX idx_orders_user ON orders (user_id, created_at DESC);