// Package stock implementa repositorios para stock.
package stock

import (
	"context"
	"errors"
	"fmt"
	"time"

	reportdb "github.com/alphacodinggroup/ponti-backend/internal/shared/db"
	models "github.com/alphacodinggroup/ponti-backend/internal/stock/repository/models"
	"github.com/alphacodinggroup/ponti-backend/internal/stock/usecases/domain"
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// GormEnginePort expone el cliente de base de datos requerido.
type GormEnginePort interface {
	Client() *gorm.DB
}

// Repository implementa el acceso a datos de stock.
type Repository struct {
	db GormEnginePort
}

// NewRepository crea una nueva instancia del repositorio de Stock.
func NewRepository(db GormEnginePort) *Repository {
	return &Repository{db: db}
}

// GetStocks retorna stocks filtrando por proyecto y opcionalmente por fecha de corte.
func (r *Repository) GetStocks(ctx context.Context, projectID int64, closeDate time.Time) ([]*domain.Stock, error) {
	gormDB := r.getDB(ctx)
	var t time.Time

	query := gormDB.Model(&models.Stock{}).
		Preload("Project").
		Preload("Supply", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("Supply.Type").
		Preload("Supply.Category").
		Preload("Investor").
		Preload("SupplyMovements").
		Joins("JOIN projects ON projects.id = stocks.project_id").
		Where("projects.id = ?", projectID)

	if closeDate != t {
		query = query.Where("stocks.close_date = ?", closeDate)
	} else {
		// Si no se especifica fecha, obtener el stock activo (sin filtro de fecha)
		query = query.Where("stocks.close_date IS NULL")
	}

	var stockModels []models.Stock
	if err := query.Order("stocks.id DESC").Find(&stockModels).Error; err != nil {
		return nil, err
	}

	// OPTIMIZACIÓN: Calcular consumed en una sola consulta para evitar N+1 problem
	if len(stockModels) > 0 {
		var consumedResults []struct {
			SupplyID int64           `gorm:"column:supply_id"`
			Consumed decimal.Decimal `gorm:"column:consumed"`
		}

		// Obtener todos los supply_ids
		supplyIDs := make([]int64, len(stockModels))
		for i, stock := range stockModels {
			supplyIDs[i] = stock.SupplyID
		}

		// Calcular consumed para todos los supplies en una sola consulta desde vista
		query := fmt.Sprintf(`
			SELECT supply_id, consumed
			FROM %s
			WHERE project_id = ? AND supply_id IN ?
		`, reportdb.ReportView("stock_consumed_by_supply"))
		err := gormDB.Raw(query, projectID, supplyIDs).Scan(&consumedResults).Error
		if err != nil {
			return nil, err
		}

		// Crear mapa de consumed por supply_id
		consumedMap := make(map[int64]decimal.Decimal)
		for _, result := range consumedResults {
			consumedMap[result.SupplyID] = result.Consumed
		}

		// Asignar consumed a cada stock
		for i := range stockModels {
			if consumed, exists := consumedMap[stockModels[i].SupplyID]; exists {
				stockModels[i].Consumed = consumed
			} else {
				stockModels[i].Consumed = decimal.Zero
			}
		}
	}

	stocks := make([]*domain.Stock, 0, len(stockModels))
	for i := range stockModels {
		stocks = append(stocks, stockModels[i].ToDomain())
	}
	return stocks, nil
}

func (r *Repository) GetStocksPeriods(ctx context.Context, projectID int64) ([]string, error) {
	var rawPeriods []time.Time

	err := r.getDB(ctx).
		Model(&models.Stock{}).
		Where("project_id = ? AND close_date IS NOT NULL", projectID).
		Distinct("close_date").
		Pluck("close_date", &rawPeriods).Error
	if err != nil {
		return nil, err
	}

	periods := make([]string, len(rawPeriods))
	for i, t := range rawPeriods {
		periods[i] = t.Format("2006-01-02")
	}

	return periods, nil
}

func (r *Repository) CreateStock(ctx context.Context, stock *domain.Stock) (int64, error) {
	if stock == nil {
		return 0, types.NewError(types.ErrBadRequest, "stock is nil", nil)
	}
	model := models.FromDomain(stock)
	if err := r.getDB(ctx).Create(model).Error; err != nil {
		return 0, types.NewError(types.ErrInternal, "failed to create stock", err)
	}
	return model.ID, nil
}

func (r *Repository) UpdateCloseDateByProject(ctx context.Context, projectID int64, stock *domain.Stock) error {
	stockUpdate := models.StockUpdateCloseDateFromDomain(stock)
	result := r.getDB(ctx).
		Model(&models.Stock{}).
		Where("project_id = ?", projectID).
		Where("close_date IS NULL").
		Updates(stockUpdate)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return types.NewError(types.ErrNotFound, "no stocks found to update", nil)
	}
	return nil
}

func (r *Repository) UpdateRealStockUnits(ctx context.Context, stockID int64, stock *domain.Stock) error {
	stockUpdate := models.StockUpdateRealUnitsFromDomain(stock)
	updateTx := r.getDB(ctx).
		Model(&models.Stock{}).
		Where("id = ?", stockID)
	if stock != nil && !stock.UpdatedAt.IsZero() {
		updateTx = updateTx.Where("updated_at = ?", stock.UpdatedAt)
	}
	result := updateTx.Updates(stockUpdate)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		if stock != nil && !stock.UpdatedAt.IsZero() {
			return types.NewError(types.ErrConflict, "stock not found or outdated", nil)
		}
		return types.NewError(types.ErrNotFound, "no stock found to update", nil)
	}
	return nil
}

