package stock

import (
	"context"
	"time"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	models "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock/repository/models"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock/usecases/domain"
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
func (r *Repository) GetStocks(ctx context.Context, projectId int64, fieldId int64, monthPeriod int64, yearPeriod int64, closeDate time.Time) ([]*domain.Stock, error) {
	db := r.db.Client().WithContext(ctx)
	var t time.Time

	query := db.Model(&models.Stock{}).
		Preload("Project").
		Preload("Field").
		Preload("Supply").
		Preload("Supply.Type").
		Preload("Investor").
		Preload("SupplyMovements").
		Joins("JOIN projects ON projects.id = stocks.project_id").
		Joins("JOIN fields ON fields.id = stocks.field_id").
		Where("projects.id = ?", projectId).
		Where("stocks.month_period = ?", monthPeriod).
		Where("stocks.year_period = ?", yearPeriod)

	if fieldId != 0 {
		query.Where("fields.id = ?", fieldId)
	}

	if closeDate != t {
		query.Where("stocks.close_date < ?", closeDate)
	}

	var stockModels []models.Stock
	if err := query.Find(&stockModels).Error; err != nil {
		return nil, err
	}

	stocks := make([]*domain.Stock, 0, len(stockModels))
	for i := range stockModels {
		stocks = append(stocks, stockModels[i].ToDomain())
	}
	return stocks, nil
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

func (r *Repository) UpdateCloseDateByProjectAndField(ctx context.Context, projectId int64, fieldId int64, monthPeriod int64, yearPeriod int64, stock *domain.Stock) error {
	stockUpdate := models.StockUpdateCloseDateFromDomain(stock)
	result := r.db.Client().WithContext(ctx).
		Model(&models.Stock{}).
		Where("project_id = ?", projectId).
		Where("field_id = ?", fieldId).
		Where("year_period = ?", yearPeriod).
		Where("month_period = ?", monthPeriod).
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

func (r *Repository) GetStockById(ctx context.Context, stockId int64) (*domain.Stock, error) {
	var stockModel models.Stock
	err := r.db.Client().WithContext(ctx).
		Preload("Project").
		Preload("Field").
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

func (r *Repository) GetLastStockByProjectIdAndFieldId(ctx context.Context, projectId int64, fieldId int64, supplyId int64) (*domain.Stock, bool, error) {
	var stockModel models.Stock
	err := r.db.Client().WithContext(ctx).
		Preload("Project").
		Preload("Field").
		Preload("Supply").
		Preload("Supply.Type").
		Preload("Investor").
		Where("project_id = ?", projectId).
		Where("field_id = ?", fieldId).
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

func (r *Repository) GetStockByPeriodAndProjectIdAndFieldId(ctx context.Context, projectId int64, fieldId int64, monthPeriod int64, yearPeriod int64) (*domain.Stock, error) {
	var stockModel models.Stock

	err := r.db.Client().WithContext(ctx).
		Preload("Project").
		Preload("Field").
		Preload("Supply").
		Preload("Supply.Type").
		Preload("Investor").
		Where("project_id = ?", projectId).
		Where("field_id = ?", fieldId).
		Where("month_period = ?", monthPeriod).
		Where("year_period = ?", yearPeriod).
		First(&stockModel).Error

	if err != nil {
		return nil, err
	}

	return stockModel.ToDomain(), nil

}
