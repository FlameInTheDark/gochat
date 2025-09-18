CREATE EXTENSION IF NOT EXISTS citus;

CREATE TABLE users
(
    id           BIGINT PRIMARY KEY,
    name         TEXT        NOT NULL,
    avatar       BIGINT,
    blocked      BOOL        NOT NULL,
    upload_limit BIGINT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_users_id_name ON users (id, name);
SELECT create_distributed_table('users', 'id');

CREATE TABLE authentications
(
    user_id       BIGINT      NOT NULL,
    email         TEXT        NOT NULL,
    password_hash TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_authentication_id_email ON authentications (user_id, email);
SELECT create_distributed_table('authentications', 'user_id');

CREATE TABLE registrations
(
    user_id            BIGINT PRIMARY KEY,
    email              TEXT        NOT NULL,
    confirmation_token TEXT        NOT NULL,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_registration_user_id ON registrations (user_id);
SELECT create_distributed_table('registrations', 'user_id');

CREATE TABLE IF NOT EXISTS recoveries
(
    user_id    BIGINT PRIMARY KEY,
    token      VARCHAR(64) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_recoveries_user_id ON recoveries (user_id);
CREATE INDEX IF NOT EXISTS idx_recoveries_token ON recoveries (token);
SELECT create_distributed_table('recoveries', 'user_id');

CREATE TABLE discriminators
(
    user_id       BIGINT NOT NULL,
    discriminator TEXT   NOT NULL
);
CREATE INDEX idx_discriminator ON discriminators (discriminator);
SELECT create_distributed_table('discriminators', 'discriminator');

CREATE TABLE guilds
(
    id          BIGINT PRIMARY KEY,
    name        TEXT        NOT NULL,
    owner_id    BIGINT      NOT NULL,
    icon        BIGINT,
    public      BOOL        NOT NULL DEFAULT false,
    permissions BIGINT      NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_guilds_id ON guilds (id);
SELECT create_distributed_table('guilds', 'id');

CREATE TABLE channels
(
    id           BIGINT PRIMARY KEY,
    name         TEXT        NOT NULL,
    type         INT         NOT NULL,
    parent_id    BIGINT,
    permissions  BIGINT,
    topic        TEXT,
    private      BOOL        NOT NULL,
    last_message BIGINT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_channels_id ON channels (id);
SELECT create_distributed_table('channels', 'id');

CREATE TABLE guild_channels
(
    guild_id   BIGINT NOT NULL,
    channel_id BIGINT NOT NULL,
    position   INT    NOT NULL
);
CREATE INDEX idx_guild_channels_ids ON guild_channels (guild_id, channel_id);
SELECT create_distributed_table('guild_channels', 'guild_id');

CREATE TABLE friends
(
    user_id    BIGINT      NOT NULL,
    friend_id  BIGINT      NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_user_friend_ids ON friends (user_id, friend_id);
SELECT create_distributed_table('friends', 'user_id');

CREATE TABLE dm_channels
(
    user_id        BIGINT NOT NULL,
    participant_id BIGINT NOT NULL,
    channel_id     BIGINT NOT NULL
);
CREATE INDEX idx_dm_channel_id_user_id_participant_id ON dm_channels (user_id, participant_id, channel_id);
CREATE UNIQUE INDEX idx_unique_dm_channel ON dm_channels (channel_id);
SELECT create_distributed_table('dm_channels', 'channel_id');

CREATE TABLE group_dm_channels
(
    channel_id BIGINT PRIMARY KEY,
    user_id    BIGINT NOT NULL
);
CREATE INDEX idx_group_dm_channels ON group_dm_channels (channel_id);
SELECT create_distributed_table('group_dm_channels', 'channel_id');

CREATE TABLE members
(
    user_id  BIGINT      NOT NULL,
    guild_id BIGINT      NOT NULL,
    username TEXT,
    avatar   BIGINT,
    join_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    timeout  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_guild_members ON members (user_id, guild_id);
SELECT create_distributed_table('members', 'guild_id');

CREATE TABLE roles
(
    id          BIGINT NOT NULL,
    guild_id    BIGINT NOT NULL,
    name        TEXT   NOT NULL,
    color       INT    NOT NULL,
    permissions BIGINT NOT NULL
);
CREATE INDEX idx_roles_id_guild_id ON roles (id, guild_id);
SELECT create_distributed_table('roles', 'guild_id');

CREATE TABLE user_roles
(
    guild_id BIGINT NOT NULL,
    user_id  BIGINT NOT NULL,
    role_id  BIGINT NOT NULL
);
CREATE INDEX idx_user_roles ON user_roles (user_id, guild_id);
SELECT create_distributed_table('user_roles', 'guild_id');

CREATE TABLE channel_roles_permissions
(
    channel_id BIGINT NOT NULL,
    role_id    BIGINT NOT NULL,
    accept     BIGINT NOT NULL,
    deny       BIGINT NOT NULL
);
CREATE INDEX idx_channel_roles_permission_ch_id_role_id ON channel_roles_permissions (channel_id, role_id);
SELECT create_distributed_table('channel_roles_permissions', 'channel_id');

CREATE TABLE channel_user_permissions
(
    channel_id BIGINT NOT NULL,
    user_id    BIGINT NOT NULL,
    accept     BIGINT NOT NULL,
    deny       BIGINT NOT NULL
);
CREATE INDEX idx_channel_user_permission_ch_id_user_id ON channel_user_permissions (channel_id, user_id);
SELECT create_distributed_table('channel_user_permissions', 'channel_id');

CREATE TABLE audit
(
    guild_id   BIGINT      NOT NULL,
    changes    JSONB       NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_audit_guild_id ON audit (guild_id);
SELECT create_distributed_table('audit', 'guild_id');
