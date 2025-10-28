CREATE TABLE IF NOT EXISTS gochat.dm_channels_last_messages
(
    channel_id      bigint,
    last_message_id bigint,
    PRIMARY KEY ( channel_id )
);
