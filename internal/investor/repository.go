package investor

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	actorsync "github.com/devpablocristo/ponti-backend/internal/actor"
	models "github.com/devpablocristo/ponti-backend/internal/investor/repository/models"
	domain "github.com/devpablocristo/ponti-backend/internal/investor/usecases/domain"
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

func (r *Repository) CreateInvestor(ctx context.Context, inv *domain.Investor) (int64, error) {
	if err := sharedrepo.ValidateEntity(inv, "investor"); err != nil {
		return 0, err
	}
	var id int64
	err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := assertInvestorReferencesActive(tx, inv); err != nil {
			return err
		}
		model := models.FromDomain(inv)
		model.Base = sharedmodels.Base{
			CreatedBy: inv.CreatedBy,
			UpdatedBy: inv.UpdatedBy,
		}
		if tenantID, ok, err := authz.OptionalTenantOrStrict(ctx); err != nil {
			return err
		} else if ok {
			model.TenantID = tenantID
		}
		if err := tx.Create(model).Error; err != nil {
			return domainerr.Internal("failed to create investor")
		}
		id = model.ID
		actorID, err := actorsync.SyncLegacyActor(tx, actorsync.LegacyActorSync{
			SourceTable: actorsync.LegacyInvestors,
			SourceID:    model.ID,
			Name:        model.Name,
			ActorKind:   actorsync.KindUnknown,
			Role:        actorsync.RoleInversor,
			CreatedAt:   model.CreatedAt,
			UpdatedAt:   model.UpdatedAt,
			CreatedBy:   model.CreatedBy,
			UpdatedBy:   model.UpdatedBy,
		})
		if err != nil {
			return err
		}
		inv.ActorID = &actorID
		return nil
	})
	return id, err
}

type investorRow struct {
	ID         int64
	Name       string
	Percentage int
	ActorID    *int64
	ArchivedAt *time.Time
}

func (r *Repository) ListInvestors(ctx context.Context, page, perPage int) ([]domain.Investor, int64, error) {
	var total int64
	base := r.db.Client().WithContext(ctx).
		Table("investors i").
		Where("i.deleted_at IS NULL")
	base = authz.MaybeTenantScope(ctx, base, "i")
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count investors")
	}

	var list []investorRow
	offset := (page - 1) * perPage
	err := base.
		Select("i.id, i.name, m.actor_id").
		Joins("LEFT JOIN legacy_actor_map m ON m.source_table = 'investors' AND m.source_id = i.id AND m.tenant_id = i.tenant_id").
		Offset(offset).
		Limit(perPage).
		Order("i.id ASC").
		Scan(&list).Error
	if err != nil {
		return nil, 0, domainerr.Internal("failed to list investors")
	}

	result := make([]domain.Investor, 0, len(list))
	for _, m := range list {
		result = append(result, domain.Investor{
			ID:      m.ID,
			Name:    m.Name,
			ActorID: m.ActorID,
		})
	}
	return result, total, nil
}

func (r *Repository) GetInvestor(ctx context.Context, id int64) (*domain.Investor, error) {
	var model models.Investor
	db0 := authz.MaybeTenantScope(ctx, r.db.Client().WithContext(ctx), "investors")
	if err := db0.
		Where("id = ? AND deleted_at IS NULL", id).
		First(&model).Error; err != nil {
		return nil, sharedrepo.HandleGormError(err, "investor", id)
	}
	out := model.ToDomain()
	actorID, err := actorsync.ActorIDForLegacy(r.db.Client().WithContext(ctx), actorsync.LegacyInvestors, id)
	if err != nil {
		return nil, err
	}
	if actorID > 0 {
		out.ActorID = &actorID
	}
	return out, nil
}

func (r *Repository) UpdateInvestor(ctx context.Context, inv *domain.Investor) error {
	if err := sharedrepo.ValidateEntity(inv, "investor"); err != nil {
		return err
	}
	if err := sharedrepo.ValidateID(inv.ID, "investor"); err != nil {
		return err
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := assertInvestorReferencesActive(tx, inv); err != nil {
			return err
		}
		updateTx := authz.MaybeTenantScope(ctx, tx.Model(&models.Investor{}), "investors").Where("id = ?", inv.ID)
		if !inv.UpdatedAt.IsZero() {
			updateTx = updateTx.Where("updated_at = ?", inv.UpdatedAt)
		}
		result := updateTx.Updates(models.FromDomain(inv))
		if result.Error != nil {
			return domainerr.Internal("failed to update investor")
		}
		if result.RowsAffected == 0 {
			if !inv.UpdatedAt.IsZero() {
				return domainerr.Conflict("investor not found or outdated")
			}
			return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("investor with id %d does not exist", inv.ID))
		}
		_, err := actorsync.SyncLegacyActor(tx, actorsync.LegacyActorSync{
			SourceTable: actorsync.LegacyInvestors,
			SourceID:    inv.ID,
			Name:        inv.Name,
			ActorKind:   actorsync.KindUnknown,
			Role:        actorsync.RoleInversor,
			UpdatedAt:   time.Now(),
			UpdatedBy:   inv.UpdatedBy,
		})
		return err
	})
}

