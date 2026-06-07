package category

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	models "github.com/devpablocristo/ponti-backend/internal/category/repository/models"
	domain "github.com/devpablocristo/ponti-backend/internal/category/usecases/domain"
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

func (r *Repository) CreateCategory(ctx context.Context, c *domain.Category) (int64, error) {
	if err := sharedrepo.ValidateEntity(c, "category"); err != nil {
		return 0, err
	}
	model := models.FromDomain(c)
	model.Base = sharedmodels.Base{
		CreatedBy: c.CreatedBy,
		UpdatedBy: c.UpdatedBy,
	}
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		if sharedrepo.IsUniqueViolation(err) {
			return 0, domainerr.Conflict("a category with that name already exists for this type")
		}
		return 0, domainerr.Internal("failed to create category")
	}
	// T1.e: dual-write de tenant_id (flag-gated).
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		if err := r.db.Client().WithContext(ctx).Exec("UPDATE categories SET tenant_id = ? WHERE id = ? AND tenant_id IS NULL", orgID, model.ID).Error; err != nil {
			return 0, domainerr.Internal("failed to set category tenant")
		}
	}
	return model.ID, nil
}

func (r *Repository) ListCategories(ctx context.Context, page, perPage int) ([]domain.Category, int64, error) {
	var total int64

	countTx := r.db.Client().WithContext(ctx).Model(&models.Category{})
	// T1.e: acotar al tenant activo (flag-gated).
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		countTx = countTx.Where("tenant_id = ?", orgID)
	}
	if err := countTx.Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count categories")
	}

	var list []models.Category
	offset := (page - 1) * perPage
	listTx := r.db.Client().WithContext(ctx).
		Offset(offset).
		Limit(perPage).
		Order("id ASC")
	// T1.e: acotar al tenant activo (flag-gated).
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		listTx = listTx.Where("tenant_id = ?", orgID)
	}
	err := listTx.Find(&list).Error
	if err != nil {
		return nil, 0, domainerr.Internal("failed to list categories")
	}

	result := make([]domain.Category, 0, len(list))
	for _, c := range list {
		result = append(result, *c.ToDomain())
	}
	return result, total, nil
}

func (r *Repository) GetCategory(ctx context.Context, id int64) (*domain.Category, error) {
	if err := sharedrepo.ValidateID(id, "category"); err != nil {
		return nil, err
	}
	var model models.Category
	q := r.db.Client().WithContext(ctx).Where("id = ?", id)
	// T1.e: guard de ownership (flag-gated) — NotFound si la category no es del tenant.
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		q = q.Where("tenant_id = ?", orgID)
	}
	if err := q.First(&model).Error; err != nil {
		return nil, sharedrepo.HandleGormError(err, "category", id)
	}
	return model.ToDomain(), nil
}

func (r *Repository) UpdateCategory(ctx context.Context, c *domain.Category) error {
	if err := sharedrepo.ValidateEntity(c, "category"); err != nil {
		return err
	}
	if err := sharedrepo.ValidateID(c.ID, "category"); err != nil {
		return err
	}
	updateTx := r.db.Client().WithContext(ctx).
		Model(&models.Category{}).
		Where("id = ?", c.ID)
	if !c.UpdatedAt.IsZero() {
		updateTx = updateTx.Where("updated_at = ?", c.UpdatedAt)
	}
	// T1.e: guard de ownership (flag-gated) — solo actualiza si es del tenant.
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		updateTx = updateTx.Where("tenant_id = ?", orgID)
	}
	result := updateTx.Updates(models.FromDomain(c))
	if result.Error != nil {
		return domainerr.Internal("failed to update category")
	}
	if result.RowsAffected == 0 {
		if !c.UpdatedAt.IsZero() {
			return domainerr.Conflict("category not found or outdated")
		}
		return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("category with id %d does not exist", c.ID))
	}
	return nil
}

func (r *Repository) DeleteCategory(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "category"); err != nil {
		return err
	}
	deleteTx := r.db.Client().WithContext(ctx).Where("id = ?", id)
	// T1.e: guard de ownership (flag-gated) — solo borra si es del tenant.
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		deleteTx = deleteTx.Where("tenant_id = ?", orgID)
	}
	result := deleteTx.Delete(&models.Category{})
	if result.Error != nil {
		return domainerr.Internal("failed to delete category")
	}
	if result.RowsAffected == 0 {
		return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("category with id %d does not exist", id))
	}
	return nil
}
