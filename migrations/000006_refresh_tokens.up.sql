CREATE TABLE refresh_tokens (
                                jti        uuid PRIMARY KEY,
                                user_id    uuid NOT NULL,
                                expires_at timestamptz NOT NULL,
                                revoked    boolean NOT NULL DEFAULT false,
                                created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_refresh_tokens_user ON refresh_tokens (user_id);