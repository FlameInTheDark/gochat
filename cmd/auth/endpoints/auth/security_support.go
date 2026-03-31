package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/FlameInTheDark/gochat/internal/cache"
	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/mailer"
	"github.com/FlameInTheDark/gochat/internal/mq"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/observability"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

const (
	loginChallengeTTL         = 5 * time.Minute
	pendingTOTPSetupTTL       = 10 * time.Minute
	emailRecoveryCodeTTL      = 5 * time.Minute
	emailRecoveryResendAfter  = time.Minute
	loginRateLimitTTL         = 5 * time.Minute
	challengeRateLimitTTL     = 5 * time.Minute
	emailRecoveryRateLimitTTL = time.Hour
	loginRateLimitMax         = int64(10)
	challengeRateLimitMax     = int64(10)
	emailRecoveryRateLimitMax = int64(5)
)

var recoveryCodeAlphabet = []byte("ABCDEFGHJKLMNPQRSTUVWXYZ23456789")

type pendingTOTPSetup struct {
	SetupID     string                `json:"setup_id"`
	UserID      int64                 `json:"user_id"`
	Email       string                `json:"email"`
	Issuer      string                `json:"issuer"`
	AccountName string                `json:"account_name"`
	Secret      helper.EncryptedValue `json:"secret"`
	ExpiresAt   time.Time             `json:"expires_at"`
}

type loginChallengeState struct {
	ChallengeID            string     `json:"challenge_id"`
	UserID                 int64      `json:"user_id"`
	FactorID               int64      `json:"factor_id"`
	FactorType             string     `json:"factor_type"`
	Email                  string     `json:"email"`
	Methods                []string   `json:"methods"`
	ExpiresAt              time.Time  `json:"expires_at"`
	AttemptCount           int        `json:"attempt_count"`
	EmailRecoveryCodeHash  string     `json:"email_recovery_code_hash,omitempty"`
	EmailRecoverySentAt    *time.Time `json:"email_recovery_sent_at,omitempty"`
	EmailRecoveryExpiresAt *time.Time `json:"email_recovery_expires_at,omitempty"`
}

type mfaRecoveryTemplateData struct {
	AppName        string
	Code           string
	ExpiresMinutes int
}

func loginChallengeMethods() []string {
	return []string{codeTypeTOTP, codeTypeRecoveryCode, codeTypeEmailRecovery}
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func hashKeyPart(value string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(value)))
	return hex.EncodeToString(sum[:])
}

func (e *entity) issueTokensForAuthentication(auth model.Authentication) (LoginResponse, error) {
	token, refresh, err := helper.IssueTokens(auth.UserId, auth.SessionVersion, e.secret)
	if err != nil {
		return LoginResponse{}, err
	}
	return LoginResponse{Token: token, RefreshToken: refresh}, nil
}

func (e *entity) syncSessionVersion(ctx context.Context, userID, version int64) error {
	if e.sessionChecker == nil {
		return nil
	}
	return e.sessionChecker.Sync(ctx, userID, version)
}

func (e *entity) publishAuthRevoked(ctx context.Context, userID, version int64) {
	if err := mq.SendUserUpdate(ctx, e.mqt, userID, &mqmsg.UserAuthRevoked{SessionVersion: version}); err != nil {
		observability.LoggerWithContext(ctx, e.log).Error("unable to publish auth revocation", "error", err.Error(), "user_id", userID)
	}
}

func (e *entity) completeSessionMutation(ctx context.Context, userID, version int64) error {
	if err := e.syncSessionVersion(ctx, userID, version); err != nil {
		return err
	}
	e.publishAuthRevoked(ctx, userID, version)
	return nil
}

func (e *entity) loadActiveFactorBundle(ctx context.Context, userID int64) (*model.AuthFactor, *model.AuthTOTPFactor, error) {
	factor, err := e.factor.GetActiveFactor(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	if factor.FactorType != authFactorTypeTOTP {
		return &factor, nil, nil
	}
	totpFactor, err := e.factor.GetTOTPFactor(ctx, userID, factor.FactorID)
	if err != nil {
		return nil, nil, err
	}
	return &factor, &totpFactor, nil
}

func (e *entity) decryptTOTPSecret(totpFactor model.AuthTOTPFactor) (string, error) {
	if e.secretBox == nil {
		return "", errors.New(ErrUnableToDecryptSecret)
	}
	secret, err := e.secretBox.Decrypt(helper.EncryptedValue{
		Nonce:      totpFactor.SecretNonce,
		Ciphertext: totpFactor.SecretCiphertext,
	})
	if err != nil {
		return "", err
	}
	return string(secret), nil
}

func (e *entity) validateCurrentPassword(auth model.Authentication, password string) error {
	if err := CompareHashAndPassword(auth.PasswordHash, password); err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, ErrUnableToCompareHash)
	}
	return nil
}

