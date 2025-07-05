CREATE TABLE IF NOT EXISTS gochat.attachments
(
    id           bigint,
    channel_id   bigint,
    name         text,
    filesize     bigint,
    content_type text,
    height       bigint,
    width        bigint,
    done         boolean,
    PRIMARY KEY ( (channel_id), id )
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

CREATE TABLE IF NOT EXISTS gochat.reactions
(
    message_id bigint,
    user_id    bigint,
    emote_id   bigint,
    PRIMARY KEY ( message_id, user_id )
);
