-- Rollback unique constraints and indexes added in 000009

DROP INDEX IF EXISTS idx_authentication_email;
DROP INDEX IF EXISTS idx_unique_member;
DROP INDEX IF EXISTS idx_unique_friend;
DROP INDEX IF EXISTS idx_unique_user_role;
DROP INDEX IF EXISTS idx_unique_channel_role_perm;
DROP INDEX IF EXISTS idx_unique_channel_user_perm;
DROP INDEX IF EXISTS idx_registration_email;
