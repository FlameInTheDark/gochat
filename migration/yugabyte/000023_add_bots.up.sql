ALTER TABLE users ADD COLUMN IF NOT EXISTS flags BIGINT NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS bots
(
    bot_user_id          BIGINT PRIMARY KEY,
    owner_user_id        BIGINT      NOT NULL,
    description          TEXT        NOT NULL DEFAULT '',
    public               BOOL        NOT NULL DEFAULT false,
    default_permissions  BIGINT      NOT NULL DEFAULT 0,
    disabled             BOOL        NOT NULL DEFAULT false,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_bots_owner ON bots (owner_user_id);
CREATE INDEX IF NOT EXISTS idx_bots_public ON bots (public);

CREATE TABLE IF NOT EXISTS bot_tokens
(
    id           BIGINT PRIMARY KEY,
    bot_user_id  BIGINT      NOT NULL,
    name         TEXT        NOT NULL,
    token_hash   TEXT        NOT NULL,
    token_prefix TEXT        NOT NULL,
    revoked_at   TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_bot_tokens_hash ON bot_tokens (token_hash);
CREATE INDEX IF NOT EXISTS idx_bot_tokens_bot ON bot_tokens (bot_user_id);

CREATE TABLE IF NOT EXISTS bot_install_grants
(
    id                    BIGINT PRIMARY KEY,
    bot_user_id           BIGINT      NOT NULL,
    owner_user_id         BIGINT      NOT NULL,
    token_hash            TEXT        NOT NULL,
    token_prefix          TEXT        NOT NULL,
    requested_permissions BIGINT      NOT NULL,
    expires_at            TIMESTAMPTZ NOT NULL,
    max_uses              INT         NOT NULL DEFAULT 1,
    uses                  INT         NOT NULL DEFAULT 0,
    revoked_at            TIMESTAMPTZ,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_bot_install_grants_hash ON bot_install_grants (token_hash);
CREATE INDEX IF NOT EXISTS idx_bot_install_grants_bot ON bot_install_grants (bot_user_id);

CREATE TABLE IF NOT EXISTS bot_guilds
(
    bot_user_id          BIGINT      NOT NULL,
    guild_id             BIGINT      NOT NULL,
    granted_permissions  BIGINT      NOT NULL,
    installer_user_id    BIGINT      NOT NULL,
    grant_id             BIGINT,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (bot_user_id, guild_id)
);
CREATE INDEX IF NOT EXISTS idx_bot_guilds_guild ON bot_guilds (guild_id);
