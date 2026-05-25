CREATE TABLE user_settings
(
    user_id  BIGINT PRIMARY KEY,
    settings JSONB,
    version  BIGINT NOT NULL DEFAULT 0
);
