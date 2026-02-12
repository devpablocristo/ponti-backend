package db

import (
	"context"

	"gorm.io/gorm"
)

type txContextKey struct{}

// WithTx attaches a gorm transaction to context so repositories can reuse it.
func WithTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txContextKey{}, tx)
}

// TxFromContext returns a transaction from context when available.
func TxFromContext(ctx context.Context) *gorm.DB {
	tx, _ := ctx.Value(txContextKey{}).(*gorm.DB)
	return tx
}
