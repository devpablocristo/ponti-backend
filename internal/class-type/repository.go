package classtype

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	models "github.com/devpablocristo/ponti-backend/internal/class-type/repository/models"
	domain "github.com/devpablocristo/ponti-backend/internal/class-type/usecases/domain"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
	sharedrepo "github.com/devpablocristo/ponti-backend/internal/shared/repository"
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

func (r *Repository) CreateClassType(ctx context.Context, c *domain.ClassType) (int64, error) {
	model := models.FromDomain(c)
	model.Base = sharedmodels.Base{
		CreatedBy: c.CreatedBy,
		UpdatedBy: c.UpdatedBy,
	}
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, domainerr.Internal("failed to create class type")
	}
	// T1.e/T3: dual-write de tenant_id (flag-gated).
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		if err := r.db.Client().WithContext(ctx).Exec("UPDATE types SET tenant_id = ? WHERE id = ? AND tenant_id IS NULL", orgID, model.ID).Error; err != nil {
			return 0, domainerr.Internal("failed to set class type tenant")
		}
	}
	return model.ID, nil
}

func (r *Repository) ListClassTypes(ctx context.Context, page, perPage int) ([]domain.ClassType, int64, error) {
	var total int64
	countQ := r.db.Client().WithContext(ctx).Model(&models.ClassType{})
	// T1.e: acotar al tenant activo (flag-gated).
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		countQ = countQ.Where("tenant_id = ?", orgID)
	}
	if err := countQ.Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count class types")
	}

	var list []models.ClassType
	offset := (page - 1) * perPage
	listQ := r.db.Client().WithContext(ctx).
		Offset(offset).
		Limit(perPage).
		Order("id ASC")
	// T1.e: acotar al tenant activo (flag-gated).
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		listQ = listQ.Where("tenant_id = ?", orgID)
	}
	err := listQ.Find(&list).Error
	if err != nil {
		return nil, 0, domainerr.Internal("failed to list class types")
	}

	result := make([]domain.ClassType, 0, len(list))
	for _, m := range list {
		result = append(result, *m.ToDomain())
	}
	return result, total, nil
}

func (r *Repository) GetClassType(ctx context.Context, id int64) (*domain.ClassType, error) {
	if err := sharedrepo.ValidateID(id, "class type"); err != nil {
		return nil, err
	}
	var model models.ClassType
	q := r.db.Client().WithContext(ctx).Where("id = ?", id)
	// T1.e: guard de ownership (flag-gated) — NotFound si no es del tenant.
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		q = q.Where("tenant_id = ?", orgID)
	}
	if err := q.First(&model).Error; err != nil {
		return nil, sharedrepo.HandleGormError(err, "class type", id)
	}
	return model.ToDomain(), nil
}

func (r *Repository) UpdateClassType(ctx context.Context, c *domain.ClassType) error {
	if err := sharedrepo.ValidateID(c.ID, "class type"); err != nil {
		return err
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		existsQ := tx.Model(&models.ClassType{}).Where("id = ?", c.ID)
		// T1.e: guard de ownership (flag-gated).
		if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
			existsQ = existsQ.Where("tenant_id = ?", orgID)
		}
		if err := existsQ.Count(&count).Error; err != nil {
			return domainerr.Internal("failed to check class type existence")
		}
		if count == 0 {
			return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("class type %d not found", c.ID))
		}
		updateTx := tx.Model(&models.ClassType{}).
			Where("id = ?", c.ID)
		// T1.e: guard de ownership (flag-gated) — solo actualiza si es del tenant.
		if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
			updateTx = updateTx.Where("tenant_id = ?", orgID)
		}
		if !c.UpdatedAt.IsZero() {
			updateTx = updateTx.Where("updated_at = ?", c.UpdatedAt)
		}
		result := updateTx.Updates(map[string]any{
			"name":       c.Name,
			"updated_by": c.UpdatedBy,
		})
		if result.Error != nil {
			return domainerr.Internal("failed to update class type")
		}
		if result.RowsAffected == 0 {
			if !c.UpdatedAt.IsZero() {
				return domainerr.Conflict("class type not found or outdated")
			}
			return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("class type %d not found", c.ID))
		}
		return nil
	})
}

func (r *Repository) DeleteClassType(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "class type"); err != nil {
		return err
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		existsQ := tx.Model(&models.ClassType{}).Where("id = ?", id)
		// T1.e: guard de ownership (flag-gated).
		if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
			existsQ = existsQ.Where("tenant_id = ?", orgID)
		}
		if err := existsQ.Count(&count).Error; err != nil {
			return domainerr.Internal("failed to check class type existence")
		}
		if count == 0 {
			return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("class type %d not found", id))
		}
		deleteTx := tx.Where("id = ?", id)
		// T1.e: guard de ownership (flag-gated) — solo borra si es del tenant.
		if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
			deleteTx = deleteTx.Where("tenant_id = ?", orgID)
		}
		result := deleteTx.Delete(&models.ClassType{})
		if result.Error != nil {
			return domainerr.Internal("failed to delete class type")
		}
		if result.RowsAffected == 0 {
			return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("class type %d not found", id))
		}
		return nil
	})
}