// DeleteInvestor permanece como helper interno; los handlers usan HardDeleteInvestor.
func (r *Repository) DeleteInvestor(ctx context.Context, id int64) error {
	return r.HardDeleteInvestor(ctx, id)
}

func (r *Repository) ListArchivedInvestors(ctx context.Context, page, perPage int) ([]domain.Investor, int64, error) {
	var total int64
	base := r.db.Client().WithContext(ctx).
		Unscoped().
		Table("investors i").
		Where("i.deleted_at IS NOT NULL")
	base = authz.MaybeTenantScope(ctx, base, "i")

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count archived investors")
	}

	var list []investorRow
	offset := (page - 1) * perPage
	if err := base.
		Select("i.id, i.name, m.actor_id, i.deleted_at AS archived_at").
		Joins("LEFT JOIN legacy_actor_map m ON m.source_table = 'investors' AND m.source_id = i.id AND m.tenant_id = i.tenant_id").
		Offset(offset).
		Limit(perPage).
		Order("i.deleted_at DESC").
		Scan(&list).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to list archived investors")
	}

	result := make([]domain.Investor, 0, len(list))
	for _, m := range list {
		result = append(result, domain.Investor{
			ID:         m.ID,
			Name:       m.Name,
			ActorID:    m.ActorID,
			ArchivedAt: m.ArchivedAt,
		})
	}
	return result, total, nil
}

func (r *Repository) ArchiveInvestor(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "investor"); err != nil {
		return err
	}
	actor, err := sharedmodels.ActorFromContext(ctx)
	if err != nil {
		return err
	}
	deletedBy := &actor

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var inv models.Investor
		investorQuery := authz.MaybeTenantScope(ctx, tx.Unscoped(), "investors")
		if err := investorQuery.Where("id = ?", id).First(&inv).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("investor %d not found", id))
			}
			return domainerr.Internal("failed to get investor")
		}
		if inv.DeletedAt.Valid {
			return domainerr.Conflict("investor already archived")
		}

		// Block when the investor still has active assignments across the
		// pivot tables. Archive would leave any of them pointing at an
		// archived row, breaking the invariant. The user must remove (or
		// archive via cascade) the assignments first.
		type pivot struct {
			table  string
			column string
		}
		pivots := []pivot{
			{"project_investors", "investor_id"},
			{"field_investors", "investor_id"},
			{"workorder_investor_splits", "investor_id"},
			{"work_order_draft_investor_splits", "investor_id"},
			{"admin_cost_investors", "investor_id"},
		}
		var totalActive int64
		for _, p := range pivots {
			if !tx.Migrator().HasTable(p.table) {
				continue
			}
			var n int64
			q := authz.MaybeTenantScope(ctx, tx.Table(p.table), p.table).
				Where(p.column+" = ? AND deleted_at IS NULL", id)
			if err := q.Count(&n).Error; err != nil {
				return domainerr.Internal(fmt.Sprintf("failed to check %s assignments", p.table))
			}
			totalActive += n
		}
		if totalActive > 0 {
			return domainerr.Conflict(fmt.Sprintf("investor has %d active assignment(s); remove them first", totalActive))
		}

		archivedAt := time.Now()
		cause, err := lifecycle.RootCause(tx, inv.TenantID, "investors", id, nil, deletedBy)
		if err != nil {
			return err
		}
		if err := authz.MaybeTenantScope(ctx, tx.Model(&models.Investor{}), "investors").
			Where("id = ?", id).
			Updates(lifecycle.ArchiveUpdates(tx, "investors", archivedAt, deletedBy, cause)).Error; err != nil {
			return domainerr.Internal("failed to archive investor")
		}
		if _, err := actorsync.SyncLegacyActor(tx, actorsync.LegacyActorSync{
			SourceTable: actorsync.LegacyInvestors,
			SourceID:    inv.ID,
			Name:        inv.Name,
			ActorKind:   actorsync.KindUnknown,
			Role:        actorsync.RoleInversor,
			ArchivedAt:  &archivedAt,
			UpdatedAt:   archivedAt,
			UpdatedBy:   inv.UpdatedBy,
			DeletedBy:   deletedBy,
		}); err != nil {
			return err
		}
		return nil
	})
}

