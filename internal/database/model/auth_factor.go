package model

import "time"

type AuthFactor struct {
	UserID      int64      `db:"user_id"`
	FactorID    int64      `db:"factor_id"`
	FactorType  string     `db:"factor_type"`
	DisplayName string     `db:"display_name"`
	Status      string     `db:"status"`
	CreatedAt   time.Time  `db:"created_at"`
	VerifiedAt  *time.Time `db:"verified_at"`
	LastUsedAt  *time.Time `db:"last_used_at"`
}

type AuthTOTPFactor struct {
	UserID           int64     `db:"user_id"`
	FactorID         int64     `db:"factor_id"`
	SecretCiphertext []byte    `db:"secret_ciphertext"`
	SecretNonce      []byte    `db:"secret_nonce"`
	Algorithm        string    `db:"algorithm"`
	Digits           int       `db:"digits"`
	PeriodSeconds    int       `db:"period_seconds"`
	CreatedAt        time.Time `db:"created_at"`
}

type AuthRecoveryCode struct {
	UserID    int64      `db:"user_id"`
	FactorID  int64      `db:"factor_id"`
	CodeID    int64      `db:"code_id"`
	CodeHash  string     `db:"code_hash"`
	UsedAt    *time.Time `db:"used_at"`
	CreatedAt time.Time  `db:"created_at"`
}
