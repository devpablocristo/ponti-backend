package manager

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	actorsync "github.com/devpablocristo/ponti-backend/internal/actor"
	models "github.com/devpablocristo/ponti-backend/internal/manager/repository/models"
	domain "github.com/devpablocristo/ponti-backend/internal/manager/usecases/domain"
	"github.com/devpablocristo/platform/persistence/gorm/go/tenancy"

	"github.com/devpablocristo/ponti-backend/internal/shared/authz"
	"github.com/devpablocristo/ponti-backend/internal/shared/lifecycle"
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

func (r *Repository) CreateManager(ctx context.Context, m *domain.Manager) (int64, error) {
	if err := sharedrepo.ValidateEntity(m, "manager"); err != nil {
		return 0, err
	}
	var id int64
	err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := assertManagerReferencesActive(tx, m); err != nil {
			return err
		}
		model := models.FromDomain(m)
		model.Base = sharedmodels.Base{
			CreatedBy: m.CreatedBy,
			UpdatedBy: m.UpdatedBy,
		}
		if tenantID, ok, err := authz.OptionalTenantOrStrict(ctx); err != nil {
			return err
		} else if ok {
			model.TenantID = tenantID
		}
		if err := tx.Create(model).Error; err != nil {
			return domainerr.Internal("failed to create manager")
		}
		id = model.ID
		actorID, err := actorsync.SyncLegacyActor(tx, actorsync.LegacyActorSync{
			SourceTable: actorsync.LegacyManagers,
			SourceID:    model.ID,
			Name:        model.Name,
			ActorKind:   actorsync.KindPerson,
			Role:        actorsync.RoleResponsable,
			CreatedAt:   model.CreatedAt,
			UpdatedAt:   model.UpdatedAt,
			CreatedBy:   model.CreatedBy,
			UpdatedBy:   model.UpdatedBy,
		})
		if err != nil {
			return err
		}
		m.ActorID = &actorID
		return nil
	})
	return id, err
}

type managerRow struct {
	ID         int64
	Name       string
	ActorID    *int64
	ArchivedAt *time.Time
}

func (r *Repository) ListManagers(ctx context.Context, page, perPage int) ([]domain.Manager, int64, error) {
	var total int64
	base := r.db.Client().WithContext(ctx).
		Table("managers m").
		Where("m.deleted_at IS NULL")
	base = tenancy.Scope(ctx, base, "m")
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count managers")
	}

	var list []managerRow
	offset := (page - 1) * perPage
	err := base.
		Select("m.id, m.name, lm.actor_id").
		Joins("LEFT JOIN legacy_actor_map lm ON lm.source_table = 'managers' AND lm.source_id = m.id AND lm.tenant_id = m.tenant_id").
		Offset(offset).
		Limit(perPage).
		Order("m.id ASC").
		Scan(&list).Error
	if err != nil {
		return nil, 0, domainerr.Internal("failed to list managers")
	}

	result := make([]domain.Manager, 0, len(list))
	for _, m := range list {
		result = append(result, domain.Manager{
			ID:      m.ID,
			Name:    m.Name,
			ActorID: m.ActorID,
		})
	}
	return result, total, nil
}

func (r *Repository) ListArchivedManagers(ctx context.Context, page, perPage int) ([]domain.Manager, int64, error) {
	var total int64
	base := r.db.Client().WithContext(ctx).
		Unscoped().
		Table("managers m").
		Where("m.deleted_at IS NOT NULL")
	base = tenancy.Scope(ctx, base, "m")

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count archived managers")
	}

	var list []managerRow
	offset := (page - 1) * perPage
	if err := base.
		Select("m.id, m.name, lm.actor_id, m.deleted_at AS archived_at").
		Joins("LEFT JOIN legacy_actor_map lm ON lm.source_table = 'managers' AND lm.source_id = m.id AND lm.tenant_id = m.tenant_id").
		Offset(offset).
		Limit(perPage).
		Order("m.deleted_at DESC").
		Scan(&list).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to list archived managers")
	}

	result := make([]domain.Manager, 0, len(list))
	for _, m := range list {
		result = append(result, domain.Manager{
			ID:         m.ID,
			Name:       m.Name,
			ActorID:    m.ActorID,
			ArchivedAt: m.ArchivedAt,
		})
	}
	return result, total, nil
}

