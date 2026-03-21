-- +migrate StatementBegin
DO $$
BEGIN
    EXECUTE 'DROP INDEX IF EXISTS idx_authentication_email';
    EXECUTE 'DROP INDEX IF EXISTS idx_unique_member';
    EXECUTE 'DROP INDEX IF EXISTS idx_unique_friend';
    EXECUTE 'DROP INDEX IF EXISTS idx_unique_user_role';
    EXECUTE 'DROP INDEX IF EXISTS idx_unique_channel_role_perm';
    EXECUTE 'DROP INDEX IF EXISTS idx_unique_channel_user_perm';
    EXECUTE 'DROP INDEX IF EXISTS idx_registration_email';
END
$$;
-- +migrate StatementEnd
