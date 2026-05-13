package actor

import (
	"fmt"
	"strings"
	"time"

	"github.com/devpablocristo/core/errors/go/domainerr"
	"github.com/devpablocristo/ponti-backend/internal/shared/authz"
	"gorm.io/gorm"
)

const (
	LegacyCustomers           = "customers"
	LegacyInvestors           = "investors"
	LegacyManagers            = "managers"
	LegacyProviders           = "providers"
	LegacyWorkOrderContractor = "workorders.contractor"
	LegacyInvoiceCompany      = "invoices.company"
	LegacyLaborContractor     = "labors.contractor_name"

	RoleCliente     = "cliente"
	RoleInversor    = "inversor"
	RoleResponsable = "responsable"
	RoleProveedor   = "proveedor"
	RoleContratista = "contratista"
	RoleFacturador  = "facturador"

	KindOrganization = "organization"
	KindPerson       = "natural_person"
	KindUnknown      = "unknown"
)

type LegacyActorSync struct {
	SourceTable string
	SourceID    int64
	Name        string
	ActorKind   string
	Role        string
	ArchivedAt  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CreatedBy   *string
	UpdatedBy   *string
	DeletedBy   *string
}

type LegacyTextActorSync struct {
	SourceTable string
	Name        string
	ActorKind   string
	Role        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CreatedBy   *string
	UpdatedBy   *string
}

func SyncLegacyActor(tx *gorm.DB, input LegacyActorSync) (int64, error) {
	if tx == nil {
		return 0, domainerr.Internal("database transaction is required")
	}
	if actorSyncDisabled(tx) {
		return 0, nil
	}
	if input.SourceID <= 0 {
		return 0, domainerr.Validation("legacy source id is required")
	}
	input.SourceTable = strings.TrimSpace(input.SourceTable)
	input.Name = strings.TrimSpace(input.Name)
	if input.SourceTable == "" || input.Name == "" || input.Role == "" {
		return 0, domainerr.Validation("legacy actor sync requires source table, name and role")
	}
	if input.ActorKind == "" {
		input.ActorKind = KindUnknown
	}
	if input.CreatedAt.IsZero() {
		input.CreatedAt = time.Now()
	}
	if input.UpdatedAt.IsZero() {
		input.UpdatedAt = time.Now()
	}

	tenantID, err := defaultTenantIDTx(tx)
	if err != nil {
		return 0, err
	}

	sourceKey := fmt.Sprintf("%d", input.SourceID)
	var actorID int64
	if err := tx.Table("legacy_actor_map").
		Where("tenant_id = ? AND source_table = ? AND source_key = ?", tenantID, input.SourceTable, sourceKey).
		Select("actor_id").
		Scan(&actorID).Error; err != nil {
		return 0, domainerr.Internal("failed to resolve legacy actor map")
	}

	if actorID == 0 {
		if err := tx.Raw(`
			INSERT INTO actors (
				tenant_id, actor_kind, display_name, normalized_name, archived_at,
				created_at, updated_at, created_by, updated_by, deleted_by
			)
			VALUES (?, ?, ?, public.normalize_actor_name(?), ?, ?, ?, ?, ?, ?)
			RETURNING id
		`, tenantID, input.ActorKind, input.Name, input.Name, input.ArchivedAt, input.CreatedAt, input.UpdatedAt, input.CreatedBy, input.UpdatedBy, input.DeletedBy).
			Scan(&actorID).Error; err != nil {
			return 0, domainerr.Internal("failed to create legacy actor")
		}
		if err := tx.Exec(`
			INSERT INTO legacy_actor_map (
				tenant_id, source_table, source_id, source_text, source_key, actor_id, confidence, mapping_status
			)
			VALUES (?, ?, ?, ?, ?, ?, 1.0, 'created_new')
			ON CONFLICT (tenant_id, source_table, source_key)
			DO UPDATE SET source_id = EXCLUDED.source_id, source_text = EXCLUDED.source_text, actor_id = EXCLUDED.actor_id
		`, tenantID, input.SourceTable, input.SourceID, input.Name, sourceKey, actorID).Error; err != nil {
			return 0, domainerr.Internal("failed to save legacy actor map")
		}
	} else {
		if err := tx.Exec(`
			UPDATE actors
			SET actor_kind = ?,
				display_name = ?,
				normalized_name = public.normalize_actor_name(?),
				updated_at = ?,
				updated_by = ?,
				deleted_at = NULL,
				deleted_by = NULL
			WHERE id = ?
		`, input.ActorKind, input.Name, input.Name, input.UpdatedAt, input.UpdatedBy, actorID).Error; err != nil {
			return 0, domainerr.Internal("failed to update legacy actor")
		}
		if err := tx.Exec(`
			UPDATE legacy_actor_map
			SET source_id = ?, source_text = ?
			WHERE tenant_id = ? AND source_table = ? AND source_key = ?
		`, input.SourceID, input.Name, tenantID, input.SourceTable, sourceKey).Error; err != nil {
			return 0, domainerr.Internal("failed to update legacy actor map")
		}
	}

	if err := upsertActorRole(tx, actorID, input.Role, input.ArchivedAt); err != nil {
		return 0, err
	}
	if err := refreshActorArchiveState(tx, actorID, input.ArchivedAt); err != nil {
		return 0, err
	}
	if err := refreshLegacyActorColumns(tx, tenantID, input.SourceTable, input.SourceID, actorID); err != nil {
		return 0, err
	}

	return actorID, nil
}

