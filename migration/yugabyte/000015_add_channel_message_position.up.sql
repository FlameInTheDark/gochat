ALTER TABLE channels
    ADD COLUMN message_position BIGINT NOT NULL DEFAULT 0;
