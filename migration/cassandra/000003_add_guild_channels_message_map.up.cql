CREATE TABLE IF NOT EXISTS gochat.guild_channels_last_messages
(
    guild_id  bigint,
    channels map<bigint, bigint>,
    PRIMARY KEY ( guild_id )
);
