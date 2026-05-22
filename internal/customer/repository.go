// Package customer implementa el repositorio de clientes.
package customer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	actorsync "github.com/devpablocristo/ponti-backend/internal/actor"
	models "github.com/devpablocristo/ponti-backend/internal/customer/repository/models"
	domain "github.com/devpablocristo/ponti-backend/internal/customer/usecases/domain"
	"github.com/devpablocristo/ponti-backend/internal/shared/authz"
	"github.com/devpablocristo/ponti-backend/internal/shared/lifecycle"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
	sharedrepo "github.com/devpablocristo/ponti-backend/internal/shared/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
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

func (r *Repository) CreateCustomer(ctx context.Context, c *domain.Customer) (int64, error) {
	if err := sharedrepo.ValidateEntity(c, "customer"); err != nil {
		return 0, err
	}
	var id int64
	err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := assertCustomerReferencesActive(tx, c); err != nil {
			return err
		}
		result, err := actorsync.EnsureCustomerFromActor(tx, actorsync.EnsureCustomerInput{
			ActorID:   c.ActorID,
			Name:      c.Name,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
			CreatedBy: c.CreatedBy,
			UpdatedBy: c.UpdatedBy,
		})
		if err != nil {
			return err
		}
		if result != nil {
			id = result.CustomerID
			c.ActorID = &result.ActorID
			return nil
		}

		model := models.FromDomain(c)
		model.Base = sharedmodels.Base{
			CreatedBy: c.CreatedBy,
			UpdatedBy: c.UpdatedBy,
		}
		if tenantID, ok, err := authz.OptionalTenantOrStrict(ctx); err != nil {
			return err
		} else if ok {
			model.TenantID = tenantID
		}
		if err := tx.Create(model).Error; err != nil {
			return domainerr.Internal("failed to create customer")
		}
		id = model.ID
		return nil
	})
	return id, err
}

type listedCustomerRow struct {
	ID      int64
	Name    string
	ActorID *int64
}

func (r *Repository) ListCustomers(ctx context.Context, page, perPage int) ([]domain.ListedCustomer, int64, error) {
	var rows []listedCustomerRow
	var total int64

	db0 := r.db.Client().WithContext(ctx).
		Table("customers c").
		Where("c.deleted_at IS NULL")
	db0 = authz.MaybeTenantScope(ctx, db0, "c")

	// Conteo total
	if err := db0.Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count customers")
	}

	// Consulta ligera: sólo id y name
	listDB := db0.
		Select("c.id, c.name, COALESCE(c.actor_id, m.actor_id) AS actor_id").
		Joins("LEFT JOIN legacy_actor_map m ON m.source_table = 'customers' AND m.source_id = c.id AND m.tenant_id = c.tenant_id")
	if err := listDB.
		Limit(perPage).
		Offset((page - 1) * perPage).
		Scan(&rows).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to list customers")
	}

	// Mapear a dominio ligero
	customers := make([]domain.ListedCustomer, len(rows))
	for i, m := range rows {
		customers[i] = domain.ListedCustomer{
			ID:      m.ID,
			Name:    m.Name,
			ActorID: m.ActorID,
		}
	}

	return customers, total, nil
}

func (r *Repository) ListArchivedCustomers(ctx context.Context, page, perPage int) ([]domain.ListedCustomer, int64, error) {
	var rows []listedCustomerRow
	var total int64

	db0 := r.db.Client().WithContext(ctx).
		Unscoped().
		Table("customers c").
		Where("c.deleted_at IS NOT NULL")
	db0 = authz.MaybeTenantScope(ctx, db0, "c")

	if err := db0.Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count archived customers")
	}

	if err := db0.
		Select("c.id, c.name, COALESCE(c.actor_id, m.actor_id) AS actor_id").
		Joins("LEFT JOIN legacy_actor_map m ON m.source_table = 'customers' AND m.source_id = c.id AND m.tenant_id = c.tenant_id").
		Limit(perPage).
		Offset((page - 1) * perPage).
		Scan(&rows).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to list archived customers")
	}

	customers := make([]domain.ListedCustomer, len(rows))
	for i, m := range rows {
		customers[i] = domain.ListedCustomer{
			ID:      m.ID,
			Name:    m.Name,
			ActorID: m.ActorID,
		}
	}

	return customers, total, nil
}

