package stock

import (
	"context"
	"time"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	models "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock/repository/models"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock/usecases/domain"
	workordermodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/repository/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type GormEnginePort interface {
	Client() *gorm.DB
}

type Repository struct {
	db GormEnginePort
}

// NewRepository crea una nueva instancia del repositorio de Stock
func NewRepository(db GormEnginePort) *Repository {
	return &Repository{db: db}
}

// GetStocks retorna stocks filtrando por nombre de proyecto, nombre de field y opcionalmente por fecha de corte
func (r *Repository) GetStocks(ctx context.Context, projectId int64, closeDate time.Time) ([]*domain.Stock, error) {
	db := r.db.Client().WithContext(ctx)
	var t time.Time

	query := db.Model(&models.Stock{}).
		Preload("Project").
		Preload("Supply").
		Preload("Supply.Type").
		Preload("Investor").
		Preload("SupplyMovements").
		Joins("JOIN projects ON projects.id = stocks.project_id").
		Where("projects.id = ?", projectId)

	if closeDate != t {
		query.Where("stocks.close_date = ?", closeDate)
	} else {
		query.Where("stocks.close_date IS NULL")
	}

	var stockModels []models.Stock
	if err := query.Order("stocks.id DESC").Find(&stockModels).Error; err != nil {
		return nil, err
	}

	stocks := make([]*domain.Stock, 0, len(stockModels))
	for i := range stockModels {
		var consumed decimal.Decimal
		err := db.Model(&workordermodels.WorkorderItem{}).
			Joins("JOIN workorders ON workorders.id = workorder_items.workorder_id").
			Where("workorders.project_id = ? AND workorder_items.supply_id = ?", projectId, stockModels[i].SupplyID).
			Select("COALESCE(SUM(workorder_items.total_used), 0)").
			Scan(&consumed).Error
		if err != nil {
			return nil, err
		}
		stockModels[i].Consumed = consumed

		stocks = append(stocks, stockModels[i].ToDomain())
	}
	return stocks, nil
}

func (r *Repository) GetStocksPeriods(ctx context.Context, projectId int64) ([]string, error) {
	var rawPeriods []time.Time

	err := r.db.Client().WithContext(ctx).
		Model(&models.Stock{}).
		Where("project_id = ? AND close_date IS NOT NULL", projectId).
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
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, types.NewError(types.ErrInternal, "failed to create stock", err)
	}
	return model.ID, nil
}

func (r *Repository) UpdateCloseDateByProject(ctx context.Context, projectId int64, stock *domain.Stock) error {
	stockUpdate := models.StockUpdateCloseDateFromDomain(stock)
	result := r.db.Client().WithContext(ctx).
		Model(&models.Stock{}).
		Where("project_id = ?", projectId).
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

func (r *Repository) UpdateRealStockUnits(ctx context.Context, stockId int64, stock *domain.Stock) error {
	stockUpdate := models.StockUpdateRealUnitsFromDomain(stock)
	result := r.db.Client().WithContext(ctx).
		Model(&models.Stock{}).
		Where("id = ?", stockId).
		Updates(stockUpdate)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return types.NewError(types.ErrNotFound, "no stock found to update", nil)
	}
	return nil
}

func (r *Repository) UpdateUnitsConsumed(ctx context.Context, stockDomain domain.Stock, quantity decimal.Decimal) error {
	result := r.db.Client().WithContext(ctx).
		Model(&models.Stock{}).
		Where("id = ?", stockDomain.ID).
		Update("units_consumed", gorm.Expr("units_consumed + ?", quantity))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return types.NewError(types.ErrNotFound, "no stock found to update", nil)
	}
	return nil
}

func (r *Repository) GetStockById(ctx context.Context, stockId int64) (*domain.Stock, error) {
	var stockModel models.Stock
	err := r.db.Client().WithContext(ctx).
		Preload("Project").
		Preload("Supply").
		Preload("Supply.Type").
		Preload("Investor").
		First(&stockModel, stockId).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, types.NewError(types.ErrNotFound, "stock not found", nil)
		}
		return nil, err
	}
	return stockModel.ToDomain(), nil
}

func (r *Repository) GetLastStockByProjectId(ctx context.Context, projectId int64, supplyId int64) (*domain.Stock, bool, error) {
	var stockModel models.Stock
	err := r.db.Client().WithContext(ctx).
		Preload("Project").
		Preload("Supply").
		Preload("Supply.Type").
		Preload("Investor").
		Where("project_id = ?", projectId).
		Where("supply_id = ?", supplyId).
		Where("close_date is null").
		First(&stockModel).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, true, nil
		}

		return nil, false, err
	}

	return stockModel.ToDomain(), false, nil

}

func (r *Repository) GetStockByPeriodAndProjectId(ctx context.Context, projectId int64) (*domain.Stock, error) {
	var stockModel models.Stock

	err := r.db.Client().WithContext(ctx).
		Preload("Project").
		Preload("Supply").
		Preload("Supply.Type").
		Preload("Investor").
		Where("project_id = ?", projectId).
		Where("close_date IS NULL").
		First(&stockModel).Error

	if err != nil {
		return nil, err
	}

	return stockModel.ToDomain(), nil

}
