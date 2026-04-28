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

	if !closeDate.IsZero() {
		start, end := stockDateRange(closeDate)
		stockQuery = stockQuery.Where("stocks.close_date >= ? AND stocks.close_date < ?", start, end)
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

	movementsBySupplyID, err := r.loadMovementsBySupplyID(ctx, projectID, closeDate)
	if err != nil {
		return nil, err
	}
	_, consumedBySupplyID, err := r.loadConsumedByStockKey(ctx, projectID, closeDate)
	if err != nil {
		return nil, err
	}

	selectedStockModels := selectStockModelsForSummary(stockModels, closeDate)
	stocksBySupplyID := make(map[int64]*domain.Stock, len(selectedStockModels))
	for supplyID, modelsForSupply := range selectedStockModels {
		stocksBySupplyID[supplyID] = stockSummaryFromModels(
			modelsForSupply,
			movementsBySupplyID[supplyID],
			consumedBySupplyID[supplyID],
			closeDate,
		)
	}

	stocks := make([]*domain.Stock, 0, len(supplies)+len(stockModels))
	for _, supply := range supplies {
		if stockModelForSupply := stocksBySupplyID[supply.ID]; stockModelForSupply != nil {
			stocks = append(stocks, stockModelForSupply)
			continue
		}

		virtualStock := &domain.Stock{
			Supply:            supply.ToDomain(),
			SupplyMovements:   mapSupplyMovementsToDomain(movementsBySupplyID[supply.ID]),
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

func mapSupplyMovementsToDomain(movements []supplymodels.SupplyMovement) []supplymovementdomain.SupplyMovement {
	domains := make([]supplymovementdomain.SupplyMovement, 0, len(movements))
	for i := range movements {
		domains = append(domains, supplymovementdomain.SupplyMovement{
			ID:                   movements[i].ID,
			StockId:              movements[i].StockId,
			Quantity:             movements[i].Quantity,
			MovementType:         movements[i].MovementType,
			MovementDate:         movements[i].MovementDate,
			ReferenceNumber:      movements[i].ReferenceNumber,
			ProjectId:            movements[i].ProjectId,
			ProjectDestinationId: movements[i].ProjectDestinationId,
			IsEntry:              movements[i].IsEntry,
		})
	}
	return domains
}

func selectStockModelsForSummary(stockModels []models.Stock, closeDate time.Time) map[int64][]models.Stock {
	selected := make(map[int64][]models.Stock)
	bestDateBySupply := make(map[int64]time.Time)

	for _, stock := range stockModels {
		if !closeDate.IsZero() {
			selected[stock.SupplyID] = append(selected[stock.SupplyID], stock)
			continue
		}

		periodDate := stockSummaryPeriodDate(stock)
		bestDate, ok := bestDateBySupply[stock.SupplyID]
		if !ok || periodDate.After(bestDate) {
			bestDateBySupply[stock.SupplyID] = periodDate
			selected[stock.SupplyID] = []models.Stock{stock}
			continue
		}
		if periodDate.Equal(bestDate) {
			selected[stock.SupplyID] = append(selected[stock.SupplyID], stock)
		}
	}

	return selected
}

func stockSummaryPeriodDate(stock models.Stock) time.Time {
	if stock.CloseDate == nil {
		return time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC)
	}
	return *stock.CloseDate
}

func stockSummaryFromModels(
	stockModels []models.Stock,
	movements []supplymodels.SupplyMovement,
	consumed decimal.Decimal,
	closeDate time.Time,
) *domain.Stock {
	if len(stockModels) == 0 {
		return nil
	}

	stock := stockModels[0].ToDomain()
	stock.SupplyMovements = mapSupplyMovementsToDomain(movements)
	stock.Consumed = consumed
	if closeDate.IsZero() {
		stock.CloseDate = nil
	}

	if len(stockModels) == 1 {
		return stock
	}

	stock.ID = 0
	stock.RealStockUnits = decimal.Zero
	stock.HasRealStockCount = false
	stock.UpdatedAt = time.Time{}

	for _, stockModel := range stockModels {
		if stockModel.HasRealStockCount {
			stock.RealStockUnits = stock.RealStockUnits.Add(stockModel.RealStockUnits)
			stock.HasRealStockCount = true
		}
		if stockModel.UpdatedAt.After(stock.UpdatedAt) {
			stock.UpdatedAt = stockModel.UpdatedAt
		}
		if stockModel.InvestorID != stockModels[0].InvestorID {
			stock.Investor = nil
		}
	}

	return stock
}

func keyFromStockModel(stock models.Stock) stockKey {
	return stockKey{
		ProjectID:  stock.ProjectID,
		SupplyID:   stock.SupplyID,
		InvestorID: stock.InvestorID,
	}
}

func (r *Repository) loadMovementsByStockKey(ctx context.Context, projectID int64, closeDate time.Time) (map[stockKey][]supplymodels.SupplyMovement, error) {
	var movements []supplymodels.SupplyMovement
	query := r.getDB(ctx).
		Where("project_id = ?", projectID).
		Where("deleted_at IS NULL").
		Order("movement_date ASC, id ASC")
	if !closeDate.IsZero() {
		_, end := stockDateRange(closeDate)
		query = query.Where("movement_date < ?", end)
	}
	if err := query.Find(&movements).Error; err != nil {
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

func (r *Repository) loadMovementsBySupplyID(ctx context.Context, projectID int64, closeDate time.Time) (map[int64][]supplymodels.SupplyMovement, error) {
	var movements []supplymodels.SupplyMovement
	query := r.getDB(ctx).
		Where("project_id = ?", projectID).
		Where("deleted_at IS NULL").
		Order("movement_date ASC, id ASC")
	if !closeDate.IsZero() {
		_, end := stockDateRange(closeDate)
		query = query.Where("movement_date < ?", end)
	}
	if err := query.Find(&movements).Error; err != nil {
		return nil, err
	}

	bySupplyID := make(map[int64][]supplymodels.SupplyMovement)
	for _, movement := range movements {
		bySupplyID[movement.SupplyID] = append(bySupplyID[movement.SupplyID], movement)
	}
	return bySupplyID, nil
}

func (r *Repository) loadConsumedByStockKey(ctx context.Context, projectID int64, closeDate time.Time) (map[stockKey]decimal.Decimal, map[int64]decimal.Decimal, error) {
	type consumedRow struct {
		SupplyID   int64           `gorm:"column:supply_id"`
		InvestorID int64           `gorm:"column:investor_id"`
		Consumed   decimal.Decimal `gorm:"column:consumed"`
	}

	var rows []consumedRow
	args := []any{projectID}
	dateFilter := ""
	if !closeDate.IsZero() {
		_, end := stockDateRange(closeDate)
		dateFilter = "AND wo.date < ?"
		args = append(args, end)
	}

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
			  `+dateFilter+`
			GROUP BY woi.supply_id, wo.investor_id
		`, args...).Scan(&rows).Error
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

func stockDateRange(date time.Time) (time.Time, time.Time) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	return start, start.AddDate(0, 0, 1)
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

	movementsByKey, err := r.loadMovementsByStockKey(ctx, projectID, time.Time{})
	if err != nil {
		return nil, false, domainerr.Internal("failed to load stock movements")
	}
	_, consumedBySupplyID, err := r.loadConsumedByStockKey(ctx, projectID, time.Time{})
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

	movementsByKey, err := r.loadMovementsByStockKey(ctx, projectID, time.Time{})
	if err != nil {
		return nil, false, domainerr.Internal("failed to load stock movements")
	}
	consumedByKey, _, err := r.loadConsumedByStockKey(ctx, projectID, time.Time{})
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
