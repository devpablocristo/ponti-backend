package actor

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/devpablocristo/core/errors/go/domainerr"
	"gorm.io/gorm"

	sharedtext "github.com/devpablocristo/ponti-backend/internal/shared/text"
)

type EnsureCustomerInput struct {
	CustomerID int64
	ActorID    *int64
	Name       string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	CreatedBy  *string
	UpdatedBy  *string
}

type EnsureCustomerResult struct {
	CustomerID int64
	ActorID    int64
	Name       string
}

type EnsureLegacyEntityInput struct {
	SourceTable string
	ActorID     *int64
	Name        string
	ActorKind   string
	Role        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CreatedBy   *string
	UpdatedBy   *string
}

func EnsureCustomerFromActor(tx *gorm.DB, input EnsureCustomerInput) (*EnsureCustomerResult, error) {
	if tx == nil {
		return nil, domainerr.Internal("database transaction is required")
	}
	if actorSyncDisabled(tx) {
		return nil, nil
	}
	input.Name = sharedtext.CanonicalizeName(input.Name)
	if input.Name == "" && (input.ActorID == nil || *input.ActorID <= 0) {
		return nil, domainerr.Validation("customer name or actor_id is required")
	}
	if input.CreatedAt.IsZero() {
		input.CreatedAt = time.Now()
	}
	if input.UpdatedAt.IsZero() {
		input.UpdatedAt = time.Now()
	}
	tenantID, err := defaultTenantIDTx(tx)
	if err != nil {
		return nil, err
	}

	actorID, actorName, err := resolveActorForName(tx, tenantID, input.ActorID, input.Name, KindOrganization)
	if err != nil {
		return nil, err
	}
	if input.Name == "" {
		input.Name = actorName
	}
	if err := ensureActorRole(tx, actorID, RoleCliente); err != nil {
		return nil, err
	}

	customer, err := findCustomerForActor(tx, tenantID, actorID)
	if err != nil {
		return nil, err
	}
	if customer.ID == 0 && input.CustomerID > 0 {
		customer, err = findCustomerByID(tx, tenantID, input.CustomerID)
		if err != nil {
			return nil, err
		}
	}
	if customer.ID == 0 {
		customer, err = findCustomerByExactName(tx, tenantID, input.Name)
		if err != nil {
			return nil, err
		}
	}
	if customer.ID == 0 {
		customer, err = createCustomerForActor(tx, tenantID, actorID, input)
		if err != nil {
			return nil, err
		}
	} else {
		if err := attachCustomerToActor(tx, tenantID, customer.ID, actorID, input.Name, customer.DeletedAt.Valid, input.UpdatedAt, input.UpdatedBy); err != nil {
			return nil, err
		}
		if customer.Name == "" {
			customer.Name = input.Name
		}
	}
	if err := upsertLegacyActorMap(tx, tenantID, LegacyCustomers, customer.ID, input.Name, actorID, 1, "auto_matched"); err != nil {
		return nil, err
	}
	if err := refreshLegacyActorColumns(tx, tenantID, LegacyCustomers, customer.ID, actorID); err != nil {
		return nil, err
	}

	return &EnsureCustomerResult{CustomerID: customer.ID, ActorID: actorID, Name: customer.Name}, nil
}