func (r *Repository) UpdateUnitsConsumed(ctx context.Context, stockDomain domain.Stock, quantity decimal.Decimal) error {
	updateTx := r.getDB(ctx).
		Model(&models.Stock{}).
		Where("id = ?", stockDomain.ID)
	if !stockDomain.UpdatedAt.IsZero() {
		updateTx = updateTx.Where("updated_at = ?", stockDomain.UpdatedAt)
	}
	result := updateTx.Update("units_consumed", gorm.Expr("units_consumed + ?", quantity))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		if !stockDomain.UpdatedAt.IsZero() {
			return types.NewError(types.ErrConflict, "stock not found or outdated", nil)
		}
		return types.NewError(types.ErrNotFound, "no stock found to update", nil)
	}
	return nil
}

func (r *Repository) GetStockByID(ctx context.Context, stockID int64) (*domain.Stock, error) {
	var stockModel models.Stock
	err := r.getDB(ctx).
		Preload("Project").
		Preload("Supply", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("Supply.Type").
		Preload("Supply.Category").
		Preload("Investor").
		First(&stockModel, stockID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.NewError(types.ErrNotFound, "stock not found", nil)
		}
		return nil, types.NewError(types.ErrInternal, "failed to get stock", err)
	}
	return stockModel.ToDomain(), nil
}

func (r *Repository) GetLastStockByProjectID(ctx context.Context, projectID int64, supplyID int64) (*domain.Stock, bool, error) {
	var stockModel models.Stock
	gormDB := r.getDB(ctx)
	err := gormDB.
		Preload("Project").
		Preload("Supply", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("Supply.Type").
		Preload("Supply.Category").
		Preload("Investor").
		Preload("SupplyMovements").
		Where("project_id = ?", projectID).
		Where("supply_id = ?", supplyID).
		Where("close_date is null").
		First(&stockModel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, true, nil
		}

		return nil, false, types.NewError(types.ErrInternal, "failed to get last stock", err)
	}

	// Para validaciones (p.ej. movimientos internos) necesitamos el stock de sistema,
	// que depende de `consumed`. Lo resolvemos desde la vista de reportes.
	var consumedResult struct {
		Consumed decimal.Decimal `gorm:"column:consumed"`
	}
	query := fmt.Sprintf(`
		SELECT consumed
		FROM %s
		WHERE project_id = ? AND supply_id = ?
	`, reportdb.ReportView("stock_consumed_by_supply"))
	if err := gormDB.Raw(query, projectID, supplyID).Scan(&consumedResult).Error; err != nil {
		return nil, false, types.NewError(types.ErrInternal, "failed to load stock consumed", err)
	}
	stockModel.Consumed = consumedResult.Consumed

	return stockModel.ToDomain(), false, nil
}

func (r *Repository) GetStockByPeriodAndProjectID(ctx context.Context, projectID int64) (*domain.Stock, error) {
	var stockModel models.Stock

	err := r.getDB(ctx).
		Preload("Project").
		Preload("Supply", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("Supply.Type").
		Preload("Supply.Category").
		Preload("Investor").
		Where("project_id = ?", projectID).
		Where("close_date IS NULL").
		First(&stockModel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.NewError(types.ErrNotFound, "stock not found", nil)
		}
		return nil, types.NewError(types.ErrInternal, "failed to get stock by period", err)
	}

	return stockModel.ToDomain(), nil
}

func (r *Repository) ListAllStocks(ctx context.Context) ([]*domain.Stock, error) {
	var stockModel []models.Stock

	gormDB := r.getDB(ctx)

	query := gormDB.
		Preload("Project").
		Preload("Supply", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("Supply.Type").
		Preload("Supply.Category").
		Preload("Investor").
		Preload("SupplyMovements")

	if err := query.Find(&stockModel).Error; err != nil {
		return nil, err
	}

	// Calcular consumed para todos los stocks agrupando por project_id + supply_id
	if len(stockModel) > 0 {
		var consumedResults []struct {
			ProjectID int64           `gorm:"column:project_id"`
			SupplyID  int64           `gorm:"column:supply_id"`
			Consumed  decimal.Decimal `gorm:"column:consumed"`
		}

		// Calcular consumed agrupado por project_id y supply_id desde vista
		query := fmt.Sprintf(`
			SELECT project_id, supply_id, consumed
			FROM %s
		`, reportdb.ReportView("stock_consumed_by_supply"))
		err := gormDB.Raw(query).Scan(&consumedResults).Error
		if err != nil {
			return nil, err
		}

		// Crear mapa de consumed por project_id:supply_id
		type stockKey struct {
			ProjectID int64
			SupplyID  int64
		}
		consumedMap := make(map[stockKey]decimal.Decimal)
		for _, result := range consumedResults {
			key := stockKey{ProjectID: result.ProjectID, SupplyID: result.SupplyID}
			consumedMap[key] = result.Consumed
		}

		// Asignar consumed a cada stock
		for i := range stockModel {
			key := stockKey{ProjectID: stockModel[i].ProjectID, SupplyID: stockModel[i].SupplyID}
			if consumed, exists := consumedMap[key]; exists {
				stockModel[i].Consumed = consumed
			} else {
				stockModel[i].Consumed = decimal.Zero
			}
		}
	}

	out := make([]*domain.Stock, 0, len(stockModel))
	for i := range stockModel {
		out = append(out, stockModel[i].ToDomain())
	}
	return out, nil
}
