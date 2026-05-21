ALTER TABLE guilds
    ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS guild_tags
(
    guild_id BIGINT NOT NULL,
    tag      TEXT   NOT NULL,
    PRIMARY KEY (guild_id, tag)
);

CREATE TABLE IF NOT EXISTS guild_discovery_stats
(
    guild_id       BIGINT PRIMARY KEY,
    members_count BIGINT NOT NULL DEFAULT 0
);

INSERT INTO guild_discovery_stats (guild_id, members_count)
SELECT g.id, COUNT(m.user_id)
FROM guilds g
LEFT JOIN members m ON m.guild_id = g.id
GROUP BY g.id
ON CONFLICT (guild_id) DO UPDATE SET members_count = EXCLUDED.members_count;
