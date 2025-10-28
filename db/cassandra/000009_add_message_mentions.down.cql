DROP TABLE IF EXISTS gochat.channel_mentions;
DROP TABLE IF EXISTS gochat.mentions;
ALTER TABLE gochat.messages ADD thread_id bigint;