func (r *Repository) GetManager(ctx context.Context, id int64) (*domain.Manager, error) {
	if err := sharedrepo.ValidateID(id, "manager"); err != nil {
		return nil, err
	}
	var model models.Manager
	db0 := tenancy.Scope(ctx, r.db.Client().WithContext(ctx), "managers")
	if err := db0.
		Where("id = ? AND deleted_at IS NULL", id).
		First(&model).Error; err != nil {
		return nil, sharedrepo.HandleGormError(err, "manager", id)
	}
	out := model.ToDomain()
	actorID, err := actorsync.ActorIDForLegacy(r.db.Client().WithContext(ctx), actorsync.LegacyManagers, id)
	if err != nil {
		return nil, err
	}
	if actorID > 0 {
		out.ActorID = &actorID
	}
	return out, nil
}

func (r *Repository) UpdateManager(ctx context.Context, m *domain.Manager) error {
	if err := sharedrepo.ValidateEntity(m, "manager"); err != nil {
		return err
	}
	if err := sharedrepo.ValidateID(m.ID, "manager"); err != nil {
		return err
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := assertManagerReferencesActive(tx, m); err != nil {
			return err
		}
		updateTx := tenancy.Scope(ctx, tx.Model(&models.Manager{}), "managers").Where("id = ?", m.ID)
		if !m.UpdatedAt.IsZero() {
			updateTx = updateTx.Where("updated_at = ?", m.UpdatedAt)
		}
		result := updateTx.Updates(models.FromDomain(m))
		if result.Error != nil {
			return domainerr.Internal("failed to update manager")
		}
		if result.RowsAffected == 0 {
			if !m.UpdatedAt.IsZero() {
				return domainerr.Conflict("manager not found or outdated")
			}
			return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("manager with id %d does not exist", m.ID))
		}
		_, err := actorsync.SyncLegacyActor(tx, actorsync.LegacyActorSync{
			SourceTable: actorsync.LegacyManagers,
			SourceID:    m.ID,
			Name:        m.Name,
			ActorKind:   actorsync.KindPerson,
			Role:        actorsync.RoleResponsable,
			UpdatedAt:   time.Now(),
			UpdatedBy:   m.UpdatedBy,
		})
		return err
	})
}

// HardDeleteManager elimina definitivamente un manager.
// Bloquea con 409 si tiene asignaciones (activas o archivadas) en project_managers.
func (r *Repository) HardDeleteManager(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "manager"); err != nil {
		return err
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		managerDB := tenancy.Scope(ctx, tx.Unscoped().Table("managers"), "managers")
		if err := managerDB.Where("id = ?", id).Count(&count).Error; err != nil {
			return domainerr.Internal("failed to check manager existence")
		}
		if count == 0 {
			return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("manager with id %d does not exist", id))
		}
		if err := lifecycle.RequireArchived(managerDB, "managers", "manager", id); err != nil {
			return err
		}

		var pmCount int64
		pmDB := tenancy.Scope(ctx, tx.Unscoped().Table("project_managers"), "project_managers")
		if err := pmDB.Where("manager_id = ?", id).Count(&pmCount).Error; err != nil {
			return domainerr.Internal("failed to check project assignments")
		}
		if pmCount > 0 {
			return domainerr.Conflict(fmt.Sprintf("manager has %d project assignment(s); archive or remove them first", pmCount))
		}

		var deletedBy *string
		if actor, err := sharedmodels.ActorFromContext(ctx); err == nil {
			deletedBy = &actor
		}
		if err := actorsync.DeleteLegacyActor(tx, actorsync.LegacyManagers, id, actorsync.RoleResponsable, deletedBy); err != nil {
			return err
		}
		if err := tenancy.Scope(ctx, tx.Unscoped(), "managers").Delete(&models.Manager{}, "id = ?", id).Error; err != nil {
			return domainerr.Internal("failed to hard delete manager")
		}
		return nil
	})
}

