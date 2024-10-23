CREATE TABLE IF NOT EXISTS gochat.attachments
(
    id     bigint,
    object text,
    PRIMARY KEY ( id )
);

CREATE TABLE IF NOT EXISTS gochat.icons
(
    id       bigint,
    guild_id bigint,
    object   text,
    PRIMARY KEY ( guild_id, id )
);

CREATE TABLE IF NOT EXISTS gochat.avatars
(
    id      bigint,
    user_id bigint,
    object  text,
    PRIMARY KEY ( user_id, id )
);

CREATE TABLE IF NOT EXISTS gochat.authentications
(
    user_id       bigint,
    email         text,
    password_hash text,
    created_at    timestamp,
    PRIMARY KEY ( user_id, email )
);

CREATE TABLE IF NOT EXISTS gochat.registrations
(
    user_id            bigint,
    email              text,
    confirmation_token text,
    created_at         timestamp,
    PRIMARY KEY ( user_id, email )
);

CREATE TABLE IF NOT EXISTS gochat.users
(
    id           bigint,
    name         text,
    determinator text,
    about        text,
    avatar       bigint,
    blocked      boolean,
    created_at   timestamp,
    PRIMARY KEY ( id, determinator )
);

CREATE TABLE IF NOT EXISTS gochat.guilds
(
    id         bigint,
    name       text,
    owner_id   bigint,
    icon       bigint,
    public     boolean,
    created_at timestamp,
    PRIMARY KEY ( id )
);

CREATE TABLE IF NOT EXISTS gochat.guild_channels
(
    guild_id    bigint,
    channel_id  bigint,
    permissions bigint,
    PRIMARY KEY ( guild_id, channel_id )
);

CREATE TABLE IF NOT EXISTS gochat.members
(
    guild_id bigint,
    user_id  bigint,
    username text,
    avatar   bigint,
    join_at  timestamp,
    PRIMARY KEY ( guild_id, user_id )
);

CREATE TABLE IF NOT EXISTS gochat.roles
(
    id          bigint,
    guild_id    bigint,
    name        text,
    color       int,
    permissions bigint,
    PRIMARY KEY ( guild_id, id )
);

CREATE TABLE IF NOT EXISTS gochat.user_roles
(
    guild_id bigint,
    user_id  bigint,
    role_id  bigint,
    PRIMARY KEY ( guild_id, user_id )
);

CREATE TABLE IF NOT EXISTS gochat.channel_role_permissions
(
    channel_id bigint,
    role_id    bigint,
    PRIMARY KEY ( channel_id, role_id )
);

CREATE TABLE IF NOT EXISTS gochat.channel_user_permissions
(
    channel_id  bigint,
    user_id     bigint,
    permissions bigint,
    PRIMARY KEY ( channel_id, user_id )
);

CREATE TABLE IF NOT EXISTS gochat.channels
(
    id          bigint,
    name        text,
    type        int,
    parent_id   bigint,
    permissions bigint,
    created_at  timestamp,
    PRIMARY KEY ( (id), parent_id, type )
);

CREATE TABLE IF NOT EXISTS gochat.dm_channels
(
    user_id        bigint,
    participant_id bigint,
    channel_id     bigint,
    PRIMARY KEY ( user_id, participant_id )
);

CREATE TABLE IF NOT EXISTS gochat.messages
(
    channel_id  bigint,
    bucket      int,
    id          bigint,
    user_id     bigint,
    content     text,
    thread_id   bigint,
    attachments list<bigint>,
    type        int,
    reference   bigint,
    thread      bigint,
    edited_at   timestamp,
    PRIMARY KEY ( (channel_id, bucket), id )
) WITH CLUSTERING ORDER BY (id DESC);

CREATE TABLE IF NOT EXISTS gochat.friends
(
    user_id    bigint,
    friend_id  bigint,
    created_at timestamp,
    PRIMARY KEY ( user_id, friend_id )
);

CREATE TABLE IF NOT EXISTS gochat.blocked
(
    user_id    bigint,
    blocked_id bigint,
    PRIMARY KEY ( user_id, blocked_id )
);

CREATE TABLE IF NOT EXISTS gochat.banned
(
    guild_id bigint,
    user_id  bigint,
    PRIMARY KEY ( guild_id, user_id )
);

CREATE TABLE IF NOT EXISTS gochat.reactions
(
    message_id bigint,
    user_id    bigint,
    emote_id   bigint,
    PRIMARY KEY ( message_id, user_id )
);

CREATE TABLE IF NOT EXISTS gochat.audit_log
(
    guild_id   bigint,
    created_at timestamp,
    changes    map<text, text>,
    PRIMARY KEY ( guild_id, created_at )
);