func EnsureLegacyEntityFromActor(tx *gorm.DB, input EnsureLegacyEntityInput) (int64, error) {
	if tx == nil {
		return 0, domainerr.Internal("database transaction is required")
	}
	if actorSyncDisabled(tx) {
		return 0, nil
	}
	table, err := legacyTableName(input.SourceTable)
	if err != nil {
		return 0, err
	}
	input.Name = sharedtext.CanonicalizeName(input.Name)
	if input.Name == "" && (input.ActorID == nil || *input.ActorID <= 0) {
		return 0, domainerr.Validation("legacy entity name or actor_id is required")
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
	actorID, actorName, err := resolveActorForName(tx, tenantID, input.ActorID, input.Name, input.ActorKind)
	if err != nil {
		return 0, err
	}
	if input.Name == "" {
		input.Name = actorName
	}
	if err := ensureActorRole(tx, actorID, input.Role); err != nil {
		return 0, err
	}

	sourceID, err := findLegacySourceForActor(tx, tenantID, input.SourceTable, actorID)
	if err != nil {
		return 0, err
	}
	if sourceID == 0 {
		sourceID, err = findLegacySourceByExactName(tx, tenantID, table, input.Name)
		if err != nil {
			return 0, err
		}
	}
	if sourceID == 0 {
		sourceID, err = createLegacySource(tx, tenantID, table, input.Name, input.CreatedBy, input.UpdatedBy)
		if err != nil {
			return 0, err
		}
	}
	if err := upsertLegacyActorMap(tx, tenantID, input.SourceTable, sourceID, input.Name, actorID, 1, "auto_matched"); err != nil {
		return 0, err
	}
	if err := refreshLegacyActorColumns(tx, tenantID, input.SourceTable, sourceID, actorID); err != nil {
		return 0, err
	}
	return sourceID, nil
}

func LinkLegacyEntityToActor(tx *gorm.DB, input LegacyActorSync, actorID int64) (int64, error) {
	if tx == nil {
		return 0, domainerr.Internal("database transaction is required")
	}
	if actorSyncDisabled(tx) {
		return 0, nil
	}
	if input.SourceID <= 0 || actorID <= 0 {
		return 0, domainerr.Validation("legacy source id and actor_id are required")
	}
	input.Name = sharedtext.CanonicalizeName(input.Name)
	if input.Name == "" {
		return 0, domainerr.Validation("legacy name is required")
	}
	if input.UpdatedAt.IsZero() {
		input.UpdatedAt = time.Now()
	}
	tenantID, err := defaultTenantIDTx(tx)
	if err != nil {
		return 0, err
	}
	if err := requireActorForLink(tx, tenantID, actorID); err != nil {
		return 0, err
	}
	if err := ensureActorRole(tx, actorID, input.Role); err != nil {
		return 0, err
	}
	if input.SourceTable == LegacyCustomers {
		if err := attachCustomerToActor(tx, tenantID, input.SourceID, actorID, input.Name, false, input.UpdatedAt, input.UpdatedBy); err != nil {
			return 0, err
		}
	}
	if err := upsertLegacyActorMap(tx, tenantID, input.SourceTable, input.SourceID, input.Name, actorID, 1, "auto_matched"); err != nil {
		return 0, err
	}
	if err := refreshLegacyActorColumns(tx, tenantID, input.SourceTable, input.SourceID, actorID); err != nil {
		return 0, err
	}
	return actorID, nil
}

type customerActorRow struct {
	ID        int64
	Name      string
	DeletedAt sql.NullTime
}

type actorNameRow struct {
	ID          int64
	DisplayName string
}

func resolveActorForName(tx *gorm.DB, tenantID string, actorID *int64, name string, actorKind string) (int64, string, error) {
	if actorID != nil && *actorID > 0 {
		var row actorNameRow
		if err := tx.Raw(`
			SELECT id, display_name
			FROM actors
			WHERE id = ? AND tenant_id = ? AND deleted_at IS NULL AND merged_into_actor_id IS NULL
			LIMIT 1
		`, *actorID, tenantID).Scan(&row).Error; err != nil {
			return 0, "", domainerr.Internal("failed to load actor")
		}
		if row.ID == 0 {
			return 0, "", domainerr.New(domainerr.KindNotFound, fmt.Sprintf("actor with id %d does not exist", *actorID))
		}
		if err := tx.Exec(`UPDATE actors SET archived_at = NULL, updated_at = now() WHERE id = ?`, row.ID).Error; err != nil {
			return 0, "", domainerr.Internal("failed to restore actor")
		}
		return row.ID, row.DisplayName, nil
	}

	rows := []actorNameRow{}
	if err := tx.Raw(`
		SELECT id, display_name
		FROM actors
		WHERE tenant_id = ?
		  AND deleted_at IS NULL
		  AND merged_into_actor_id IS NULL
		  AND normalized_name = public.normalize_actor_name(?)
		ORDER BY CASE WHEN archived_at IS NULL THEN 0 ELSE 1 END, id
		LIMIT 2
	`, tenantID, name).Scan(&rows).Error; err != nil {
		return 0, "", domainerr.Internal("failed to search actor")
	}
	if len(rows) > 1 {
		return 0, "", domainerr.New(domainerr.KindConflict, "multiple actors match this name; select one explicitly")
	}
	if len(rows) == 1 {
		if err := tx.Exec(`UPDATE actors SET archived_at = NULL, updated_at = now() WHERE id = ?`, rows[0].ID).Error; err != nil {
			return 0, "", domainerr.Internal("failed to restore actor")
		}
		return rows[0].ID, rows[0].DisplayName, nil
	}

	if actorKind == "" {
		actorKind = KindUnknown
	}
	var createdID int64
	if err := tx.Raw(`
		INSERT INTO actors (tenant_id, actor_kind, display_name, normalized_name, created_at, updated_at)
		VALUES (?, ?, ?, public.normalize_actor_name(?), now(), now())
		RETURNING id
	`, tenantID, actorKind, name, name).Scan(&createdID).Error; err != nil {
		return 0, "", domainerr.Internal("failed to create actor")
	}
	return createdID, name, nil
}

func ensureActorRole(tx *gorm.DB, actorID int64, role string) error {
	role = strings.TrimSpace(role)
	if role == "" {
		return domainerr.Validation("actor role is required")
	}
	if err := tx.Exec(`
		INSERT INTO actor_roles (actor_id, role, created_at, archived_at)
		VALUES (?, ?, now(), NULL)
		ON CONFLICT (actor_id, role)
		DO UPDATE SET archived_at = NULL
	`, actorID, role).Error; err != nil {
		return domainerr.Internal("failed to ensure actor role")
	}
	return nil
}

func requireActorForLink(tx *gorm.DB, tenantID string, actorID int64) error {
	var count int64
	if err := tx.Table("actors").
		Where("id = ? AND tenant_id = ? AND deleted_at IS NULL AND merged_into_actor_id IS NULL", actorID, tenantID).
		Count(&count).Error; err != nil {
		return domainerr.Internal("failed to validate actor")
	}
	if count == 0 {
		return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("actor with id %d does not exist", actorID))
	}
	return tx.Exec(`UPDATE actors SET archived_at = NULL, updated_at = now() WHERE id = ?`, actorID).Error
}