func (e *entity) generateRecoveryCodes() ([]string, []model.AuthRecoveryCode, error) {
	codes := make([]string, 10)
	stored := make([]model.AuthRecoveryCode, 10)
	now := time.Now()
	for i := range codes {
		code, err := randomFromAlphabet(12, recoveryCodeAlphabet)
		if err != nil {
			return nil, nil, err
		}
		hash, err := HashPassword(code)
		if err != nil {
			return nil, nil, err
		}
		codes[i] = code
		stored[i] = model.AuthRecoveryCode{
			CodeID:    e.idGenerator(),
			CodeHash:  hash,
			CreatedAt: now,
		}
	}
	return codes, stored, nil
}

func randomFromAlphabet(length int, alphabet []byte) (string, error) {
	if length <= 0 {
		return "", nil
	}
	buf := make([]byte, length)
	rb := make([]byte, length)
	if _, err := rand.Read(rb); err != nil {
		return "", err
	}
	for i := range buf {
		buf[i] = alphabet[int(rb[i])%len(alphabet)]
	}
	return string(buf), nil
}

func randomDigits(length int) (string, error) {
	if length <= 0 {
		return "", nil
	}
	buf := make([]byte, length)
	rb := make([]byte, length)
	if _, err := rand.Read(rb); err != nil {
		return "", err
	}
	for i := range buf {
		buf[i] = byte('0' + (rb[i] % 10))
	}
	return string(buf), nil
}

func (e *entity) verifySecurityCode(ctx context.Context, tx *sqlx.Tx, userID int64, factor model.AuthFactor, codeType, code string) error {
	switch codeType {
	case codeTypeTOTP:
		totpFactor, err := e.factor.GetTOTPFactor(ctx, userID, factor.FactorID)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetTOTPFactor)
		}
		secret, err := e.decryptTOTPSecret(totpFactor)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToDecryptSecret)
		}
		valid, err := totp.ValidateCustom(code, secret, time.Now(), totp.ValidateOpts{
			Period:    uint(totpFactor.PeriodSeconds),
			Skew:      1,
			Digits:    otp.Digits(totpFactor.Digits),
			Algorithm: otp.AlgorithmSHA1,
		})
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToValidateSession)
		}
		if !valid {
			return fiber.NewError(fiber.StatusUnauthorized, ErrInvalidTwoFactorCode)
		}
		if err := e.factor.TouchFactorLastUsedTx(ctx, tx, userID, factor.FactorID, time.Now()); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToValidateSession)
		}
		return nil

	case codeTypeRecoveryCode:
		recoveryCodes, err := e.factor.ListUnusedRecoveryCodes(ctx, userID, factor.FactorID)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToListRecoveryCodes)
		}
		for _, recoveryCode := range recoveryCodes {
			if err := CompareHashAndPassword(recoveryCode.CodeHash, code); err != nil {
				continue
			}
			consumed, err := e.factor.ConsumeRecoveryCodeTx(ctx, tx, userID, recoveryCode.CodeID, time.Now())
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToConsumeRecoveryCode)
			}
			if !consumed {
				break
			}
			if err := e.factor.TouchFactorLastUsedTx(ctx, tx, userID, factor.FactorID, time.Now()); err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToValidateSession)
			}
			return nil
		}
		return fiber.NewError(fiber.StatusUnauthorized, ErrInvalidRecoveryCode)

	default:
		return fiber.NewError(fiber.StatusBadRequest, ErrCodeTypeInvalid)
	}
}

func (e *entity) rateLimit(ctx context.Context, key string, limit int64, ttl time.Duration, publicMessage string) error {
	if e.cache == nil {
		return nil
	}
	count, err := e.cache.Incr(ctx, key)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to access rate limiter")
	}
	if count == 1 {
		_ = e.cache.SetTTL(ctx, key, int64(ttl.Seconds()))
	}
	if count > limit {
		return fiber.NewError(fiber.StatusTooManyRequests, publicMessage)
	}
	return nil
}

func (e *entity) loginRateLimit(ctx context.Context, email, ip string) error {
	return e.rateLimit(ctx, fmt.Sprintf("auth:rl:login:%s:%s", hashKeyPart(normalizeEmail(email)), hashKeyPart(ip)), loginRateLimitMax, loginRateLimitTTL, ErrTooManyAuthAttempts)
}

