-- +migrate StatementBegin
DO $$
BEGIN
    EXECUTE 'DROP INDEX IF EXISTS idx_roles_guild_id_position_id';
    EXECUTE 'ALTER TABLE roles DROP COLUMN IF EXISTS position';
END
$$;
-- +migrate StatementEnd