// ArchiveManager archiva (soft delete) un manager con validación.
func (r *Repository) ArchiveManager(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "manager"); err != nil {
		return err
	}
	actor, err := sharedmodels.ActorFromContext(ctx)
	if err != nil {
		return err
	}
	deletedBy := &actor

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var m models.Manager
		managerQuery := tenancy.Scope(ctx, tx.Unscoped(), "managers")
		if err := managerQuery.Where("id = ?", id).First(&m).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("manager %d not found", id))
			}
			return domainerr.Internal("failed to get manager")
		}
		if m.DeletedAt.Valid {
			return domainerr.Conflict("manager already archived")
		}

		// Block when the manager still has active assignments. Archiving
		// would leave project_managers rows pointing at an archived manager,
		// violating "archived = no existe". User must remove the assignments
		// (or archive the parent project, which cascades).
		var activeAssignments int64
		assignmentsQuery := tenancy.Scope(ctx, tx.Table("project_managers"), "project_managers").
			Where("manager_id = ? AND deleted_at IS NULL", id)
		if err := assignmentsQuery.Count(&activeAssignments).Error; err != nil {
			return domainerr.Internal("failed to check project assignments")
		}
		if activeAssignments > 0 {
			return domainerr.Conflict(fmt.Sprintf("manager has %d active project assignment(s); remove them first", activeAssignments))
		}

		archivedAt := time.Now()
		cause, err := lifecycle.RootCause(tx, m.TenantID, "managers", id, nil, deletedBy)
		if err != nil {
			return err
		}
		if err := tenancy.Scope(ctx, tx.Model(&models.Manager{}), "managers").
			Where("id = ?", id).
			Updates(lifecycle.ArchiveUpdates(tx, "managers", archivedAt, deletedBy, cause)).Error; err != nil {
			return domainerr.Internal("failed to archive manager")
		}
		if _, err := actorsync.SyncLegacyActor(tx, actorsync.LegacyActorSync{
			SourceTable: actorsync.LegacyManagers,
			SourceID:    m.ID,
			Name:        m.Name,
			ActorKind:   actorsync.KindPerson,
			Role:        actorsync.RoleResponsable,
			ArchivedAt:  &archivedAt,
			UpdatedAt:   archivedAt,
			UpdatedBy:   m.UpdatedBy,
			DeletedBy:   deletedBy,
		}); err != nil {
			return err
		}
		return nil
	})
}

// assertManagerReferencesActive blocks Create/Update of a manager that
// references an archived actor. Manager.ActorID is optional (legacy rows
// can be nil) — only validate when present.
func assertManagerReferencesActive(tx *gorm.DB, m *domain.Manager) error {
	if m == nil {
		return nil
	}
	refs := []lifecycle.ActiveRef{}
	if m.ActorID != nil {
		refs = append(refs, lifecycle.ActiveRef{Table: "actors", Label: "actor", ID: *m.ActorID})
	}
	return lifecycle.RequireAllActive(tx, refs)
}

// RestoreManager restaura un manager archivado.
func (r *Repository) RestoreManager(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "manager"); err != nil {
		return err
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		restoredAt := time.Now()
		var m models.Manager
		managerQuery := tenancy.Scope(ctx, tx.Unscoped(), "managers")
		if err := managerQuery.Where("id = ?", id).First(&m).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("manager %d not found", id))
			}
			return domainerr.Internal("failed to get manager")
		}
		if !m.DeletedAt.Valid {
			return domainerr.Conflict("manager is not archived")
		}

		if err := tenancy.Scope(ctx, tx.Unscoped().Model(&models.Manager{}), "managers").
			Where("id = ?", id).
			Updates(lifecycle.RestoreUpdates(tx, "managers", restoredAt)).Error; err != nil {
			return domainerr.Internal("failed to restore manager")
		}
		if _, err := actorsync.SyncLegacyActor(tx, actorsync.LegacyActorSync{
			SourceTable: actorsync.LegacyManagers,
			SourceID:    m.ID,
			Name:        m.Name,
			ActorKind:   actorsync.KindPerson,
			Role:        actorsync.RoleResponsable,
			UpdatedAt:   restoredAt,
			UpdatedBy:   m.UpdatedBy,
		}); err != nil {
			return err
		}
		return nil
	})
}
