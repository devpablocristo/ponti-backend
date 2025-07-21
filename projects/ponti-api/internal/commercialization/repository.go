package commercialization

import (
	"context"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"gorm.io/gorm"

	models "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/commercialization/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/commercialization/usecases/domain"
)

type GormEnginePort interface {
	Client() *gorm.DB
}

type Repository struct {
	db GormEnginePort
}

func NewRepository(db GormEnginePort) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateBulk(ctx context.Context, items []domain.CropCommercialization) error {
	if len(items) == 0 {
		return nil
	}

	modelList := make([]models.CropCommercialization, len(items))
	for i, item := range items {
		modelList[i] = *models.FromDomain(&item)
	}

	if err := r.db.Client().WithContext(ctx).Create(&modelList).Error; err != nil {
		return types.NewError(types.ErrInternal, "failed to bulk insert crop commercializations", err)
	}

	return nil
}

func (r *Repository) ListByProject(ctx context.Context, projectID int64) ([]domain.CropCommercialization, error) {

	tx := r.db.Client().
		WithContext(ctx).
		Model(&models.CropCommercialization{}).
		Where("project_id = ?", projectID)

	var rows []models.CropCommercialization

	if err := tx.Find(&rows).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list crop commercialization", err)
	}

	out := make([]domain.CropCommercialization, len(rows))
	for i, m := range rows {
		out[i] = *m.ToDomain()
	}

	return out, nil
}
