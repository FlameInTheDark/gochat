-- Add missing UNIQUE constraints and indexes for data integrity and query performance.
-- These constraints also help Citus optimize shard-local operations.

-- Index for email lookup on authentications.
-- NOTE: authentications is distributed by user_id, so a globally-enforced unique
-- index on email alone is not supported by Citus. This is a plain index for
-- scatter-gather query performance (e.g. login by email).
CREATE INDEX IF NOT EXISTS idx_authentication_email
    ON authentications (email);

-- Prevent duplicate memberships
CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_member
    ON members (guild_id, user_id);

-- Prevent duplicate friend entries
CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_friend
    ON friends (user_id, friend_id);

-- Prevent duplicate role assignments
CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_user_role
    ON user_roles (guild_id, user_id, role_id);

-- Prevent duplicate channel role permission entries
CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_channel_role_perm
    ON channel_roles_permissions (channel_id, role_id);

-- Prevent duplicate channel user permission entries
CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_channel_user_perm
    ON channel_user_permissions (channel_id, user_id);

-- Speed up registration lookup by email (currently scatters, but index helps per-shard)
CREATE INDEX IF NOT EXISTS idx_registration_email
    ON registrations (email);
