package authfactor

import (
	"context"
	"fmt"
	"time"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

func (e *Entity) GetActiveFactor(ctx context.Context, userID int64) (model.AuthFactor, error) {
	var factor model.AuthFactor
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("auth_factors").
		Where(squirrel.Eq{"user_id": userID, "status": "active"}).
		OrderBy("created_at DESC").
		Limit(1)
	sql, args, err := q.ToSql()
	if err != nil {
		return model.AuthFactor{}, fmt.Errorf("unable to create SQL query: %w", err)
	}
	if err := e.c.GetContext(ctx, &factor, sql, args...); err != nil {
		return factor, fmt.Errorf("unable to get active auth factor: %w", err)
	}
	return factor, nil
}

func (e *Entity) GetTOTPFactor(ctx context.Context, userID, factorID int64) (model.AuthTOTPFactor, error) {
	var factor model.AuthTOTPFactor
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("auth_totp_factors").
		Where(squirrel.Eq{"user_id": userID, "factor_id": factorID}).
		Limit(1)
	sql, args, err := q.ToSql()
	if err != nil {
		return model.AuthTOTPFactor{}, fmt.Errorf("unable to create SQL query: %w", err)
	}
	if err := e.c.GetContext(ctx, &factor, sql, args...); err != nil {
		return factor, fmt.Errorf("unable to get TOTP factor: %w", err)
	}
	return factor, nil
}

func (e *Entity) ListUnusedRecoveryCodes(ctx context.Context, userID, factorID int64) ([]model.AuthRecoveryCode, error) {
	var codes []model.AuthRecoveryCode
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("auth_recovery_codes").
		Where(squirrel.Eq{"user_id": userID, "factor_id": factorID, "used_at": nil}).
		OrderBy("created_at ASC")
	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("unable to create SQL query: %w", err)
	}
	if err := e.c.SelectContext(ctx, &codes, sql, args...); err != nil {
		return nil, fmt.Errorf("unable to list recovery codes: %w", err)
	}
	return codes, nil
}

func (e *Entity) CountUnusedRecoveryCodes(ctx context.Context, userID, factorID int64) (int, error) {
	var count int
	q := squirrel.Select("COUNT(*)").
		PlaceholderFormat(squirrel.Dollar).
		From("auth_recovery_codes").
		Where(squirrel.Eq{"user_id": userID, "factor_id": factorID, "used_at": nil})
	sql, args, err := q.ToSql()
	if err != nil {
		return 0, fmt.Errorf("unable to create SQL query: %w", err)
	}
	if err := e.c.GetContext(ctx, &count, sql, args...); err != nil {
		return 0, fmt.Errorf("unable to count recovery codes: %w", err)
	}
	return count, nil
}

func (e *Entity) CreateTOTPFactorTx(ctx context.Context, tx *sqlx.Tx, factor model.AuthFactor, totp model.AuthTOTPFactor) error {
	q := squirrel.Insert("auth_factors").
		PlaceholderFormat(squirrel.Dollar).
		Columns("user_id", "factor_id", "factor_type", "display_name", "status", "created_at", "verified_at", "last_used_at").
		Values(factor.UserID, factor.FactorID, factor.FactorType, factor.DisplayName, factor.Status, factor.CreatedAt, factor.VerifiedAt, factor.LastUsedAt)
	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	if _, err := tx.ExecContext(ctx, sql, args...); err != nil {
		return fmt.Errorf("unable to create auth factor: %w", err)
	}

	q = squirrel.Insert("auth_totp_factors").
		PlaceholderFormat(squirrel.Dollar).
		Columns("user_id", "factor_id", "secret_ciphertext", "secret_nonce", "algorithm", "digits", "period_seconds", "created_at").
		Values(totp.UserID, totp.FactorID, totp.SecretCiphertext, totp.SecretNonce, totp.Algorithm, totp.Digits, totp.PeriodSeconds, totp.CreatedAt)
	sql, args, err = q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	if _, err := tx.ExecContext(ctx, sql, args...); err != nil {
		return fmt.Errorf("unable to create TOTP factor: %w", err)
	}

	return nil
}

func (e *Entity) ReplaceRecoveryCodesTx(ctx context.Context, tx *sqlx.Tx, userID, factorID int64, codes []model.AuthRecoveryCode) error {
	deleteQuery := squirrel.Delete("auth_recovery_codes").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"user_id": userID, "factor_id": factorID})
	sql, args, err := deleteQuery.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	if _, err := tx.ExecContext(ctx, sql, args...); err != nil {
		return fmt.Errorf("unable to delete recovery codes: %w", err)
	}

	if len(codes) == 0 {
		return nil
	}

	insertQuery := squirrel.Insert("auth_recovery_codes").
		PlaceholderFormat(squirrel.Dollar).
		Columns("user_id", "factor_id", "code_id", "code_hash", "used_at", "created_at")
	for _, code := range codes {
		insertQuery = insertQuery.Values(code.UserID, code.FactorID, code.CodeID, code.CodeHash, code.UsedAt, code.CreatedAt)
	}
	sql, args, err = insertQuery.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	if _, err := tx.ExecContext(ctx, sql, args...); err != nil {
		return fmt.Errorf("unable to insert recovery codes: %w", err)
	}
	return nil
}

func (e *Entity) TouchFactorLastUsedTx(ctx context.Context, tx *sqlx.Tx, userID, factorID int64, usedAt time.Time) error {
	q := squirrel.Update("auth_factors").
		PlaceholderFormat(squirrel.Dollar).
		Set("last_used_at", usedAt).
		Where(squirrel.Eq{"user_id": userID, "factor_id": factorID})
	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	if _, err := tx.ExecContext(ctx, sql, args...); err != nil {
		return fmt.Errorf("unable to update factor last used at: %w", err)
	}
	return nil
}

func (e *Entity) ConsumeRecoveryCodeTx(ctx context.Context, tx *sqlx.Tx, userID, codeID int64, usedAt time.Time) (bool, error) {
	q := squirrel.Update("auth_recovery_codes").
		PlaceholderFormat(squirrel.Dollar).
		Set("used_at", usedAt).
		Where(squirrel.Eq{"user_id": userID, "code_id": codeID, "used_at": nil})
	sql, args, err := q.ToSql()
	if err != nil {
		return false, fmt.Errorf("unable to create SQL query: %w", err)
	}
	res, err := tx.ExecContext(ctx, sql, args...)
	if err != nil {
		return false, fmt.Errorf("unable to consume recovery code: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("unable to inspect recovery code update: %w", err)
	}
	return rows > 0, nil
}

func (e *Entity) DeleteFactorsTx(ctx context.Context, tx *sqlx.Tx, userID int64) error {
	for _, table := range []string{"auth_recovery_codes", "auth_totp_factors", "auth_factors"} {
		q := squirrel.Delete(table).
			PlaceholderFormat(squirrel.Dollar).
			Where(squirrel.Eq{"user_id": userID})
		sql, args, err := q.ToSql()
		if err != nil {
			return fmt.Errorf("unable to create SQL query: %w", err)
		}
		if _, err := tx.ExecContext(ctx, sql, args...); err != nil {
			return fmt.Errorf("unable to delete auth factor data from %s: %w", table, err)
		}
	}
	return nil
}
