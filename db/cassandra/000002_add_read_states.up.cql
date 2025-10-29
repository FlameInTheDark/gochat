CREATE TABLE IF NOT EXISTS gochat.read_states
(
    user_id  bigint,
    channels map<bigint, bigint>,
    PRIMARY KEY ( user_id )
);