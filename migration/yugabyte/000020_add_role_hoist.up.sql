-- +migrate StatementBegin
ALTER TABLE roles ADD COLUMN IF NOT EXISTS hoist BOOLEAN NOT NULL DEFAULT FALSE;
-- +migrate StatementEnd
