CREATE TABLE IF NOT EXISTS application_commands (
    id BIGINT PRIMARY KEY,
    application_id BIGINT NOT NULL REFERENCES bots(bot_user_id) ON DELETE CASCADE,
    guild_id BIGINT NULL,
    version BIGINT NOT NULL,
    type INT NOT NULL,
    name TEXT NOT NULL,
    name_localizations JSONB NOT NULL DEFAULT '{}'::jsonb,
    description TEXT NOT NULL DEFAULT '',
    description_localizations JSONB NOT NULL DEFAULT '{}'::jsonb,
    options JSONB NOT NULL DEFAULT '[]'::jsonb,
    default_member_permissions BIGINT NULL,
    contexts JSONB NOT NULL DEFAULT '[]'::jsonb,
    integration_types JSONB NOT NULL DEFAULT '[]'::jsonb,
    nsfw BOOL NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT application_commands_type_check CHECK (type IN (1, 2, 3))
);

CREATE UNIQUE INDEX IF NOT EXISTS application_commands_scope_name_uidx
    ON application_commands (application_id, COALESCE(guild_id, 0), type, lower(name));

CREATE INDEX IF NOT EXISTS application_commands_guild_idx
    ON application_commands (guild_id, type, lower(name));

CREATE INDEX IF NOT EXISTS application_commands_application_idx
    ON application_commands (application_id, guild_id);

CREATE TABLE IF NOT EXISTS application_command_interactions (
    id BIGINT PRIMARY KEY,
    application_id BIGINT NOT NULL REFERENCES bots(bot_user_id) ON DELETE CASCADE,
    command_id BIGINT NOT NULL,
    type INT NOT NULL,
    guild_id BIGINT NULL,
    channel_id BIGINT NOT NULL,
    invoker_user_id BIGINT NOT NULL,
    token_hash TEXT NOT NULL,
    token_prefix TEXT NOT NULL,
    app_permissions BIGINT NOT NULL DEFAULT 0,
    locale TEXT NOT NULL DEFAULT '',
    guild_locale TEXT NOT NULL DEFAULT '',
    context INT NOT NULL DEFAULT 0,
    ack_state TEXT NOT NULL DEFAULT 'pending',
    initial_response_message_id BIGINT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    responded_at TIMESTAMPTZ NULL,
    CONSTRAINT application_command_interactions_type_check CHECK (type IN (2, 4)),
    CONSTRAINT application_command_interactions_ack_state_check CHECK (ack_state IN ('pending', 'deferred', 'responded', 'expired'))
);

CREATE UNIQUE INDEX IF NOT EXISTS application_command_interactions_token_uidx
    ON application_command_interactions (application_id, token_hash);

CREATE INDEX IF NOT EXISTS application_command_interactions_expiry_idx
    ON application_command_interactions (expires_at);

CREATE INDEX IF NOT EXISTS application_command_interactions_command_idx
    ON application_command_interactions (command_id, created_at DESC);