func findCustomerForActor(tx *gorm.DB, tenantID string, actorID int64) (customerActorRow, error) {
	var row customerActorRow
	if err := tx.Raw(`
		SELECT id, name, deleted_at
		FROM customers
		WHERE tenant_id = ? AND actor_id = ?
		ORDER BY CASE WHEN deleted_at IS NULL THEN 0 ELSE 1 END, id
		LIMIT 1
	`, tenantID, actorID).Scan(&row).Error; err != nil {
		return row, domainerr.Internal("failed to find customer for actor")
	}
	if row.ID != 0 {
		return row, nil
	}
	if err := tx.Raw(`
		SELECT c.id, c.name, c.deleted_at
		FROM legacy_actor_map m
		JOIN customers c ON c.id = m.source_id AND c.tenant_id = m.tenant_id
		WHERE m.tenant_id = ? AND m.source_table = 'customers' AND m.actor_id = ?
		ORDER BY CASE WHEN c.deleted_at IS NULL THEN 0 ELSE 1 END, c.id
		LIMIT 1
	`, tenantID, actorID).Scan(&row).Error; err != nil {
		return row, domainerr.Internal("failed to find mapped customer")
	}
	return row, nil
}

func findCustomerByID(tx *gorm.DB, tenantID string, customerID int64) (customerActorRow, error) {
	var row customerActorRow
	if err := tx.Raw(`
		SELECT id, name, deleted_at
		FROM customers
		WHERE tenant_id = ? AND id = ?
		LIMIT 1
	`, tenantID, customerID).Scan(&row).Error; err != nil {
		return row, domainerr.Internal("failed to find customer")
	}
	return row, nil
}

func findCustomerByExactName(tx *gorm.DB, tenantID string, name string) (customerActorRow, error) {
	var row customerActorRow
	if err := tx.Raw(`
		SELECT id, name, deleted_at
		FROM customers
		WHERE tenant_id = ? AND name = ?
		ORDER BY CASE WHEN deleted_at IS NULL THEN 0 ELSE 1 END, id
		LIMIT 1
	`, tenantID, name).Scan(&row).Error; err != nil {
		return row, domainerr.Internal("failed to find customer by name")
	}
	return row, nil
}

func createCustomerForActor(tx *gorm.DB, tenantID string, actorID int64, input EnsureCustomerInput) (customerActorRow, error) {
	var row customerActorRow
	if err := tx.Raw(`
		INSERT INTO customers (tenant_id, name, actor_id, created_at, updated_at, created_by, updated_by)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		RETURNING id, name, deleted_at
	`, tenantID, input.Name, actorID, input.CreatedAt, input.UpdatedAt, input.CreatedBy, input.UpdatedBy).Scan(&row).Error; err != nil {
		return row, domainerr.Internal("failed to create customer")
	}
	return row, nil
}

