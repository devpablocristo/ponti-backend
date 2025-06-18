package lot

import (
	"context"
	"errors"
	"fmt"

	gorm "gorm.io/gorm"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	models "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
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

// --- CREATE ---
func (r *Repository) CreateLot(ctx context.Context, l *domain.Lot) (int64, error) {
	var lotID int64

	err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing models.Lot
		if err := tx.Where("name = ? AND field_id = ?", l.Name, l.FieldID).
			First(&existing).Error; err == nil {
			lotID = existing.ID
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return types.NewError(types.ErrInternal, "failed to check lot", err)
		}
		model := models.FromDomain(l)
		if err := tx.Create(model).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to create lot", err)
		}
		lotID = model.ID
		return nil
	})
	if err != nil {
		return 0, err
	}
	return lotID, nil
}

// --- LIST ---
func (r *Repository) ListLots(ctx context.Context, fieldID int64) ([]domain.Lot, error) {
	var lots []models.Lot
	if err := r.db.Client().WithContext(ctx).
		Where("field_id = ?", fieldID).
		Find(&lots).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list lots", err)
	}
	res := make([]domain.Lot, len(lots))
	for i := range lots {
		res[i] = *lots[i].ToDomain()
	}
	return res, nil
}

// --- GET ---
func (r *Repository) GetLot(ctx context.Context, id int64) (*domain.Lot, error) {
	var m models.Lot
	err := r.db.Client().WithContext(ctx).
		First(&m, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.NewError(types.ErrNotFound, fmt.Sprintf("lot %d not found", id), err)
		}
		return nil, types.NewError(types.ErrInternal, "failed to get lot", err)
	}
	return m.ToDomain(), nil
}

// --- UPDATE ---
func (r *Repository) UpdateLot(ctx context.Context, l *domain.Lot) error {
	if l.ID <= 0 {
		return types.NewInvalidIDError(fmt.Sprintf("invalid lot id: %d", l.ID), nil)
	}
	model := models.FromDomain(l)
	model.ID = l.ID
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.Lot{}).Where("id = ?", l.ID).Count(&count).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to check lot existence", err)
		}
		if count == 0 {
			return types.NewError(types.ErrNotFound, fmt.Sprintf("lot %d not found", l.ID), nil)
		}
		if err := tx.Model(&models.Lot{}).
			Where("id = ?", l.ID).
			Updates(map[string]any{
				"name":             l.Name,
				"field_id":         l.FieldID,
				"hectares":         l.Hectares,
				"previous_crop_id": l.PreviousCrop.ID,
				"current_crop_id":  l.CurrentCrop.ID,
				"season":           l.Season,
			}).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to update lot", err)
		}
		return nil
	})
}

// --- DELETE ---
func (r *Repository) DeleteLot(ctx context.Context, id int64) error {
	if id <= 0 {
		return types.NewInvalidIDError(fmt.Sprintf("invalid lot id: %d", id), nil)
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.Lot{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to check lot existence", err)
		}
		if count == 0 {
			return types.NewError(types.ErrNotFound, fmt.Sprintf("lot %d not found", id), nil)
		}
		if err := tx.Delete(&models.Lot{}, id).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to delete lot", err)
		}
		return nil
	})
}

func (r *Repository) ListLotsByProject(ctx context.Context, projectID int64) ([]domain.Lot, error) {
	var lots []models.Lot
	err := r.db.Client().WithContext(ctx).
		Joins("JOIN fields ON lots.field_id = fields.id").
		Where("fields.project_id = ?", projectID).
		Find(&lots).Error
	if err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list lots by project", err)
	}
	res := make([]domain.Lot, len(lots))
	for i := range lots {
		res[i] = *lots[i].ToDomain()
	}
	return res, nil
}

func (r *Repository) ListLotsByProjectAndField(ctx context.Context, projectID, fieldID int64) ([]domain.Lot, error) {
	var lots []models.Lot
	err := r.db.Client().WithContext(ctx).
		Joins("JOIN fields ON lots.field_id = fields.id").
		Where("fields.project_id = ? AND fields.id = ?", projectID, fieldID).
		Find(&lots).Error
	if err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list lots by project and field", err)
	}
	res := make([]domain.Lot, len(lots))
	for i := range lots {
		res[i] = *lots[i].ToDomain()
	}
	return res, nil
}

func (r *Repository) ListLotsByProjectFieldAndCrop(ctx context.Context, projectID, fieldID, cropID int64, cropType string) ([]domain.Lot, error) {
	var lots []models.Lot
	db := r.db.Client().WithContext(ctx).
		Joins("JOIN fields ON lots.field_id = fields.id").
		Where("fields.project_id = ? AND fields.id = ?", projectID, fieldID)
	switch cropType {
	case "current":
		db = db.Where("lots.current_crop_id = ?", cropID)
	case "previous":
		db = db.Where("lots.previous_crop_id = ?", cropID)
	case "both":
		db = db.Where("lots.current_crop_id = ? OR lots.previous_crop_id = ?", cropID, cropID)
	}
	err := db.Find(&lots).Error
	if err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list lots by project, field and crop", err)
	}
	res := make([]domain.Lot, len(lots))
	for i := range lots {
		res[i] = *lots[i].ToDomain()
	}
	return res, nil
}
