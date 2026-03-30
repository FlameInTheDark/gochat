# Auth Security

This document covers the password change and TOTP-based two-factor authentication flows implemented by the auth service.

## Scope

- All password and MFA operations live in `cmd/auth`.
- V1 supports one active TOTP factor per user.
- Email recovery is only a login fallback. It is not accepted for password changes, disabling 2FA, or regenerating recovery codes.

## User-facing flows

### Password change

- Route: `POST /api/v1/auth/password/change`
- Requires an authenticated access token.
- Request body always includes `current_password` and `new_password`.
- If 2FA is enabled, the same request must also include `code_type` (`totp` or `recovery_code`) and `code`.
- Success returns fresh access and refresh tokens for the current client and revokes older sessions.

### Enable TOTP

1. Call `POST /api/v1/auth/2fa/totp/setup` with the current password.
2. The auth service returns a short-lived TOTP setup payload with:
   - `setup_id`
   - `otpauth_uri`
   - `manual_key`
   - `issuer`
   - `account_name`
   - `expires_at`
3. The client renders the QR locally from `otpauth_uri`.
4. Call `POST /api/v1/auth/2fa/totp/confirm` with `setup_id` and a 6-digit authenticator code.
5. Success enables TOTP, returns fresh tokens, and returns recovery codes once.

### Login with 2FA enabled

1. `POST /api/v1/auth/login` still validates email and password first.
2. When TOTP is enabled, it responds with `202 Accepted` and a `challenge_id`.
3. The client completes the second step with one of:
   - `POST /api/v1/auth/login/2fa/totp`
   - `POST /api/v1/auth/login/2fa/recovery-code`
   - `POST /api/v1/auth/login/2fa/email/start` followed by `POST /api/v1/auth/login/2fa/email/verify`
4. Successful verification returns normal access and refresh tokens.

### Recovery codes and disable flow

- `POST /api/v1/auth/2fa/recovery-codes/regenerate` requires the current password plus a TOTP or recovery code.
- `DELETE /api/v1/auth/2fa` requires the current password plus a TOTP or recovery code.
- Both operations rotate the session version and invalidate older sessions.

## Service workflow

### Persistent state

- PostgreSQL stores:
  - `authentications.session_version`
  - `auth_factors`
  - `auth_totp_factors`
  - `auth_recovery_codes`
- TOTP secrets are encrypted before storage.
- Recovery codes are stored only as one-time password hashes.

### Ephemeral state

- Redis/KeyDB stores pending TOTP setup state with a 10-minute TTL.
- Redis/KeyDB stores login challenge state with a 5-minute TTL, attempt counters, and email recovery metadata.

### Session invalidation

- Password changes, TOTP enable, TOTP disable, and recovery-code regeneration bump `session_version`.
- JWTs now carry `session_version`.
- HTTP middleware and the WebSocket gateway validate that claim against Redis/KeyDB with PostgreSQL fallback.
- After a successful security mutation, auth publishes a user-scoped revocation event so existing WebSocket sessions close and reconnect with fresh tokens only.

## Operational notes

- Config adds `mfa_encryption_key` for TOTP secret encryption.
- Config adds `mfa_recovery_template` for the email recovery code template.
- Auth service rate limits now cover:
  - password login attempts by email and IP
  - second-step verification attempts by challenge and IP
  - email recovery sends by user and IP, including a resend cooldown
