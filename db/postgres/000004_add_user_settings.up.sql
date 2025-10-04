CREATE TABLE user_settings
(
    user_id  BIGINT PRIMARY KEY,
    settings JSONB,
    version  BIGINT NOT NULL DEFAULT 0
);
SELECT create_distributed_table('user_settings', 'user_id');