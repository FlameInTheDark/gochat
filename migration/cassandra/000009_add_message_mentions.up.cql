ALTER TABLE gochat.messages DROP thread_id;

CREATE TABLE IF NOT EXISTS gochat.mentions
(
    user_id bigint,
    channel_id bigint,
    message_id bigint,
    author_id bigint,
    PRIMARY KEY ((user_id, channel_id), message_id)
);

CREATE TABLE IF NOT EXISTS gochat.channel_mentions
(
    guild_id bigint,
    channel_id bigint,
    message_id bigint,
    role_id bigint,
    author_id bigint,
    type int,
    PRIMARY KEY ( channel_id, message_id )
)