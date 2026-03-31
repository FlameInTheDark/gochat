-- +migrate StatementBegin
DO $$
BEGIN
    EXECUTE 'DROP INDEX IF EXISTS idx_unique_dm_channel_by_channel_user';
    EXECUTE 'CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_dm_channel ON dm_channels (channel_id)';
END
$$;
-- +migrate StatementEnd
