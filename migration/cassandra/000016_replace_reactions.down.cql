DROP TABLE IF EXISTS gochat.reaction_members_by_bucket;
DROP TABLE IF EXISTS gochat.reactions_by_message_user;
DROP TABLE IF EXISTS gochat.reaction_summaries_by_message;

CREATE TABLE IF NOT EXISTS gochat.reactions
(
    message_id bigint,
    user_id    bigint,
    emote_id   bigint,
    PRIMARY KEY ( message_id, user_id )
);
