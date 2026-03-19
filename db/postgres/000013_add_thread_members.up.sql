CREATE TABLE thread_members
(
    thread_id BIGINT      NOT NULL,
    user_id   BIGINT      NOT NULL,
    flags     INT         NOT NULL DEFAULT 0,
    join_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_thread_members_thread_id_user_id ON thread_members (thread_id, user_id);
CREATE INDEX idx_thread_members_user_id_thread_id ON thread_members (user_id, thread_id);

SELECT create_distributed_table('thread_members', 'thread_id');
