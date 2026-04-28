// Package stock implementa repositorios para stock.
package stock

import (
	"context"
	"errors"
	"time"

	"github.com/devpablocristo/core/errors/go/domainerr"
	models "github.com/devpablocristo/ponti-backend/internal/stock/repository/models"
	"github.com/devpablocristo/ponti-backend/internal/stock/usecases/domain"
	supplymodels "github.com/devpablocristo/ponti-backend/internal/supply/repository/models"
	supplymovementdomain "github.com/devpablocristo/ponti-backend/internal/supply/usecases/domain"
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

type stockKey struct {
	ProjectID  int64
	SupplyID   int64
	InvestorID int64
}

// NewRepository crea una nueva instancia del repositorio de Stock.
func NewRepository(db GormEnginePort) *Repository {
	return &Repository{db: db}
}

// GetStocks retorna stocks filtrando por proyecto y opcionalmente por fecha de corte.
func (r *Repository) GetStocks(ctx context.Context, projectID int64, closeDate time.Time) ([]*domain.Stock, error) {
	gormDB := r.getDB(ctx)

	stockQuery := gormDB.Model(&models.Stock{}).
		Preload("Project").
		Preload("Supply", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("Supply.Type").
		Preload("Supply.Category").
		Preload("Investor").
		Preload("SupplyMovements").
		Where("stocks.project_id = ?", projectID)

	if closeDate.IsZero() {
		stockQuery = stockQuery.Where("stocks.close_date IS NULL")
	} else {
		stockQuery = stockQuery.Where("stocks.close_date = ?", closeDate)
	}

	var stockModels []models.Stock
	if err := stockQuery.Find(&stockModels).Error; err != nil {
		return nil, err
	}

	var supplies []supplymodels.Supply
	if err := gormDB.
		Model(&supplymodels.Supply{}).
		Preload("Category").
		Preload("Type").
		Where("project_id = ?", projectID).
		Order("LOWER(name) ASC").
		Find(&supplies).Error; err != nil {
		return nil, err
	}

	movementsByKey, err := r.loadMovementsByStockKey(ctx, projectID)
	if err != nil {
		return nil, err
	}
	consumedByKey, consumedBySupplyID, err := r.loadConsumedByStockKey(ctx, projectID)
	if err != nil {
		return nil, err
	}

	stocksBySupplyID := make(map[int64][]models.Stock, len(stockModels))
	for i := range stockModels {
		key := keyFromStockModel(stockModels[i])
		if closeDate.IsZero() {
			stockModels[i].SupplyMovements = movementsByKey[key]
			stockModels[i].Consumed = consumedByKey[key]
		} else {
			stockModels[i].Consumed = consumedBySupplyID[stockModels[i].SupplyID]
		}
		stocksBySupplyID[stockModels[i].SupplyID] = append(stocksBySupplyID[stockModels[i].SupplyID], stockModels[i])
	}

	// Para supplies sin stock activo (close_date IS NULL), traer el último stock cerrado
	// con sus movimientos y tratarlo como activo. Así los ingresos ya cargados no se pierden
	// cuando el cierre de período deja pares (supply, investor) sin stock activo nuevo.
	if closeDate.IsZero() {
		missing := make([]int64, 0, len(supplies))
		for _, supply := range supplies {
			if len(stocksBySupplyID[supply.ID]) == 0 {
				missing = append(missing, supply.ID)
			}
		}

		if len(missing) > 0 {
			var latestClosedIDs []int64
			distinctQuery := `
				SELECT DISTINCT ON (supply_id, investor_id) id
				FROM stocks
				WHERE project_id = ? AND supply_id IN ? AND close_date IS NOT NULL
				ORDER BY supply_id, investor_id, close_date DESC, id DESC
			`
			if err := gormDB.Raw(distinctQuery, projectID, missing).Scan(&latestClosedIDs).Error; err != nil {
				return nil, err
			}

			if len(latestClosedIDs) > 0 {
				var latestClosed []models.Stock
				if err := gormDB.
					Preload("Project").
					Preload("Supply", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
					Preload("Supply.Type").
					Preload("Supply.Category").
					Preload("Investor").
					Preload("SupplyMovements").
					Where("id IN ?", latestClosedIDs).
					Find(&latestClosed).Error; err != nil {
					return nil, err
				}

				for _, stock := range latestClosed {
					key := keyFromStockModel(stock)
					stock.SupplyMovements = movementsByKey[key]
					stock.Consumed = consumedByKey[key]
					stock.CloseDate = nil // mostrar como activo en la UI
					stocksBySupplyID[stock.SupplyID] = append(stocksBySupplyID[stock.SupplyID], stock)
				}
			}
		}
	}

	stocks := make([]*domain.Stock, 0, len(supplies)+len(stockModels))
	for _, supply := range supplies {
		if stockModelsForSupply := stocksBySupplyID[supply.ID]; len(stockModelsForSupply) > 0 {
			stocks = append(stocks, mapStockModelsToDomain(stockModelsForSupply)...)
			continue
		}

		virtualStock := &domain.Stock{
			Supply:            supply.ToDomain(),
			SupplyMovements:   []supplymovementdomain.SupplyMovement{},
			RealStockUnits:    decimal.Zero,
			Consumed:          consumedBySupplyID[supply.ID],
			HasRealStockCount: false,
		}

		if !closeDate.IsZero() {
			cd := closeDate
			virtualStock.CloseDate = &cd
		}

		stocks = append(stocks, virtualStock)
	}

	return stocks, nil
}

func mapStockModelsToDomain(stockModels []models.Stock) []*domain.Stock {
	stocks := make([]*domain.Stock, 0, len(stockModels))
	for i := range stockModels {
		stocks = append(stocks, stockModels[i].ToDomain())
	}
	return stocks
}

func keyFromStockModel(stock models.Stock) stockKey {
	return stockKey{
		ProjectID:  stock.ProjectID,
		SupplyID:   stock.SupplyID,
		InvestorID: stock.InvestorID,
	}
}

func (r *Repository) loadMovementsByStockKey(ctx context.Context, projectID int64) (map[stockKey][]supplymodels.SupplyMovement, error) {
	var movements []supplymodels.SupplyMovement
	if err := r.getDB(ctx).
		Where("project_id = ?", projectID).
		Where("deleted_at IS NULL").
		Order("movement_date ASC, id ASC").
		Find(&movements).Error; err != nil {
		return nil, err
	}

	byKey := make(map[stockKey][]supplymodels.SupplyMovement)
	for _, movement := range movements {
		key := stockKey{
			ProjectID:  movement.ProjectId,
			SupplyID:   movement.SupplyID,
			InvestorID: movement.InvestorID,
		}
		byKey[key] = append(byKey[key], movement)
	}
	return byKey, nil
}

func (r *Repository) loadConsumedByStockKey(ctx context.Context, projectID int64) (map[stockKey]decimal.Decimal, map[int64]decimal.Decimal, error) {
	type consumedRow struct {
		SupplyID   int64           `gorm:"column:supply_id"`
		InvestorID int64           `gorm:"column:investor_id"`
		Consumed   decimal.Decimal `gorm:"column:consumed"`
	}

	var rows []consumedRow
	err := r.getDB(ctx).Raw(`
		SELECT
			woi.supply_id,
			wo.investor_id,
			COALESCE(SUM(woi.total_used), 0) AS consumed
		FROM workorder_items woi
		JOIN workorders wo ON wo.id = woi.workorder_id
		WHERE wo.project_id = ?
		  AND wo.deleted_at IS NULL
		  AND woi.deleted_at IS NULL
		  AND woi.supply_id IS NOT NULL
		GROUP BY woi.supply_id, wo.investor_id
	`, projectID).Scan(&rows).Error
	if err != nil {
		return nil, nil, err
	}

	byKey := make(map[stockKey]decimal.Decimal, len(rows))
	bySupply := make(map[int64]decimal.Decimal, len(rows))
	for _, row := range rows {
		key := stockKey{
			ProjectID:  projectID,
			SupplyID:   row.SupplyID,
			InvestorID: row.InvestorID,
		}
		byKey[key] = row.Consumed
		bySupply[row.SupplyID] = bySupply[row.SupplyID].Add(row.Consumed)
	}
	return byKey, bySupply, nil
}

// GetActiveStocksByProjectID retorna todos los stocks con close_date IS NULL
// de un proyecto, con sus relaciones precargadas. Se usa para replicar cada
// (supply, investor) activo al cerrar un período.
func (r *Repository) GetActiveStocksByProjectID(ctx context.Context, projectID int64) ([]*domain.Stock, error) {
	var stockModels []models.Stock
	if err := r.getDB(ctx).
		Preload("Project").
		Preload("Supply", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("Supply.Type").
		Preload("Supply.Category").
		Preload("Investor").
		Where("project_id = ? AND close_date IS NULL", projectID).
		Find(&stockModels).Error; err != nil {
		return nil, err
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
		return 0, domainerr.Validation("stock is nil")
	}
	model := models.FromDomain(stock)
	if err := r.getDB(ctx).Create(model).Error; err != nil {
		return 0, domainerr.Internal("failed to create stock")
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
		return domainerr.NotFound("no stocks found to update")
	}
	return nil
}

func (r *Repository) UpdateRealStockUnits(ctx context.Context, stockID int64, stock *domain.Stock) error {
	if stock == nil {
		return domainerr.Validation("stock is nil")
	}

	updateTx := r.getDB(ctx).
		Model(&models.Stock{}).
		Where("id = ?", stockID)

	if stock.Project != nil && stock.Project.ID > 0 {
		updateTx = updateTx.Where("project_id = ?", stock.Project.ID)
	}

	if !stock.UpdatedAt.IsZero() {
		updateTx = updateTx.Where("updated_at = ?", stock.UpdatedAt)
	}

	newUpdatedAt := time.Now().UTC()
	values := map[string]any{
		"real_stock_units":     stock.RealStockUnits,
		"has_real_stock_count": stock.HasRealStockCount,
		"updated_at":           newUpdatedAt,
		"updated_by":           stock.UpdatedBy,
	}

	result := updateTx.Updates(values)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		if !stock.UpdatedAt.IsZero() {
			return domainerr.Conflict("stock not found or outdated")
		}
		return domainerr.NotFound("no stock found to update")
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
			return domainerr.Conflict("stock not found or outdated")
		}
		return domainerr.NotFound("no stock found to update")
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
			return nil, domainerr.NotFound("stock not found")
		}
		return nil, domainerr.Internal("failed to get stock")
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

		return nil, false, domainerr.Internal("failed to get last stock")
	}

	movementsByKey, err := r.loadMovementsByStockKey(ctx, projectID)
	if err != nil {
		return nil, false, domainerr.Internal("failed to load stock movements")
	}
	_, consumedBySupplyID, err := r.loadConsumedByStockKey(ctx, projectID)
	if err != nil {
		return nil, false, domainerr.Internal("failed to load stock consumed")
	}
	for key, movements := range movementsByKey {
		if key.SupplyID == supplyID {
			stockModel.SupplyMovements = append(stockModel.SupplyMovements, movements...)
		}
	}
	stockModel.Consumed = consumedBySupplyID[supplyID]

	return stockModel.ToDomain(), false, nil
}

func (r *Repository) GetLastStockByProjectInvestorID(ctx context.Context, projectID int64, supplyID int64, investorID int64) (*domain.Stock, bool, error) {
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
		Where("investor_id = ?", investorID).
		Where("close_date is null").
		Order("id DESC").
		First(&stockModel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, true, nil
		}

		return nil, false, domainerr.Internal("failed to get last stock")
	}

	movementsByKey, err := r.loadMovementsByStockKey(ctx, projectID)
	if err != nil {
		return nil, false, domainerr.Internal("failed to load stock movements")
	}
	consumedByKey, _, err := r.loadConsumedByStockKey(ctx, projectID)
	if err != nil {
		return nil, false, domainerr.Internal("failed to load stock consumed")
	}
	key := keyFromStockModel(stockModel)
	stockModel.SupplyMovements = movementsByKey[key]
	stockModel.Consumed = consumedByKey[key]

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
			return nil, domainerr.NotFound("stock not found")
		}
		return nil, domainerr.Internal("failed to get stock by period")
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

	if len(stockModel) > 0 {
		movementsByKey, err := r.loadAllMovementsByStockKey(ctx)
		if err != nil {
			return nil, err
		}
		consumedByKey, err := r.loadAllConsumedByStockKey(ctx)
		if err != nil {
			return nil, err
		}

		for i := range stockModel {
			key := keyFromStockModel(stockModel[i])
			stockModel[i].SupplyMovements = movementsByKey[key]
			stockModel[i].Consumed = consumedByKey[key]
		}
	}

	out := make([]*domain.Stock, 0, len(stockModel))
	for i := range stockModel {
		out = append(out, stockModel[i].ToDomain())
	}
	return out, nil
}

