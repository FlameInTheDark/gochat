CREATE TABLE friend_requests
(
    user_id   BIGINT PRIMARY KEY,
    friend_id BIGINT NOT NULL
);
SELECT create_distributed_table('friend_requests', 'user_id');

CREATE TABLE blocked_users
(
    user_id         BIGINT PRIMARY KEY,
    blocked_user_id BIGINT NOT NULL
);
SELECT create_distributed_table('blocked_users', 'user_id');