func (r *Repository) RestoreInvestor(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "investor"); err != nil {
		return err
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		restoredAt := time.Now()
		var inv models.Investor
		investorQuery := authz.MaybeTenantScope(ctx, tx.Unscoped(), "investors")
		if err := investorQuery.Where("id = ?", id).First(&inv).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("investor %d not found", id))
			}
			return domainerr.Internal("failed to get investor")
		}
		if !inv.DeletedAt.Valid {
			return domainerr.Conflict("investor is not archived")
		}

		if err := authz.MaybeTenantScope(ctx, tx.Unscoped().Model(&models.Investor{}), "investors").
			Where("id = ?", id).
			Updates(lifecycle.RestoreUpdates(tx, "investors", restoredAt)).Error; err != nil {
			return domainerr.Internal("failed to restore investor")
		}
		if _, err := actorsync.SyncLegacyActor(tx, actorsync.LegacyActorSync{
			SourceTable: actorsync.LegacyInvestors,
			SourceID:    inv.ID,
			Name:        inv.Name,
			ActorKind:   actorsync.KindUnknown,
			Role:        actorsync.RoleInversor,
			UpdatedAt:   restoredAt,
			UpdatedBy:   inv.UpdatedBy,
		}); err != nil {
			return err
		}
		return nil
	})
}

// HardDeleteInvestor elimina definitivamente un inversor.
// Bloquea con 409 si tiene registros (activos o archivados) en project_investors,
// field_investors o admin_cost_investors.
func (r *Repository) HardDeleteInvestor(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "investor"); err != nil {
		return err
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		investorDB := authz.MaybeTenantScope(ctx, tx.Unscoped().Table("investors"), "investors")
		if err := investorDB.Where("id = ?", id).Count(&count).Error; err != nil {
			return domainerr.Internal("failed to check investor existence")
		}
		if count == 0 {
			return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("investor with id %d does not exist", id))
		}
		if err := lifecycle.RequireArchived(investorDB, "investors", "investor", id); err != nil {
			return err
		}

		// Validar dependientes en las tres tablas pivot (incluso archivados).
		type dep struct {
			table string
			label string
		}
		deps := []dep{
			{"project_investors", "project assignments"},
			{"field_investors", "field assignments"},
			{"admin_cost_investors", "admin cost assignments"},
		}
		for _, d := range deps {
			var n int64
			depDB := authz.MaybeTenantScope(ctx, tx.Unscoped().Table(d.table), d.table)
			if err := depDB.Where("investor_id = ?", id).Count(&n).Error; err != nil {
				return domainerr.Internal(fmt.Sprintf("failed to check %s", d.table))
			}
			if n > 0 {
				return domainerr.Conflict(fmt.Sprintf("investor has %d %s; archive or remove them first", n, d.label))
			}
		}

		var deletedBy *string
		if actor, err := sharedmodels.ActorFromContext(ctx); err == nil {
			deletedBy = &actor
		}
		if err := actorsync.DeleteLegacyActor(tx, actorsync.LegacyInvestors, id, actorsync.RoleInversor, deletedBy); err != nil {
			return err
		}
		if err := authz.MaybeTenantScope(ctx, tx.Unscoped(), "investors").Delete(&models.Investor{}, "id = ?", id).Error; err != nil {
			return domainerr.Internal("failed to hard delete investor")
		}
		return nil
	})
}

// assertInvestorReferencesActive blocks Create/Update of an investor that
// references an archived actor. Investor.ActorID is optional (legacy rows
// can be nil) — only validate when present.
func assertInvestorReferencesActive(tx *gorm.DB, inv *domain.Investor) error {
	if inv == nil {
		return nil
	}
	refs := []lifecycle.ActiveRef{}
	if inv.ActorID != nil {
		refs = append(refs, lifecycle.ActiveRef{Table: "actors", Label: "actor", ID: *inv.ActorID})
	}
	return lifecycle.RequireAllActive(tx, refs)
}
