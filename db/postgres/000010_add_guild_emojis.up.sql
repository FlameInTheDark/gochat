-- +migrate StatementBegin
DO $$
BEGIN
    EXECUTE $sql$
    CREATE TABLE guild_emojis
    (
        guild_id           BIGINT      NOT NULL,
        id                 BIGINT      NOT NULL,
        name               TEXT        NOT NULL,
        name_normalized    TEXT        NOT NULL,
        creator_id         BIGINT      NOT NULL,
        done               BOOL        NOT NULL DEFAULT false,
        animated           BOOL        NOT NULL DEFAULT false,
        declared_file_size BIGINT      NOT NULL,
        actual_file_size   BIGINT,
        content_type       TEXT,
        width              BIGINT,
        height             BIGINT,
        upload_expires_at  TIMESTAMPTZ NOT NULL,
        created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
        updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
        PRIMARY KEY (guild_id, id)
    )
    $sql$;
    EXECUTE 'CREATE INDEX idx_guild_emojis_id ON guild_emojis (id)';
    EXECUTE 'CREATE UNIQUE INDEX idx_guild_emojis_unique_name ON guild_emojis (guild_id, name_normalized)';
    PERFORM create_distributed_table('guild_emojis', 'guild_id', colocate_with => 'guilds');

    EXECUTE $sql$
    CREATE TABLE emoji_lookup
    (
        id         BIGINT PRIMARY KEY,
        guild_id   BIGINT      NOT NULL,
        name       TEXT        NOT NULL,
        done       BOOL        NOT NULL DEFAULT false,
        animated   BOOL        NOT NULL DEFAULT false,
        width      BIGINT,
        height     BIGINT,
        created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
    )
    $sql$;
    EXECUTE 'CREATE INDEX idx_emoji_lookup_guild_id ON emoji_lookup (guild_id)';
    PERFORM create_distributed_table('emoji_lookup', 'id');
END
$$;
-- +migrate StatementEnd
