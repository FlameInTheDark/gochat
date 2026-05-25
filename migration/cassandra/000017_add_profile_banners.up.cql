CREATE TABLE IF NOT EXISTS gochat.banners
(
    id           bigint,
    user_id      bigint,
    url          text,
    content_type text,
    width        bigint,
    height       bigint,
    filesize     bigint,
    done         boolean,
    PRIMARY KEY ( user_id, id )
);