func (e *entity) challengeRateLimit(ctx context.Context, challengeID, ip string) error {
	return e.rateLimit(ctx, fmt.Sprintf("auth:rl:challenge:%s:%s", hashKeyPart(challengeID), hashKeyPart(ip)), challengeRateLimitMax, challengeRateLimitTTL, ErrTooManyAuthAttempts)
}

func (e *entity) emailRecoveryRateLimit(ctx context.Context, userID int64, ip string) error {
	return e.rateLimit(ctx, fmt.Sprintf("auth:rl:email-recovery:%d:%s", userID, hashKeyPart(ip)), emailRecoveryRateLimitMax, emailRecoveryRateLimitTTL, ErrTooManyRecoveryEmails)
}

func loginChallengeKey(challengeID string) string {
	return "auth:login_challenge:" + challengeID
}

func pendingTOTPSetupKey(setupID string) string {
	return "auth:totp_setup:" + setupID
}

func ttlSeconds(expiresAt time.Time) int64 {
	ttl := time.Until(expiresAt)
	if ttl < time.Second {
		return 1
	}
	return int64(ttl.Seconds())
}

func (e *entity) saveLoginChallenge(ctx context.Context, state loginChallengeState) error {
	return e.cache.SetTimedJSON(ctx, loginChallengeKey(state.ChallengeID), state, ttlSeconds(state.ExpiresAt), cache.NoneProactive())
}

func (e *entity) loadLoginChallenge(ctx context.Context, challengeID string) (loginChallengeState, error) {
	var state loginChallengeState
	if err := e.cache.GetJSON(ctx, loginChallengeKey(challengeID), &state); err != nil {
		return loginChallengeState{}, err
	}
	return state, nil
}

func (e *entity) deleteLoginChallenge(ctx context.Context, challengeID string) error {
	return e.cache.Delete(ctx, loginChallengeKey(challengeID))
}

func (e *entity) savePendingTOTPSetup(ctx context.Context, state pendingTOTPSetup) error {
	return e.cache.SetTimedJSON(ctx, pendingTOTPSetupKey(state.SetupID), state, ttlSeconds(state.ExpiresAt), cache.NoneProactive())
}

func (e *entity) loadPendingTOTPSetup(ctx context.Context, setupID string) (pendingTOTPSetup, error) {
	var state pendingTOTPSetup
	if err := e.cache.GetJSON(ctx, pendingTOTPSetupKey(setupID), &state); err != nil {
		return pendingTOTPSetup{}, err
	}
	return state, nil
}

func (e *entity) deletePendingTOTPSetup(ctx context.Context, setupID string) error {
	return e.cache.Delete(ctx, pendingTOTPSetupKey(setupID))
}

func (e *entity) createLoginChallenge(ctx context.Context, auth model.Authentication, factor model.AuthFactor) (LoginChallengeResponse, error) {
	challengeID, err := helper.RandomToken(40)
	if err != nil {
		return LoginChallengeResponse{}, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGenerateToken)
	}
	expiresAt := time.Now().Add(loginChallengeTTL)
	state := loginChallengeState{
		ChallengeID: challengeID,
		UserID:      auth.UserId,
		FactorID:    factor.FactorID,
		FactorType:  factor.FactorType,
		Email:       auth.Email,
		Methods:     loginChallengeMethods(),
		ExpiresAt:   expiresAt,
	}
	if err := e.saveLoginChallenge(ctx, state); err != nil {
		return LoginChallengeResponse{}, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateLoginChallenge)
	}
	return LoginChallengeResponse{ChallengeID: challengeID, Methods: state.Methods, ExpiresAt: expiresAt}, nil
}

func (e *entity) incrementChallengeAttempts(ctx context.Context, state loginChallengeState) error {
	state.AttemptCount++
	if state.AttemptCount >= int(challengeRateLimitMax) {
		_ = e.deleteLoginChallenge(ctx, state.ChallengeID)
		return fiber.NewError(fiber.StatusTooManyRequests, ErrTooManyAuthAttempts)
	}
	if err := e.saveLoginChallenge(ctx, state); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateLoginChallenge)
	}
	return nil
}

func (e *entity) sendEmailRecoveryCode(ctx context.Context, email, code string) error {
	return e.mailer.SendTemplate(ctx, mailer.User{Email: email}, e.appName+" login recovery code", "mfa_recovery", mfaRecoveryTemplateData{
		AppName:        e.appName,
		Code:           code,
		ExpiresMinutes: int(emailRecoveryCodeTTL / time.Minute),
	})
}
