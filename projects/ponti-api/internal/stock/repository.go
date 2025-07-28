package stock

import (
	"context"
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock/repository/models"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock/usecases/domain"
	"gorm.io/gorm"
	"time"
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
func (r *Repository) GetStocks(ctx context.Context, projectId int64, fieldId int64, closeDate time.Time) ([]*domain.Stock, error) {
	db := r.db.Client().WithContext(ctx)
	var t time.Time

	query := db.Model(&models.Stock{}).
		Preload("Project").
		Preload("Field").
		Preload("Supply").
		Preload("Investor").
		Joins("JOIN projects ON projects.id = stocks.project_id").
		Joins("JOIN fields ON fields.id = stocks.field_id").
		Where("projects.id = ?", projectId).
		Where("fields.id = ?", fieldId)

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
