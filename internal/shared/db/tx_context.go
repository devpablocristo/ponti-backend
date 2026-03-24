package db

import (
	"context"

	gormdb "github.com/devpablocristo/core/databases/gorm/go"
	"gorm.io/gorm"
)

// WithTx delega al helper estándar de core.
func WithTx(ctx context.Context, tx *gorm.DB) context.Context {
	return gormdb.WithTx(ctx, tx)
}

// TxFromContext delega al helper estándar de core.
func TxFromContext(ctx context.Context) *gorm.DB {
	return gormdb.TxFromContext(ctx)
}
