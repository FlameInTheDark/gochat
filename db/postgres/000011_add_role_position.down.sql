DROP INDEX IF EXISTS idx_roles_guild_id_position_id;

ALTER TABLE roles DROP COLUMN IF EXISTS position;
