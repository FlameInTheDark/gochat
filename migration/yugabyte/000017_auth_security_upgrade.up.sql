ALTER TABLE authentications
    ADD COLUMN IF NOT EXISTS session_version BIGINT NOT NULL DEFAULT 1;

CREATE TABLE auth_factors
(
    user_id      BIGINT      NOT NULL,
    factor_id    BIGINT      NOT NULL,
    factor_type  TEXT        NOT NULL,
    display_name TEXT        NOT NULL,
    status       TEXT        NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    verified_at  TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    PRIMARY KEY (user_id, factor_id)
);
CREATE INDEX idx_auth_factors_user_status ON auth_factors (user_id, status);
CREATE UNIQUE INDEX idx_auth_factors_active_totp ON auth_factors (user_id) WHERE factor_type = 'totp' AND status = 'active';

CREATE TABLE auth_totp_factors
(
    user_id           BIGINT      NOT NULL,
    factor_id         BIGINT      NOT NULL,
    secret_ciphertext BYTEA       NOT NULL,
    secret_nonce      BYTEA       NOT NULL,
    algorithm         TEXT        NOT NULL,
    digits            INT         NOT NULL,
    period_seconds    INT         NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, factor_id)
);
CREATE INDEX idx_auth_totp_factors_user_factor ON auth_totp_factors (user_id, factor_id);

CREATE TABLE auth_recovery_codes
(
    user_id    BIGINT      NOT NULL,
    factor_id  BIGINT      NOT NULL,
    code_id    BIGINT      NOT NULL,
    code_hash  TEXT        NOT NULL,
    used_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, code_id)
);
CREATE INDEX idx_auth_recovery_codes_factor ON auth_recovery_codes (user_id, factor_id);
