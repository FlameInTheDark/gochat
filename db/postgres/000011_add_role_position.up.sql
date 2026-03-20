-- +migrate StatementBegin
DO $$
BEGIN
    EXECUTE 'ALTER TABLE roles ADD COLUMN position INT NOT NULL DEFAULT 0';

    WITH ordered_roles AS (
        SELECT guild_id, id, ROW_NUMBER() OVER (PARTITION BY guild_id ORDER BY id ASC) - 1 AS position
        FROM roles
    )
    UPDATE roles
    SET position = ordered_roles.position
    FROM ordered_roles
    WHERE roles.guild_id = ordered_roles.guild_id
      AND roles.id = ordered_roles.id;

    EXECUTE 'CREATE INDEX idx_roles_guild_id_position_id ON roles (guild_id, position, id)';
END
$$;
-- +migrate StatementEnd