func (r *Repository) GetCustomer(ctx context.Context, id int64) (*domain.Customer, error) {
	var model models.Customer
	db0 := authz.MaybeTenantScope(ctx, r.db.Client().WithContext(ctx), "customers")
	err := db0.
		Where("id = ? AND deleted_at IS NULL", id).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerr.New(domainerr.KindNotFound, fmt.Sprintf("customer with id %d not found", id))
		}
		return nil, domainerr.Internal("failed to get customer")
	}
	out := model.ToDomain()
	if out.ActorID == nil {
		actorID, err := actorsync.ActorIDForLegacy(r.db.Client().WithContext(ctx), actorsync.LegacyCustomers, id)
		if err != nil {
			return nil, err
		}
		if actorID > 0 {
			out.ActorID = &actorID
		}
	}
	if out.ActorID != nil {
		actorID := *out.ActorID
		out.ActorID = &actorID
	}
	return out, nil
}

func (r *Repository) UpdateCustomer(ctx context.Context, c *domain.Customer) error {
	if err := sharedrepo.ValidateEntity(c, "customer"); err != nil {
		return err
	}
	if err := sharedrepo.ValidateID(c.ID, "customer"); err != nil {
		return err
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := assertCustomerReferencesActive(tx, c); err != nil {
			return err
		}
		updateTx := authz.MaybeTenantScope(ctx, tx.Model(&models.Customer{}), "customers").Where("id = ?", c.ID)
		if !c.UpdatedAt.IsZero() {
			updateTx = updateTx.Where("updated_at = ?", c.UpdatedAt)
		}
		result := updateTx.Updates(models.FromDomain(c))
		if result.Error != nil {
			return domainerr.Internal("failed to update customer")
		}
		if result.RowsAffected == 0 {
			if !c.UpdatedAt.IsZero() {
				return domainerr.Conflict("customer not found or outdated")
			}
			return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("customer with id %d does not exist", c.ID))
		}
		if c.ActorID != nil && *c.ActorID > 0 {
			_, err := actorsync.LinkLegacyEntityToActor(tx, actorsync.LegacyActorSync{
				SourceTable: actorsync.LegacyCustomers,
				SourceID:    c.ID,
				Name:        c.Name,
				ActorKind:   actorsync.KindOrganization,
				Role:        actorsync.RoleCliente,
				UpdatedAt:   time.Now(),
				UpdatedBy:   c.UpdatedBy,
			}, *c.ActorID)
			return err
		}
		_, err := actorsync.SyncLegacyActor(tx, actorsync.LegacyActorSync{
			SourceTable: actorsync.LegacyCustomers,
			SourceID:    c.ID,
			Name:        c.Name,
			ActorKind:   actorsync.KindOrganization,
			Role:        actorsync.RoleCliente,
			UpdatedAt:   time.Now(),
			UpdatedBy:   c.UpdatedBy,
		})
		return err
	})
}

