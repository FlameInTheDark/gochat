DROP TABLE IF EXISTS guild_discovery_stats;
DROP TABLE IF EXISTS guild_tags;

ALTER TABLE guilds
    DROP COLUMN IF EXISTS description;
