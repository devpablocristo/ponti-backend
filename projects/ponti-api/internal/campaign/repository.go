package campaign

import (
	"context"
	"errors"
	"fmt"

	gorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	pkgtypes "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/campaign/repository/models"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/campaign/usecases/domain"
	gorm0 "gorm.io/gorm"
)

type repository struct {
	db gorm.Repository
}

func NewRepository(db gorm.Repository) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) CreateCampaign(ctx context.Context, c *domain.Campaign) (int64, error) {
	if c == nil {
		return 0, pkgtypes.NewError(pkgtypes.ErrValidation, "customer is nil", nil)
	}
	model := models.FromDomain(c)
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to create customer", err)
	}
	return model.ID, nil
}

func (r *repository) ListCampaigns(ctx context.Context) ([]domain.Campaign, error) {
	var list []models.Campaign
	if err := r.db.Client().WithContext(ctx).Find(&list).Error; err != nil {
		return nil, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to list customers", err)
	}
	result := make([]domain.Campaign, 0, len(list))
	for _, c := range list {
		result = append(result, *c.ToDomain())
	}
	return result, nil
}

func (r *repository) GetCampaign(ctx context.Context, id int64) (*domain.Campaign, error) {
	var model models.Campaign
	err := r.db.Client().WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm0.ErrRecordNotFound) {
			return nil, pkgtypes.NewError(pkgtypes.ErrNotFound, fmt.Sprintf("customer with id %d not found", id), err)
		}
		return nil, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to get customer", err)
	}
	return model.ToDomain(), nil
}
