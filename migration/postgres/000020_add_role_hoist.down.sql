-- +migrate StatementBegin
ALTER TABLE roles DROP COLUMN IF EXISTS hoist;
-- +migrate StatementEnd
