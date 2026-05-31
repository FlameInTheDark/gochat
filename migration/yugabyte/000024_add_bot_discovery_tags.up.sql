CREATE TABLE IF NOT EXISTS bot_tags
(
    bot_user_id BIGINT NOT NULL,
    tag         TEXT   NOT NULL,
    PRIMARY KEY (bot_user_id, tag)
);

CREATE INDEX IF NOT EXISTS idx_bot_tags_tag ON bot_tags (tag);
