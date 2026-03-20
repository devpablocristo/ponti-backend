package crop

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	models "github.com/devpablocristo/ponti-backend/internal/crop/repository/models"
	domain "github.com/devpablocristo/ponti-backend/internal/crop/usecases/domain"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
	sharedrepo "github.com/devpablocristo/ponti-backend/internal/shared/repository"
	types "github.com/devpablocristo/ponti-backend/pkg/types"
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

func (r *Repository) CreateCrop(ctx context.Context, c *domain.Crop) (int64, error) {
	if err := sharedrepo.ValidateEntity(c, "crop"); err != nil {
		return 0, err
	}
	model := models.FromDomainCrop(c)
	model.Base = sharedmodels.Base{
		CreatedBy: c.CreatedBy,
		UpdatedBy: c.UpdatedBy,
	}
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, types.NewError(types.ErrInternal, "failed to create crop", err)
	}
	return model.ID, nil
}

func (r *Repository) ListCrops(ctx context.Context, page, perPage int) ([]domain.Crop, int64, error) {
	var total int64
	if err := r.db.Client().WithContext(ctx).Model(&models.Crop{}).Count(&total).Error; err != nil {
		return nil, 0, types.NewError(types.ErrInternal, "failed to count crops", err)
	}

	var list []models.Crop
	offset := (page - 1) * perPage
	err := r.db.Client().WithContext(ctx).
		Offset(offset).
		Limit(perPage).
		Order("id ASC").
		Find(&list).Error
	if err != nil {
		return nil, 0, types.NewError(types.ErrInternal, "failed to list crops", err)
	}

	result := make([]domain.Crop, 0, len(list))
	for _, c := range list {
		result = append(result, *c.ToDomain())
	}
	return result, total, nil
}

func (r *Repository) GetCrop(ctx context.Context, id int64) (*domain.Crop, error) {
	if err := sharedrepo.ValidateID(id, "crop"); err != nil {
		return nil, err
	}
	var model models.Crop
	if err := r.db.Client().WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		return nil, sharedrepo.HandleGormError(err, "crop", id)
	}
	return model.ToDomain(), nil
}

func (r *Repository) UpdateCrop(ctx context.Context, c *domain.Crop) error {
	if err := sharedrepo.ValidateEntity(c, "crop"); err != nil {
		return err
	}
	if err := sharedrepo.ValidateID(c.ID, "crop"); err != nil {
		return err
	}
	updateTx := r.db.Client().WithContext(ctx).
		Model(&models.Crop{}).
		Where("id = ?", c.ID)
	if !c.UpdatedAt.IsZero() {
		updateTx = updateTx.Where("updated_at = ?", c.UpdatedAt)
	}
	result := updateTx.Updates(models.FromDomainCrop(c))
	if result.Error != nil {
		return types.NewError(types.ErrInternal, "failed to update crop", result.Error)
	}
	if result.RowsAffected == 0 {
		if !c.UpdatedAt.IsZero() {
			return types.NewError(types.ErrConflict, "crop not found or outdated", nil)
		}
		return types.NewError(types.ErrNotFound, fmt.Sprintf("crop with id %d does not exist", c.ID), nil)
	}
	return nil
}

func (r *Repository) DeleteCrop(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "crop"); err != nil {
		return err
	}
	result := r.db.Client().WithContext(ctx).
		Delete(&models.Crop{}, "id = ?", id)
	if result.Error != nil {
		return types.NewError(types.ErrInternal, "failed to delete crop", result.Error)
	}
	if result.RowsAffected == 0 {
		return types.NewError(types.ErrNotFound, fmt.Sprintf("crop with id %d does not exist", id), nil)
	}
	return nil
}