func SyncLegacyTextActor(tx *gorm.DB, input LegacyTextActorSync) (int64, error) {
	if tx == nil {
		return 0, domainerr.Internal("database transaction is required")
	}
	if actorSyncDisabled(tx) {
		return 0, nil
	}
	input.SourceTable = strings.TrimSpace(input.SourceTable)
	input.Name = strings.TrimSpace(input.Name)
	if input.SourceTable == "" || input.Name == "" || input.Role == "" {
		return 0, domainerr.Validation("legacy text actor sync requires source table, name and role")
	}
	if input.ActorKind == "" {
		input.ActorKind = KindUnknown
	}
	if input.CreatedAt.IsZero() {
		input.CreatedAt = time.Now()
	}
	if input.UpdatedAt.IsZero() {
		input.UpdatedAt = time.Now()
	}

	tenantID, err := defaultTenantIDTx(tx)
	if err != nil {
		return 0, err
	}

	var sourceKey string
	if err := tx.Raw(`SELECT public.normalize_actor_name(?)`, input.Name).Scan(&sourceKey).Error; err != nil {
		return 0, domainerr.Internal("failed to normalize legacy text actor")
	}
	if sourceKey == "" {
		return 0, domainerr.Validation("legacy text actor normalized name is empty")
	}

	var actorID int64
	if err := tx.Table("legacy_actor_map").
		Where("tenant_id = ? AND source_table = ? AND source_key = ?", tenantID, input.SourceTable, sourceKey).
		Select("actor_id").
		Scan(&actorID).Error; err != nil {
		return 0, domainerr.Internal("failed to resolve legacy text actor map")
	}

	if actorID == 0 {
		if err := tx.Raw(`
			INSERT INTO actors (
				tenant_id, actor_kind, display_name, normalized_name,
				created_at, updated_at, created_by, updated_by
			)
			VALUES (?, ?, ?, public.normalize_actor_name(?), ?, ?, ?, ?)
			RETURNING id
		`, tenantID, input.ActorKind, input.Name, input.Name, input.CreatedAt, input.UpdatedAt, input.CreatedBy, input.UpdatedBy).
			Scan(&actorID).Error; err != nil {
			return 0, domainerr.Internal("failed to create legacy text actor")
		}
		if err := tx.Exec(`
			INSERT INTO legacy_actor_map (
				tenant_id, source_table, source_text, source_key, actor_id, confidence, mapping_status
			)
			VALUES (?, ?, ?, ?, ?, 0.6, 'manual_review')
			ON CONFLICT (tenant_id, source_table, source_key)
			DO UPDATE SET source_text = EXCLUDED.source_text, actor_id = EXCLUDED.actor_id
		`, tenantID, input.SourceTable, input.Name, sourceKey, actorID).Error; err != nil {
			return 0, domainerr.Internal("failed to save legacy text actor map")
		}
	} else {
		if err := tx.Exec(`
			UPDATE actors
			SET display_name = ?,
				normalized_name = public.normalize_actor_name(?),
				updated_at = ?,
				updated_by = ?
			WHERE id = ?
		`, input.Name, input.Name, input.UpdatedAt, input.UpdatedBy, actorID).Error; err != nil {
			return 0, domainerr.Internal("failed to update legacy text actor")
		}
	}

	if err := upsertActorRole(tx, actorID, input.Role, nil); err != nil {
		return 0, err
	}
	return actorID, nil
}

