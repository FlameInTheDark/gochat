ALTER TABLE thread_members DROP CONSTRAINT IF EXISTS thread_members_pkey;
ALTER TABLE user_roles DROP CONSTRAINT IF EXISTS user_roles_pkey;
ALTER TABLE friends DROP CONSTRAINT IF EXISTS friends_pkey;
ALTER TABLE channel_user_permissions DROP CONSTRAINT IF EXISTS channel_user_permissions_pkey;
ALTER TABLE members DROP CONSTRAINT IF EXISTS members_pkey;
ALTER TABLE channel_roles_permissions DROP CONSTRAINT IF EXISTS channel_roles_permissions_pkey;
ALTER TABLE dm_channels DROP CONSTRAINT IF EXISTS dm_channels_pkey;

CREATE UNIQUE INDEX IF NOT EXISTS idx_thread_members_thread_id_user_id ON thread_members (thread_id, user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_user_role ON user_roles (guild_id, user_id, role_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_friend ON friends (user_id, friend_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_channel_user_perm ON channel_user_permissions (channel_id, user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_member ON members (guild_id, user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_channel_role_perm ON channel_roles_permissions (channel_id, role_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_dm_channel_by_channel_user ON dm_channels (channel_id, user_id);
