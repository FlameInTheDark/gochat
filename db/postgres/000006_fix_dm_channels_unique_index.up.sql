DROP INDEX IF EXISTS idx_unique_dm_channel;
DROP INDEX IF EXISTS idx_unique_dm_channel_102317;

CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_dm_channel_by_channel_user
    ON dm_channels (channel_id, user_id);