func DeleteLegacyActor(tx *gorm.DB, sourceTable string, sourceID int64, role string, deletedBy *string) error {
	if actorSyncDisabled(tx) {
		return nil
	}
	actorID, err := ActorIDForLegacy(tx, sourceTable, sourceID)
	if err != nil || actorID == 0 {
		return err
	}
	now := time.Now()
	if err := tx.Exec(`
		UPDATE actor_roles
		SET archived_at = ?
		WHERE actor_id = ? AND role = ?
	`, now, actorID, role).Error; err != nil {
		return domainerr.Internal("failed to archive deleted legacy actor role")
	}
	var activeRoles int64
	if err := tx.Table("actor_roles").
		Where("actor_id = ? AND archived_at IS NULL", actorID).
		Count(&activeRoles).Error; err != nil {
		return domainerr.Internal("failed to check actor roles")
	}
	if activeRoles == 0 {
		if err := tx.Exec(`
			UPDATE actors
			SET archived_at = COALESCE(archived_at, ?),
				deleted_at = ?,
				deleted_by = ?,
				updated_at = ?
			WHERE id = ?
		`, now, now, deletedBy, now, actorID).Error; err != nil {
			return domainerr.Internal("failed to mark legacy actor deleted")
		}
	}
	return nil
}

func ActorIDForLegacy(tx *gorm.DB, sourceTable string, sourceID int64) (int64, error) {
	if tx == nil {
		return 0, domainerr.Internal("database transaction is required")
	}
	if actorSyncDisabled(tx) {
		return 0, nil
	}
	tenantID, err := defaultTenantIDTx(tx)
	if err != nil {
		return 0, err
	}
	var actorID int64
	if err := tx.Table("legacy_actor_map").
		Where("tenant_id = ? AND source_table = ? AND source_key = ?", tenantID, sourceTable, fmt.Sprintf("%d", sourceID)).
		Select("actor_id").
		Scan(&actorID).Error; err != nil {
		return 0, domainerr.Internal("failed to resolve legacy actor")
	}
	return actorID, nil
}

