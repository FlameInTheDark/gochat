CREATE TABLE IF NOT EXISTS gochat.message_threads
(
    channel_id bigint,
    message_id bigint,
    thread_id  bigint,
    PRIMARY KEY ((channel_id), message_id)
);
