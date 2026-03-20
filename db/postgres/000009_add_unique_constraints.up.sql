-- +migrate StatementBegin
DO $$
BEGIN
    EXECUTE 'CREATE INDEX IF NOT EXISTS idx_authentication_email ON authentications (email)';
    EXECUTE 'CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_member ON members (guild_id, user_id)';
    EXECUTE 'CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_friend ON friends (user_id, friend_id)';
    EXECUTE 'CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_user_role ON user_roles (guild_id, user_id, role_id)';
    EXECUTE 'CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_channel_role_perm ON channel_roles_permissions (channel_id, role_id)';
    EXECUTE 'CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_channel_user_perm ON channel_user_permissions (channel_id, user_id)';
    EXECUTE 'CREATE INDEX IF NOT EXISTS idx_registration_email ON registrations (email)';
END
$$;
-- +migrate StatementEnd
