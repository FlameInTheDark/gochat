ALTER TABLE users
    ADD COLUMN IF NOT EXISTS banner BIGINT;

CREATE TABLE IF NOT EXISTS user_notes
(
    owner_user_id  BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    target_user_id BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    note           TEXT        NOT NULL,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (owner_user_id, target_user_id)
);