func RefreshProjectActorMirrors(tx *gorm.DB, projectID int64) error {
	if actorSyncDisabled(tx) {
		return nil
	}
	if projectID <= 0 {
		return domainerr.Validation("project_id is required")
	}
	var tenantID string
	if err := tx.Table("projects").
		Select("tenant_id").
		Where("id = ? AND deleted_at IS NULL", projectID).
		Scan(&tenantID).Error; err != nil {
		return domainerr.Internal("failed to resolve project tenant")
	}
	if strings.TrimSpace(tenantID) == "" {
		return domainerr.NotFound("project not found")
	}
	if err := tx.Exec(`
		UPDATE projects p
		SET customer_actor_id = m.actor_id
		FROM legacy_actor_map m
		WHERE p.id = ?
		  AND p.tenant_id = ?
		  AND m.source_table = 'customers'
		  AND m.source_id = p.customer_id
		  AND m.tenant_id = p.tenant_id
	`, projectID, tenantID).Error; err != nil {
		return domainerr.Internal("failed to refresh project customer actor")
	}
	if err := tx.Exec(`
		DELETE FROM project_responsibles
		WHERE project_id IN (SELECT id FROM projects WHERE id = ? AND tenant_id = ?)
	`, projectID, tenantID).Error; err != nil {
		return domainerr.Internal("failed to clear project responsibles actors")
	}
	if err := tx.Exec(`
		INSERT INTO project_responsibles (project_id, actor_id, created_at, updated_at, created_by, updated_by)
		SELECT pm.project_id, m.actor_id, COALESCE(pm.created_at, now()), COALESCE(pm.updated_at, now()), pm.created_by::text, pm.updated_by::text
		FROM project_managers pm
		JOIN legacy_actor_map m ON m.source_table = 'managers' AND m.source_id = pm.manager_id AND m.tenant_id = pm.tenant_id
		WHERE pm.project_id = ? AND pm.tenant_id = ? AND pm.deleted_at IS NULL
		ON CONFLICT DO NOTHING
	`, projectID, tenantID).Error; err != nil {
		return domainerr.Internal("failed to refresh project responsibles actors")
	}
	if err := tx.Exec(`
		DELETE FROM project_investor_allocations
		WHERE project_id IN (SELECT id FROM projects WHERE id = ? AND tenant_id = ?)
	`, projectID, tenantID).Error; err != nil {
		return domainerr.Internal("failed to clear project investor actors")
	}
	if err := tx.Exec(`
		INSERT INTO project_investor_allocations (project_id, actor_id, percentage, created_at, updated_at, created_by, updated_by)
		SELECT pi.project_id, m.actor_id, pi.percentage, COALESCE(pi.created_at, now()), COALESCE(pi.updated_at, now()), pi.created_by::text, pi.updated_by::text
		FROM project_investors pi
		JOIN legacy_actor_map m ON m.source_table = 'investors' AND m.source_id = pi.investor_id AND m.tenant_id = pi.tenant_id
		WHERE pi.project_id = ? AND pi.tenant_id = ? AND pi.deleted_at IS NULL
		ON CONFLICT DO NOTHING
	`, projectID, tenantID).Error; err != nil {
		return domainerr.Internal("failed to refresh project investor actors")
	}
	if err := tx.Exec(`
		DELETE FROM project_admin_cost_allocations
		WHERE project_id IN (SELECT id FROM projects WHERE id = ? AND tenant_id = ?)
	`, projectID, tenantID).Error; err != nil {
		return domainerr.Internal("failed to clear project admin cost actor allocations")
	}
	if err := tx.Exec(`
		INSERT INTO project_admin_cost_allocations (project_id, actor_id, percentage, created_at, updated_at, created_by, updated_by)
		SELECT aci.project_id, m.actor_id, aci.percentage, COALESCE(aci.created_at, now()), COALESCE(aci.updated_at, now()), aci.created_by::text, aci.updated_by::text
		FROM admin_cost_investors aci
		JOIN legacy_actor_map m ON m.source_table = 'investors' AND m.source_id = aci.investor_id AND m.tenant_id = aci.tenant_id
		WHERE aci.project_id = ? AND aci.tenant_id = ? AND aci.deleted_at IS NULL
		ON CONFLICT DO NOTHING
	`, projectID, tenantID).Error; err != nil {
		return domainerr.Internal("failed to refresh project admin cost actors")
	}
	if err := tx.Exec(`
		DELETE FROM field_lease_participants
		WHERE field_id IN (SELECT id FROM fields WHERE project_id = ? AND tenant_id = ?)
	`, projectID, tenantID).Error; err != nil {
		return domainerr.Internal("failed to clear field lease actor participants")
	}
	if err := tx.Exec(`
		INSERT INTO field_lease_participants (field_id, actor_id, percentage, created_at, updated_at, created_by, updated_by)
		SELECT fi.field_id, m.actor_id, fi.percentage, COALESCE(fi.created_at, now()), COALESCE(fi.updated_at, now()), fi.created_by::text, fi.updated_by::text
		FROM field_investors fi
		JOIN fields f ON f.id = fi.field_id AND f.tenant_id = fi.tenant_id
		JOIN legacy_actor_map m ON m.source_table = 'investors' AND m.source_id = fi.investor_id AND m.tenant_id = fi.tenant_id
		WHERE f.project_id = ? AND fi.tenant_id = ? AND fi.deleted_at IS NULL
		ON CONFLICT DO NOTHING
	`, projectID, tenantID).Error; err != nil {
		return domainerr.Internal("failed to refresh field lease actors")
	}
	return nil
}

