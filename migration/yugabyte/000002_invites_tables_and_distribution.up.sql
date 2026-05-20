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
