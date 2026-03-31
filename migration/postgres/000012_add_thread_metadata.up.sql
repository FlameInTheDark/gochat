ALTER TABLE channels
    ADD COLUMN creator_id BIGINT,
    ADD COLUMN closed BOOL NOT NULL DEFAULT false;