func RefreshSupplyMovementActorColumns(tx *gorm.DB, movementID int64) error {
	if actorSyncDisabled(tx) {
		return nil
	}
	if movementID <= 0 {
		return domainerr.Validation("movement_id is required")
	}
	if err := tx.Exec(`
		UPDATE supply_movements sm
		SET investor_actor_id = (
				SELECT actor_id FROM legacy_actor_map
				WHERE source_table = 'investors' AND source_id = sm.investor_id
				  AND tenant_id = sm.tenant_id
				LIMIT 1
			),
			provider_actor_id = (
				SELECT actor_id FROM legacy_actor_map
				WHERE source_table = 'providers' AND source_id = sm.provider_id
				  AND tenant_id = sm.tenant_id
				LIMIT 1
			)
		WHERE sm.id = ?
	`, movementID).Error; err != nil {
		return domainerr.Internal("failed to refresh supply movement actor columns")
	}
	return nil
}

func RefreshWorkOrderActorColumns(tx *gorm.DB, workOrderID int64) error {
	if actorSyncDisabled(tx) {
		return nil
	}
	if workOrderID <= 0 {
		return domainerr.Validation("work_order_id is required")
	}
	if err := tx.Exec(`
		UPDATE workorders wo
		SET investor_actor_id = (
				SELECT actor_id FROM legacy_actor_map
				WHERE source_table = 'investors' AND source_id = wo.investor_id
				  AND tenant_id = wo.tenant_id
				LIMIT 1
			),
			contractor_actor_id = (
				SELECT actor_id FROM legacy_actor_map
				WHERE source_table = 'workorders.contractor'
				  AND source_key = public.normalize_actor_name(wo.contractor)
				  AND tenant_id = wo.tenant_id
				LIMIT 1
			),
			contractor_name_snapshot = NULLIF(wo.contractor, '')
		WHERE wo.id = ?
	`, workOrderID).Error; err != nil {
		return domainerr.Internal("failed to refresh work order actor columns")
	}
	if err := tx.Exec(`
		UPDATE workorder_investor_splits wis
		SET actor_id = (
				SELECT actor_id FROM legacy_actor_map
				WHERE source_table = 'investors' AND source_id = wis.investor_id
				  AND tenant_id = wis.tenant_id
				LIMIT 1
			)
		WHERE wis.workorder_id = ?
	`, workOrderID).Error; err != nil {
		return domainerr.Internal("failed to refresh work order split actor columns")
	}
	return nil
}

func RefreshInvoiceActorColumns(tx *gorm.DB, invoiceID int64) error {
	if actorSyncDisabled(tx) {
		return nil
	}
	if invoiceID <= 0 {
		return domainerr.Validation("invoice_id is required")
	}
	if err := tx.Exec(`
		UPDATE invoices i
		SET investor_actor_id = (
				SELECT actor_id FROM legacy_actor_map
				WHERE source_table = 'investors' AND source_id = i.investor_id
				  AND tenant_id = i.tenant_id
				LIMIT 1
			),
			company_actor_id = (
				SELECT actor_id FROM legacy_actor_map
				WHERE source_table = 'invoices.company'
				  AND source_key = public.normalize_actor_name(i.company)
				  AND tenant_id = i.tenant_id
				LIMIT 1
			),
			company_name_snapshot = NULLIF(i.company, '')
		WHERE i.id = ?
	`, invoiceID).Error; err != nil {
		return domainerr.Internal("failed to refresh invoice actor columns")
	}
	return nil
}

func RefreshStockActorColumns(tx *gorm.DB, stockID int64) error {
	if actorSyncDisabled(tx) {
		return nil
	}
	if stockID <= 0 {
		return domainerr.Validation("stock_id is required")
	}
	if err := tx.Exec(`
		UPDATE stocks s
		SET investor_actor_id = m.actor_id
		FROM legacy_actor_map m
		WHERE s.id = ?
		  AND m.source_table = 'investors'
		  AND m.source_id = s.investor_id
		  AND m.tenant_id = s.tenant_id
	`, stockID).Error; err != nil {
		return domainerr.Internal("failed to refresh stock actor columns")
	}
	return nil
}

func upsertActorRole(tx *gorm.DB, actorID int64, role string, archivedAt *time.Time) error {
	if err := tx.Exec(`
		INSERT INTO actor_roles (actor_id, role, archived_at)
		VALUES (?, ?, ?)
		ON CONFLICT (actor_id, role)
		DO UPDATE SET archived_at = EXCLUDED.archived_at
	`, actorID, role, archivedAt).Error; err != nil {
		return domainerr.Internal("failed to sync actor role")
	}
	return nil
}

