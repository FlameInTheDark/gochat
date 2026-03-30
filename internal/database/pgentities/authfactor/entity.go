package authfactor

import (
	"context"
	"time"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/jmoiron/sqlx"
)

type AuthFactor interface {
	GetActiveFactor(ctx context.Context, userID int64) (model.AuthFactor, error)
	GetTOTPFactor(ctx context.Context, userID, factorID int64) (model.AuthTOTPFactor, error)
	ListUnusedRecoveryCodes(ctx context.Context, userID, factorID int64) ([]model.AuthRecoveryCode, error)
	CountUnusedRecoveryCodes(ctx context.Context, userID, factorID int64) (int, error)
	CreateTOTPFactorTx(ctx context.Context, tx *sqlx.Tx, factor model.AuthFactor, totp model.AuthTOTPFactor) error
	ReplaceRecoveryCodesTx(ctx context.Context, tx *sqlx.Tx, userID, factorID int64, codes []model.AuthRecoveryCode) error
	TouchFactorLastUsedTx(ctx context.Context, tx *sqlx.Tx, userID, factorID int64, usedAt time.Time) error
	ConsumeRecoveryCodeTx(ctx context.Context, tx *sqlx.Tx, userID, codeID int64, usedAt time.Time) (bool, error)
	DeleteFactorsTx(ctx context.Context, tx *sqlx.Tx, userID int64) error
}

type Entity struct {
	c *sqlx.DB
}

func New(c *sqlx.DB) AuthFactor {
	return &Entity{c: c}
}
