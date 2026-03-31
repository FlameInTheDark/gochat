DROP TABLE IF EXISTS auth_recovery_codes;
DROP TABLE IF EXISTS auth_totp_factors;
DROP TABLE IF EXISTS auth_factors;

ALTER TABLE authentications
    DROP COLUMN IF EXISTS session_version;
