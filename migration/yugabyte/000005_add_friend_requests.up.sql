CREATE TABLE friend_requests
(
    user_id   BIGINT PRIMARY KEY,
    friend_id BIGINT NOT NULL
);

CREATE TABLE blocked_users
(
    user_id         BIGINT PRIMARY KEY,
    blocked_user_id BIGINT NOT NULL
);