func (r *Repository) ArchiveCustomer(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "customer"); err != nil {
		return err
	}
	actor, err := sharedmodels.ActorFromContext(ctx)
	if err != nil {
		return err
	}
	deletedBy := &actor

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var customer models.Customer
		customerQuery := authz.MaybeTenantScope(ctx, tx.Unscoped(), "customers")
		if err := customerQuery.Where("id = ?", id).First(&customer).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("customer %d not found", id))
			}
			return domainerr.Internal("failed to get customer")
		}
		if customer.DeletedAt.Valid {
			return domainerr.Conflict("customer already archived")
		}

		// Capture active project IDs before cascade so we can refresh actor
		// mirrors per project after the archive.
		var activeProjectIDs []int64
		activeProjectsQuery := tx.Table("projects").
			Where("customer_id = ? AND deleted_at IS NULL", id)
		if customer.TenantID != uuid.Nil {
			activeProjectsQuery = activeProjectsQuery.Where("tenant_id = ?", customer.TenantID)
		}
		if err := activeProjectsQuery.Pluck("id", &activeProjectIDs).Error; err != nil {
			return domainerr.Internal("failed to list active projects")
		}

		archivedAt := time.Now()
		batch, err := lifecycle.CreateArchiveBatch(tx, customer.TenantID, "customers", id, nil, deletedBy)
		if err != nil {
			return err
		}
		cause := lifecycle.CauseFromBatch(batch)

		// Centralized cascade: walks the Policies graph
		// (customers→projects→{fields→lots, workorders, drafts, labors,
		// supplies, supply_movements, stocks} + all CascadeTables like
		// project_managers/investors/admin_cost_investors/field_investors/
		// lot_dates/workorder_items/etc.) with this Cause. Replaces the
		// hardcoded `archiveCustomerProjects` + `archiveCustomerProjectChildren`
		// + `customerProjectScopedArchiveTables` triplet.
		if err := lifecycle.RunCascadeArchive(tx, "customers", id, customer.TenantID, archivedAt, deletedBy, cause); err != nil {
			return err
		}

		if err := authz.MaybeTenantScope(ctx, tx.Model(&models.Customer{}), "customers").
			Where("id = ?", id).
			Updates(lifecycle.ArchiveUpdates(tx, "customers", archivedAt, deletedBy, cause)).Error; err != nil {
			return domainerr.Internal("failed to archive customer")
		}
		if _, err := actorsync.SyncLegacyActor(tx, actorsync.LegacyActorSync{
			SourceTable: actorsync.LegacyCustomers,
			SourceID:    customer.ID,
			Name:        customer.Name,
			ActorKind:   actorsync.KindOrganization,
			Role:        actorsync.RoleCliente,
			ArchivedAt:  &archivedAt,
			UpdatedAt:   archivedAt,
			UpdatedBy:   customer.UpdatedBy,
			DeletedBy:   deletedBy,
		}); err != nil {
			return err
		}
		for _, projectID := range activeProjectIDs {
			if err := actorsync.RefreshProjectActorMirrors(tx, projectID); err != nil {
				return err
			}
		}
		return nil
	})
}

// customerProjectScopedArchiveTables is the list of pivots / project-scoped
// tables consumed by the legacy `restoreCustomerProjectGraph` (the archive
// side now goes through `lifecycle.RunCascadeArchive` via the Policies map).
// Restore migration to `RunCascadeRestore` is a follow-up — see plan §10C.
var customerProjectScopedArchiveTables = []string{
	"project_managers",
	"project_investors",
	"workorders",
	"work_order_drafts",
	"labors",
	"supplies",
	"supply_movements",
	"stocks",
	"crop_commercializations",
	"project_dollar_values",
	"admin_cost_investors",
}

// archiveCustomerProjectChildren — UNUSED after migration to RunCascadeArchive.
// Kept only as the inner helper of the legacy restoreCustomerProjectGraph;
// restore migration to RunCascadeRestore is a follow-up (see plan §10C).
func archiveCustomerProjectChildren(tx *gorm.DB, tenantID uuid.UUID, fieldIDs []int64, lotIDs []int64, workOrderIDs []int64, draftIDs []int64, archivedAt time.Time, deletedBy *string, cause lifecycle.Cause) error {
	if len(fieldIDs) > 0 {
		if err := lifecycle.ArchiveScopedRows(tx, "field_investors", tenantID, archivedAt, deletedBy, cause, "field_id IN ?", fieldIDs); err != nil {
			return err
		}
	}
	if len(lotIDs) > 0 {
		if err := lifecycle.ArchiveScopedRows(tx, "lot_dates", tenantID, archivedAt, deletedBy, cause, "lot_id IN ?", lotIDs); err != nil {
			return err
		}
	}
	if len(workOrderIDs) > 0 {
		if err := lifecycle.ArchiveScopedRows(tx, "workorder_items", tenantID, archivedAt, deletedBy, cause, "workorder_id IN ?", workOrderIDs); err != nil {
			return err
		}
		if err := lifecycle.ArchiveScopedRows(tx, "workorder_investor_splits", tenantID, archivedAt, deletedBy, cause, "workorder_id IN ?", workOrderIDs); err != nil {
			return err
		}
	}
	if len(draftIDs) > 0 {
		if err := lifecycle.ArchiveScopedRows(tx, "work_order_draft_items", tenantID, archivedAt, deletedBy, cause, "draft_id IN ?", draftIDs); err != nil {
			return err
		}
		if err := lifecycle.ArchiveScopedRows(tx, "work_order_draft_investor_splits", tenantID, archivedAt, deletedBy, cause, "draft_id IN ?", draftIDs); err != nil {
			return err
		}
	}
	return nil
}

