CREATE TABLE password_reset_codes (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash text NOT NULL,
    expires_at timestamptz NOT NULL,
    used_at timestamptz,
    attempts integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL,

    CONSTRAINT password_reset_codes_hash_not_empty
        CHECK (btrim(code_hash) <> ''),
    CONSTRAINT password_reset_codes_attempts_non_negative
        CHECK (attempts >= 0)
);

CREATE INDEX password_reset_codes_user_id_idx
    ON password_reset_codes (user_id);

CREATE INDEX password_reset_codes_expires_at_idx
    ON password_reset_codes (expires_at);

CREATE UNIQUE INDEX password_reset_codes_user_active_unique
    ON password_reset_codes (user_id)
    WHERE used_at IS NULL;