func attachCustomerToActor(tx *gorm.DB, tenantID string, customerID int64, actorID int64, name string, restore bool, updatedAt time.Time, updatedBy *string) error {
	updates := map[string]any{
		"actor_id":   actorID,
		"updated_at": updatedAt,
		"updated_by": updatedBy,
	}
	if restore {
		updates["deleted_at"] = nil
		updates["deleted_by"] = nil
	}
	canonicalName := sharedtext.CanonicalizeName(name)
	if canonicalName != "" {
		updates["name"] = canonicalName
	}
	if err := tx.Table("customers").
		Where("tenant_id = ? AND id = ?", tenantID, customerID).
		Updates(updates).Error; err != nil {
		return domainerr.Internal("failed to link customer to actor")
	}
	if canonicalName != "" {
		if err := renameActorDisplayName(tx, actorID, canonicalName, updatedAt, updatedBy); err != nil {
			return err
		}
	}
	return nil
}

func renameActorDisplayName(tx *gorm.DB, actorID int64, name string, updatedAt time.Time, updatedBy *string) error {
	if actorID <= 0 || sharedtext.CanonicalizeName(name) == "" {
		return nil
	}
	if err := tx.Exec(`
		UPDATE actors
		SET display_name = ?,
			normalized_name = public.normalize_actor_name(?),
			updated_at = ?,
			updated_by = ?
		WHERE id = ?
		  AND (display_name <> ? OR normalized_name <> public.normalize_actor_name(?))
	`, name, name, updatedAt, updatedBy, actorID, name, name).Error; err != nil {
		return domainerr.Internal("failed to rename actor")
	}
	return nil
}

func findLegacySourceForActor(tx *gorm.DB, tenantID string, sourceTable string, actorID int64) (int64, error) {
	var id int64
	if err := tx.Table("legacy_actor_map").
		Where("tenant_id = ? AND source_table = ? AND actor_id = ?", tenantID, sourceTable, actorID).
		Order("source_id").
		Limit(1).
		Select("source_id").
		Scan(&id).Error; err != nil {
		return 0, domainerr.Internal("failed to find legacy entity for actor")
	}
	return id, nil
}

func findLegacySourceByExactName(tx *gorm.DB, tenantID string, table string, name string) (int64, error) {
	var id int64
	if err := tx.Table(table).
		Where("tenant_id = ? AND name = ? AND deleted_at IS NULL", tenantID, name).
		Order("id").
		Limit(1).
		Select("id").
		Scan(&id).Error; err != nil {
		return 0, domainerr.Internal("failed to find legacy entity by name")
	}
	return id, nil
}

func createLegacySource(tx *gorm.DB, tenantID string, table string, name string, createdBy *string, updatedBy *string) (int64, error) {
	var id int64
	query := fmt.Sprintf(`
		INSERT INTO %s (tenant_id, name, created_at, updated_at, created_by, updated_by)
		VALUES (?, ?, now(), now(), ?, ?)
		RETURNING id
	`, table)
	if err := tx.Raw(query, tenantID, name, createdBy, updatedBy).Scan(&id).Error; err != nil {
		return 0, domainerr.Internal("failed to create legacy entity")
	}
	return id, nil
}

func upsertLegacyActorMap(tx *gorm.DB, tenantID string, sourceTable string, sourceID int64, sourceText string, actorID int64, confidence float64, status string) error {
	if err := tx.Exec(`
		INSERT INTO legacy_actor_map (
			tenant_id, source_table, source_id, source_text, source_key, actor_id, confidence, mapping_status
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (tenant_id, source_table, source_key)
		DO UPDATE SET source_id = EXCLUDED.source_id,
		              source_text = EXCLUDED.source_text,
		              actor_id = EXCLUDED.actor_id,
		              confidence = EXCLUDED.confidence,
		              mapping_status = EXCLUDED.mapping_status
	`, tenantID, sourceTable, sourceID, sourceText, fmt.Sprintf("%d", sourceID), actorID, confidence, status).Error; err != nil {
		return domainerr.Internal("failed to save legacy actor map")
	}
	return nil
}

func legacyTableName(sourceTable string) (string, error) {
	switch sourceTable {
	case LegacyManagers:
		return "managers", nil
	case LegacyInvestors:
		return "investors", nil
	case LegacyProviders:
		return "providers", nil
	default:
		return "", domainerr.Validation("unsupported legacy actor source")
	}
}
