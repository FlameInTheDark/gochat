CREATE TABLE IF NOT EXISTS gochat.thread_created_messages
(
    thread_id  bigint PRIMARY KEY,
    channel_id bigint,
    message_id bigint
);
