-- Optional if already installed/managed
CREATE EXTENSION IF NOT EXISTS citus;

-- Tables
CREATE TABLE guild_invite_codes (
                                           invite_code  varchar(8) PRIMARY KEY,
                                           invite_id    bigint      NOT NULL,
                                           guild_id     bigint      NOT NULL,
                                           CONSTRAINT guild_invite_codes_len_chk CHECK (char_length(invite_code) = 8)
);

CREATE INDEX IF NOT EXISTS guild_invite_codes_invite_guild_idx
    ON guild_invite_codes (invite_id, guild_id);

CREATE TABLE guild_invites (
                                      guild_id    bigint      NOT NULL,
                                      invite_id   bigint      NOT NULL,
                                      author_id   bigint      NOT NULL,
                                      created_at  timestamptz NOT NULL DEFAULT now(),
                                      expires_at  timestamptz NOT NULL,
                                      CONSTRAINT guild_invites_pk PRIMARY KEY (guild_id, invite_id),
                                      CONSTRAINT guild_invites_expires_after_created_chk CHECK (expires_at > created_at)
);

CREATE INDEX IF NOT EXISTS guild_invites_guild_expires_idx
    ON guild_invites (guild_id, expires_at);

CREATE INDEX IF NOT EXISTS guild_invites_guild_author_idx
    ON guild_invites (guild_id, author_id);

-- Distribute with Citus (do NOT create any functions in this migration)
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_dist_partition
     WHERE logicalrelid = to_regclass('guild_invite_codes')
  ) THEN
    PERFORM create_distributed_table('guild_invite_codes', 'invite_code');
  END IF;
END$$;

DO $$
DECLARE has_guilds boolean;
BEGIN
  SELECT EXISTS (
    SELECT 1 FROM pg_dist_partition
     WHERE logicalrelid = to_regclass('guilds')
  ) INTO has_guilds;

  IF NOT EXISTS (
    SELECT 1 FROM pg_dist_partition
     WHERE logicalrelid = to_regclass('guild_invites')
  ) THEN
    IF has_guilds THEN
      PERFORM create_distributed_table('guild_invites', 'guild_id', colocate_with => 'guilds');
    ELSE
      PERFORM create_distributed_table('guild_invites', 'guild_id');
    END IF;
  END IF;
END$$;
