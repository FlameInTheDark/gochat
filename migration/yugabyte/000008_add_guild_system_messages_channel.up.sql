ALTER TABLE guilds
    ADD COLUMN IF NOT EXISTS system_messages BIGINT;
