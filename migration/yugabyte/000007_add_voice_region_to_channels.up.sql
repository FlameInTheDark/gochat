ALTER TABLE channels
    ADD COLUMN IF NOT EXISTS voice_region TEXT;
