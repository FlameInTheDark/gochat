ALTER TABLE gochat.messages ADD interaction_id bigint;
ALTER TABLE gochat.messages ADD interaction_application_id bigint;
ALTER TABLE gochat.messages ADD interaction_command_id bigint;
ALTER TABLE gochat.messages ADD interaction_command_name text;
ALTER TABLE gochat.messages ADD interaction_user_id bigint;

CREATE TABLE IF NOT EXISTS gochat.application_command_interactions
(
    id              bigint PRIMARY KEY,
    application_id  bigint,
    command_id      bigint,
    type            int,
    guild_id        bigint,
    channel_id      bigint,
    invoker_user_id bigint,
    data            text,
    created_at      timestamp
);
