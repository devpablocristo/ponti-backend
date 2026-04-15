package stock

import (
	"context"

	shareddb "github.com/devpablocristo/ponti-backend/internal/shared/db"
	"gorm.io/gorm"
)

func (r *Repository) getDB(ctx context.Context) *gorm.DB {
	if tx := shareddb.TxFromContext(ctx); tx != nil {
		return tx.WithContext(ctx)
	}
	return r.db.Client().WithContext(ctx)
}

func (r *Repository) ExecuteInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(shareddb.WithTx(ctx, tx))
	})
}