func archiveProjectScopedCustomerTable(tx *gorm.DB, table string, projectIDs []int64, tenantID uuid.UUID, archivedAt time.Time, deletedBy *string, cause lifecycle.Cause) error {
	if !tx.Migrator().HasTable(table) {
		return nil
	}
	update := tx.Table(table).Where("project_id IN ? AND deleted_at IS NULL", projectIDs)
	if tenantID != uuid.Nil && tx.Migrator().HasColumn(table, "tenant_id") {
		update = update.Where("tenant_id = ?", tenantID)
	}
	if err := update.Updates(lifecycle.ArchiveUpdates(tx, table, archivedAt, deletedBy, cause)).Error; err != nil {
		return domainerr.Internal(fmt.Sprintf("failed to archive %s", table))
	}
	return nil
}

func (r *Repository) RestoreCustomer(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "customer"); err != nil {
		return err
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		restoredAt := time.Now()
		var customer models.Customer
		customerQuery := authz.MaybeTenantScope(ctx, tx.Unscoped(), "customers")
		if err := customerQuery.Where("id = ?", id).First(&customer).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("customer %d not found", id))
			}
			return domainerr.Internal("failed to get customer")
		}
		rowState, err := lifecycle.ReadRowState(tx, "customers", id)
		if err != nil {
			return err
		}
		cause := lifecycle.CauseFromRow(rowState, "customers", id)

		if !customer.DeletedAt.Valid {
			return restoreCustomerActiveProjectGraph(tx, id, customer.TenantID, restoredAt, cause)
		}

		if err := restoreCustomerProjects(tx, id, customer.TenantID, restoredAt, cause); err != nil {
			return err
		}

		if err := authz.MaybeTenantScope(ctx, tx.Unscoped().Model(&models.Customer{}), "customers").
			Where("id = ?", id).
			Updates(lifecycle.RestoreUpdates(tx, "customers", restoredAt)).Error; err != nil {
			return domainerr.Internal("failed to restore customer")
		}
		if _, err := actorsync.SyncLegacyActor(tx, actorsync.LegacyActorSync{
			SourceTable: actorsync.LegacyCustomers,
			SourceID:    customer.ID,
			Name:        customer.Name,
			ActorKind:   actorsync.KindOrganization,
			Role:        actorsync.RoleCliente,
			UpdatedAt:   restoredAt,
			UpdatedBy:   customer.UpdatedBy,
		}); err != nil {
			return err
		}
		return nil
	})
}

func restoreCustomerActiveProjectGraph(tx *gorm.DB, customerID int64, tenantID uuid.UUID, restoredAt time.Time, cause lifecycle.Cause) error {
	var projectIDs []int64
	projectLookup := tx.Table("projects").Where("customer_id = ? AND deleted_at IS NULL", customerID)
	if tenantID != uuid.Nil && tx.Migrator().HasColumn("projects", "tenant_id") {
		projectLookup = projectLookup.Where("tenant_id = ?", tenantID)
	}
	if err := projectLookup.Pluck("id", &projectIDs).Error; err != nil {
		return domainerr.Internal("failed to list active customer projects")
	}
	if len(projectIDs) == 0 {
		return nil
	}
	if err := restoreCustomerProjectGraph(tx, projectIDs, tenantID, restoredAt, cause); err != nil {
		return err
	}
	for _, projectID := range projectIDs {
		if err := actorsync.RefreshProjectActorMirrors(tx, projectID); err != nil {
			return err
		}
	}
	return nil
}

func restoreCustomerProjects(tx *gorm.DB, customerID int64, tenantID uuid.UUID, restoredAt time.Time, cause lifecycle.Cause) error {
	var projectIDs []int64
	projectLookup := tx.Table("projects").Where("customer_id = ? AND deleted_at IS NOT NULL", customerID)
	projectLookup = lifecycle.ApplyCauseScope(projectLookup, "projects", cause)
	if tenantID != uuid.Nil && tx.Migrator().HasColumn("projects", "tenant_id") {
		projectLookup = projectLookup.Where("tenant_id = ?", tenantID)
	}
	if err := projectLookup.Pluck("id", &projectIDs).Error; err != nil {
		return domainerr.Internal("failed to list archived projects")
	}
	if len(projectIDs) == 0 {
		return nil
	}

	if err := restoreCustomerProjectGraph(tx, projectIDs, tenantID, restoredAt, cause); err != nil {
		return err
	}

	projectUpdate := tx.Table("projects").Where("id IN ? AND deleted_at IS NOT NULL", projectIDs)
	projectUpdate = lifecycle.ApplyCauseScope(projectUpdate, "projects", cause)
	if tenantID != uuid.Nil && tx.Migrator().HasColumn("projects", "tenant_id") {
		projectUpdate = projectUpdate.Where("tenant_id = ?", tenantID)
	}
	if err := projectUpdate.Updates(lifecycle.RestoreUpdates(tx, "projects", restoredAt)).Error; err != nil {
		return domainerr.Internal("failed to restore customer projects")
	}

	for _, projectID := range projectIDs {
		if err := actorsync.RefreshProjectActorMirrors(tx, projectID); err != nil {
			return err
		}
	}

	return nil
}

