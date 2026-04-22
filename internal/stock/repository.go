// Package stock implementa repositorios para stock continuo por proyecto.
package stock

import (
	"context"
	"fmt"
	"time"

	"github.com/devpablocristo/core/errors/go/domainerr"
	shareddb "github.com/devpablocristo/ponti-backend/internal/shared/db"
	models "github.com/devpablocristo/ponti-backend/internal/stock/repository/models"
	"github.com/devpablocristo/ponti-backend/internal/stock/usecases/domain"
	"gorm.io/gorm"
)

// GormEnginePort expone el cliente de base de datos requerido.
type GormEnginePort interface {
	Client() *gorm.DB
}

// Repository implementa el acceso a datos del stock canónico.
type Repository struct {
	db GormEnginePort
}

// NewRepository crea una nueva instancia del repositorio de Stock.
func NewRepository(db GormEnginePort) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetStocks(ctx context.Context, projectID int64, cutoffDate time.Time) ([]*domain.Stock, error) {
	var rows []models.StockSummaryRow
	query := `
		SELECT *
		FROM v4_report.stock_summary_for_project_as_of(?, ?)
		ORDER BY supply_name ASC, supply_id ASC
	`

	if err := r.getDB(ctx).Raw(query, projectID, cutoffDateParam(cutoffDate)).Scan(&rows).Error; err != nil {
		return nil, domainerr.Internal("failed to list stock summary")
	}

	out := make([]*domain.Stock, 0, len(rows))
	for i := range rows {
		out = append(out, rows[i].ToDomain())
	}
	return out, nil
}

func (r *Repository) GetStockBySupplyID(
	ctx context.Context,
	projectID int64,
	supplyID int64,
	cutoffDate time.Time,
) (*domain.Stock, error) {
	var row models.StockSummaryRow
	query := `
		SELECT *
		FROM v4_report.stock_summary_for_project_as_of(?, ?)
		WHERE supply_id = ?
		LIMIT 1
	`

	if err := r.getDB(ctx).Raw(query, projectID, cutoffDateParam(cutoffDate), supplyID).Scan(&row).Error; err != nil {
		return nil, domainerr.Internal("failed to get stock summary")
	}
	if row.SupplyID == 0 {
		return nil, domainerr.NotFound("stock summary not found")
	}
	return row.ToDomain(), nil
}

func (r *Repository) CreateStockCount(ctx context.Context, count *domain.StockCount) (int64, error) {
	if count == nil {
		return 0, domainerr.Validation("stock count is nil")
	}

	model := models.StockCountFromDomain(count)
	if err := r.getDB(ctx).Create(model).Error; err != nil {
		return 0, domainerr.Internal("failed to create stock count")
	}
	return model.ID, nil
}

func (r *Repository) getDB(ctx context.Context) *gorm.DB {
	if tx := shareddb.TxFromContext(ctx); tx != nil {
		return tx.WithContext(ctx)
	}
	return r.db.Client().WithContext(ctx)
}

func cutoffDateParam(cutoffDate time.Time) any {
	if cutoffDate.IsZero() {
		return nil
	}
	return cutoffDate.Format("2006-01-02")
}

func (r *Repository) ExecuteInTransaction(ctx context.Context, fn func(context.Context) error) error {
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := shareddb.WithTx(ctx, tx)
		if err := fn(txCtx); err != nil {
			return fmt.Errorf("transaction failed: %w", err)
		}
		return nil
	})
}
