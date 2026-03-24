DROP TABLE IF EXISTS gochat.reactions;

CREATE TABLE IF NOT EXISTS gochat.reactions_by_message_user
(
    message_id  bigint,
    user_id     bigint,
    bucket_key  text,
    reaction_id bigint,
    custom      boolean,
    emoji_id    bigint,
    emoji_name  text,
    PRIMARY KEY ( (message_id, user_id), bucket_key )
);

CREATE TABLE IF NOT EXISTS gochat.reaction_members_by_bucket
(
    message_id  bigint,
    bucket_key  text,
    reaction_id bigint,
    user_id     bigint,
    custom      boolean,
    emoji_id    bigint,
    emoji_name  text,
    PRIMARY KEY ( (message_id, bucket_key), reaction_id )
) WITH CLUSTERING ORDER BY (reaction_id DESC);

CREATE TABLE IF NOT EXISTS gochat.reaction_summaries_by_message
(
    message_id bigint,
    bucket_key text,
    custom     boolean,
    emoji_id   bigint,
    emoji_name text,
    count      int,
    PRIMARY KEY ( (message_id), bucket_key )
);
