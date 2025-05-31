package campaign

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/campaign/repository/models"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/campaign/usecases/domain"
)

type GormEnginePort interface {
	Client() *gorm.DB
}
type Repository struct {
	db GormEnginePort
}

func NewRepository(db GormEnginePort) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) CreateCampaign(ctx context.Context, c *domain.Campaign) (int64, error) {
	if c == nil {
		return 0, types.NewError(types.ErrValidation, "customer is nil", nil)
	}
	model := models.FromDomain(c)
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, types.NewError(types.ErrInternal, "failed to create customer", err)
	}
	return model.ID, nil
}

func (r *Repository) ListCampaigns(ctx context.Context) ([]domain.Campaign, error) {
	var list []models.Campaign
	if err := r.db.Client().WithContext(ctx).Find(&list).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list customers", err)
	}
	result := make([]domain.Campaign, 0, len(list))
	for _, c := range list {
		result = append(result, *c.ToDomain())
	}
	return result, nil
}

func (r *Repository) GetCampaign(ctx context.Context, id int64) (*domain.Campaign, error) {
	var model models.Campaign
	err := r.db.Client().WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.NewError(types.ErrNotFound, fmt.Sprintf("customer with id %d not found", id), err)
		}
		return nil, types.NewError(types.ErrInternal, "failed to get customer", err)
	}
	return model.ToDomain(), nil
}