func restoreCustomerProjectGraph(tx *gorm.DB, projectIDs []int64, tenantID uuid.UUID, restoredAt time.Time, cause lifecycle.Cause) error {
	// Pluck child IDs FIRST (filtering on `deleted_at IS NOT NULL` + Cause)
	// so the restore of the project-scoped tables below doesn't wipe out the
	// rows we need to find. Restore order:
	//   1. List ids of fields/lots/workorders/drafts archived with this Cause.
	//   2. Restore the project-scoped tables (workorders, drafts, labors,
	//      supplies, etc.) en bulk by project_id.
	//   3. Restore the items/splits using the captured IDs.
	//   4. Restore fields/lots explicitly (they live outside project_id scope).
	//
	// Previously step 2 ran first, which left the captured-ID lookup with
	// nothing to pluck (the rows it queried had just been moved to deleted_at
	// IS NULL). That hid `workorder_items` from cascade restore — bug
	// surfaced when `RunCascadeArchive` started archiving them properly.
	fieldIDs, err := restoreCustomerProjectChildIDs(tx, "fields", "id", tenantID, cause, "project_id IN ?", projectIDs)
	if err != nil {
		return err
	}
	var lotIDs []int64
	if len(fieldIDs) > 0 {
		lotIDs, err = restoreCustomerProjectChildIDs(tx, "lots", "id", tenantID, cause, "field_id IN ?", fieldIDs)
		if err != nil {
			return err
		}
	}
	workOrderIDs, err := restoreCustomerProjectChildIDs(tx, "workorders", "id", tenantID, cause, "project_id IN ?", projectIDs)
	if err != nil {
		return err
	}
	draftIDs, err := restoreCustomerProjectChildIDs(tx, "work_order_drafts", "id", tenantID, cause, "project_id IN ?", projectIDs)
	if err != nil {
		return err
	}

	for _, table := range customerProjectScopedArchiveTables {
		if err := restoreProjectScopedCustomerTable(tx, table, projectIDs, tenantID, restoredAt, cause); err != nil {
			return err
		}
	}

	if len(fieldIDs) > 0 {
		if err := lifecycle.RestoreScopedRows(tx, "field_investors", tenantID, restoredAt, cause, "field_id IN ?", fieldIDs); err != nil {
			return err
		}
		if err := lifecycle.RestoreScopedRows(tx, "fields", tenantID, restoredAt, cause, "id IN ?", fieldIDs); err != nil {
			return err
		}
	}
	if len(lotIDs) > 0 {
		if err := lifecycle.RestoreScopedRows(tx, "lot_dates", tenantID, restoredAt, cause, "lot_id IN ?", lotIDs); err != nil {
			return err
		}
		if err := lifecycle.RestoreScopedRows(tx, "lots", tenantID, restoredAt, cause, "id IN ?", lotIDs); err != nil {
			return err
		}
	}
	if len(workOrderIDs) > 0 {
		if err := lifecycle.RestoreScopedRows(tx, "workorder_items", tenantID, restoredAt, cause, "workorder_id IN ?", workOrderIDs); err != nil {
			return err
		}
		if err := lifecycle.RestoreScopedRows(tx, "workorder_investor_splits", tenantID, restoredAt, cause, "workorder_id IN ?", workOrderIDs); err != nil {
			return err
		}
	}
	if len(draftIDs) > 0 {
		if err := lifecycle.RestoreScopedRows(tx, "work_order_draft_items", tenantID, restoredAt, cause, "draft_id IN ?", draftIDs); err != nil {
			return err
		}
		if err := lifecycle.RestoreScopedRows(tx, "work_order_draft_investor_splits", tenantID, restoredAt, cause, "draft_id IN ?", draftIDs); err != nil {
			return err
		}
	}
	return nil
}