func refreshActorArchiveState(tx *gorm.DB, actorID int64, archivedAt *time.Time) error {
	if archivedAt == nil {
		return tx.Exec(`UPDATE actors SET archived_at = NULL WHERE id = ?`, actorID).Error
	}
	var activeRoles int64
	if err := tx.Table("actor_roles").
		Where("actor_id = ? AND archived_at IS NULL", actorID).
		Count(&activeRoles).Error; err != nil {
		return domainerr.Internal("failed to check actor roles")
	}
	if activeRoles == 0 {
		return tx.Exec(`UPDATE actors SET archived_at = ? WHERE id = ?`, archivedAt, actorID).Error
	}
	return nil
}

func refreshLegacyActorColumns(tx *gorm.DB, tenantID string, sourceTable string, sourceID int64, actorID int64) error {
	switch sourceTable {
	case LegacyCustomers:
		if err := tx.Exec(
			`UPDATE customers SET actor_id = ? WHERE id = ? AND tenant_id = ?`,
			actorID,
			sourceID,
			tenantID,
		).Error; err != nil {
			return err
		}
		return tx.Exec(
			`UPDATE projects SET customer_actor_id = ? WHERE customer_id = ? AND tenant_id = ?`,
			actorID,
			sourceID,
			tenantID,
		).Error
	case LegacyInvestors:
		if err := tx.Exec(
			`UPDATE workorders SET investor_actor_id = ? WHERE investor_id = ? AND tenant_id = ?`,
			actorID,
			sourceID,
			tenantID,
		).Error; err != nil {
			return err
		}
		if err := tx.Exec(
			`UPDATE workorder_investor_splits SET actor_id = ? WHERE investor_id = ? AND tenant_id = ?`,
			actorID,
			sourceID,
			tenantID,
		).Error; err != nil {
			return err
		}
		if err := tx.Exec(
			`UPDATE stocks SET investor_actor_id = ? WHERE investor_id = ? AND tenant_id = ?`,
			actorID,
			sourceID,
			tenantID,
		).Error; err != nil {
			return err
		}
		if err := tx.Exec(
			`UPDATE supply_movements SET investor_actor_id = ? WHERE investor_id = ? AND tenant_id = ?`,
			actorID,
			sourceID,
			tenantID,
		).Error; err != nil {
			return err
		}
		return tx.Exec(
			`UPDATE invoices SET investor_actor_id = ? WHERE investor_id = ? AND tenant_id = ?`,
			actorID,
			sourceID,
			tenantID,
		).Error
	case LegacyProviders:
		return tx.Exec(
			`UPDATE supply_movements SET provider_actor_id = ? WHERE provider_id = ? AND tenant_id = ?`,
			actorID,
			sourceID,
			tenantID,
		).Error
	case LegacyManagers:
		return tx.Exec(`
			INSERT INTO project_responsibles (project_id, actor_id, created_at, updated_at, created_by, updated_by)
			SELECT pm.project_id, ?, COALESCE(pm.created_at, now()), COALESCE(pm.updated_at, now()), pm.created_by::text, pm.updated_by::text
			FROM project_managers pm
			WHERE pm.manager_id = ? AND pm.tenant_id = ? AND pm.deleted_at IS NULL
			ON CONFLICT DO NOTHING
		`, actorID, sourceID, tenantID).Error
	default:
		return nil
	}
}

func defaultTenantIDTx(tx *gorm.DB) (string, error) {
	if tenantID, ok := authz.TenantFromContext(tx.Statement.Context); ok {
		return tenantID.String(), nil
	}
	var tenantID string
	if err := tx.Table("auth_tenants").
		Where("name = ?", "default").
		Order("id").
		Select("id").
		Scan(&tenantID).Error; err != nil {
		return "", domainerr.Internal("failed to resolve default tenant")
	}
	if tenantID == "" {
		return "", domainerr.Internal("default tenant does not exist")
	}
	return tenantID, nil
}

func actorSyncDisabled(tx *gorm.DB) bool {
	return tx != nil && tx.Name() == "sqlite"
}