func (r *Repository) loadAllMovementsByStockKey(ctx context.Context) (map[stockKey][]supplymodels.SupplyMovement, error) {
	var movements []supplymodels.SupplyMovement
	if err := r.getDB(ctx).
		Where("deleted_at IS NULL").
		Order("movement_date ASC, id ASC").
		Find(&movements).Error; err != nil {
		return nil, err
	}

	byKey := make(map[stockKey][]supplymodels.SupplyMovement)
	for _, movement := range movements {
		key := stockKey{
			ProjectID:  movement.ProjectId,
			SupplyID:   movement.SupplyID,
			InvestorID: movement.InvestorID,
		}
		byKey[key] = append(byKey[key], movement)
	}
	return byKey, nil
}

func (r *Repository) loadAllConsumedByStockKey(ctx context.Context) (map[stockKey]decimal.Decimal, error) {
	type consumedRow struct {
		ProjectID  int64           `gorm:"column:project_id"`
		SupplyID   int64           `gorm:"column:supply_id"`
		InvestorID int64           `gorm:"column:investor_id"`
		Consumed   decimal.Decimal `gorm:"column:consumed"`
	}

	var rows []consumedRow
	err := r.getDB(ctx).Raw(`
		SELECT
			wo.project_id,
			woi.supply_id,
			wo.investor_id,
			COALESCE(SUM(woi.total_used), 0) AS consumed
		FROM workorder_items woi
		JOIN workorders wo ON wo.id = woi.workorder_id
		WHERE wo.deleted_at IS NULL
		  AND woi.deleted_at IS NULL
		  AND woi.supply_id IS NOT NULL
		GROUP BY wo.project_id, woi.supply_id, wo.investor_id
	`).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	byKey := make(map[stockKey]decimal.Decimal, len(rows))
	for _, row := range rows {
		key := stockKey{
			ProjectID:  row.ProjectID,
			SupplyID:   row.SupplyID,
			InvestorID: row.InvestorID,
		}
		byKey[key] = row.Consumed
	}
	return byKey, nil
}
