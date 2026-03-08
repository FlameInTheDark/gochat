CREATE TABLE IF NOT EXISTS gochat.banned
(
    guild_id text,
    user_id  text,
    PRIMARY KEY ((guild_id), user_id)
);

ALTER TABLE gochat.banned ADD reason text;