func restoreProjectScopedCustomerTable(tx *gorm.DB, table string, projectIDs []int64, tenantID uuid.UUID, restoredAt time.Time, cause lifecycle.Cause) error {
	if len(projectIDs) == 0 {
		return nil
	}
	return lifecycle.RestoreScopedRows(tx, table, tenantID, restoredAt, cause, "project_id IN ?", projectIDs)
}

func restoreCustomerProjectChildIDs(tx *gorm.DB, table string, idColumn string, tenantID uuid.UUID, cause lifecycle.Cause, where string, args ...any) ([]int64, error) {
	if !tx.Migrator().HasTable(table) {
		return []int64{}, nil
	}
	query := tx.Table(table).Where(where, args...).Where("deleted_at IS NOT NULL")
	query = lifecycle.ApplyCauseScope(query, table, cause)
	if tenantID != uuid.Nil && tx.Migrator().HasColumn(table, "tenant_id") {
		query = query.Where("tenant_id = ?", tenantID)
	}
	var ids []int64
	if err := query.Pluck(idColumn, &ids).Error; err != nil {
		return nil, domainerr.Internal("failed to list archived " + table)
	}
	return ids, nil
}

// HardDeleteCustomer elimina definitivamente un cliente.
// Bloquea con 409 si tiene proyectos (activos o archivados).
func (r *Repository) HardDeleteCustomer(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "customer"); err != nil {
		return err
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var customer models.Customer
		customerDB := authz.MaybeTenantScope(ctx, tx.Unscoped(), "customers")
		if err := customerDB.Where("id = ?", id).First(&customer).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("customer with id %d does not exist", id))
			}
			return domainerr.Internal("failed to check customer existence")
		}
		if !customer.DeletedAt.Valid {
			return domainerr.Conflict("customer must be archived before hard delete")
		}

		var activeCount, archivedCount int64
		if err := tx.Table("projects").
			Where("customer_id = ? AND deleted_at IS NULL", id).
			Where("tenant_id IN (SELECT tenant_id FROM customers WHERE id = ?)", id).
			Count(&activeCount).Error; err != nil {
			return domainerr.Internal("failed to count active projects")
		}
		if err := tx.Unscoped().Table("projects").
			Where("customer_id = ? AND deleted_at IS NOT NULL", id).
			Where("tenant_id IN (SELECT tenant_id FROM customers WHERE id = ?)", id).
			Count(&archivedCount).Error; err != nil {
			return domainerr.Internal("failed to count archived projects")
		}
		if activeCount+archivedCount > 0 {
			return domainerr.Conflict(fmt.Sprintf("customer has %d project(s) (%d active, %d archived); archive or hard-delete them first", activeCount+archivedCount, activeCount, archivedCount))
		}

		var deletedBy *string
		if actor, err := sharedmodels.ActorFromContext(ctx); err == nil {
			deletedBy = &actor
		}
		if err := actorsync.DeleteLegacyActor(tx, actorsync.LegacyCustomers, id, actorsync.RoleCliente, deletedBy); err != nil {
			return err
		}
		if err := authz.MaybeTenantScope(ctx, tx.Unscoped(), "customers").Delete(&models.Customer{}, "id = ?", id).Error; err != nil {
			return domainerr.Internal("failed to hard delete customer")
		}
		return nil
	})
}

// assertCustomerReferencesActive blocks Create/Update of a customer that
// references archived entities. The customer schema accepts archived actor
// rows at the DB level (no `WHERE deleted_at IS NULL` on the FK), so we
// enforce the invariant in the application layer inside the caller's
// transaction. Returns a Conflict domainerr with the English label "actor is
// archived" that `translateBackendError` maps to a Spanish toast on the FE.
func assertCustomerReferencesActive(tx *gorm.DB, c *domain.Customer) error {
	if c == nil {
		return nil
	}
	refs := []lifecycle.ActiveRef{}
	if c.ActorID != nil {
		refs = append(refs, lifecycle.ActiveRef{Table: "actors", Label: "actor", ID: *c.ActorID})
	}
	return lifecycle.RequireAllActive(tx, refs)
}

