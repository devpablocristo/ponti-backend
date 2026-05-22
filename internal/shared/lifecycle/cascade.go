package lifecycle

import (
	"fmt"
	"time"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func ListScopedIDs(tx *gorm.DB, table string, idColumn string, tenantID uuid.UUID, where string, args ...any) ([]int64, error) {
	if tx == nil || table == "" || idColumn == "" {
		return nil, domainerr.Internal("failed to list lifecycle ids")
	}
	if !hasTable(tx, table) {
		return []int64{}, nil
	}
	query := tx.Table(table).Where(where, args...)
	if tenantID != uuid.Nil && hasColumn(tx, table, "tenant_id") {
		query = query.Where("tenant_id = ?", tenantID)
	}
	var ids []int64
	if err := query.Pluck(idColumn, &ids).Error; err != nil {
		return nil, domainerr.Internal(fmt.Sprintf("failed to list %s ids", table))
	}
	return ids, nil
}

func ArchiveScopedRows(tx *gorm.DB, table string, tenantID uuid.UUID, archivedAt time.Time, deletedBy *string, cause Cause, where string, args ...any) error {
	if tx == nil || table == "" {
		return domainerr.Internal("failed to archive lifecycle rows")
	}
	if !hasTable(tx, table) || !hasColumn(tx, table, "deleted_at") {
		return nil
	}
	update := tx.Table(table).Where(where, args...).Where("deleted_at IS NULL")
	if tenantID != uuid.Nil && hasColumn(tx, table, "tenant_id") {
		update = update.Where("tenant_id = ?", tenantID)
	}
	if err := update.Updates(ArchiveUpdates(tx, table, archivedAt, deletedBy, cause)).Error; err != nil {
		return domainerr.Internal(fmt.Sprintf("failed to archive %s", table))
	}
	return nil
}

func RestoreScopedRows(tx *gorm.DB, table string, tenantID uuid.UUID, restoredAt time.Time, cause Cause, where string, args ...any) error {
	if tx == nil || table == "" {
		return domainerr.Internal("failed to restore lifecycle rows")
	}
	if !hasTable(tx, table) || !hasColumn(tx, table, "deleted_at") {
		return nil
	}
	update := tx.Table(table).Where(where, args...).Where("deleted_at IS NOT NULL")
	update = ApplyCauseScope(update, table, cause)
	if tenantID != uuid.Nil && hasColumn(tx, table, "tenant_id") {
		update = update.Where("tenant_id = ?", tenantID)
	}
	if err := update.Updates(RestoreUpdates(tx, table, restoredAt)).Error; err != nil {
		return domainerr.Internal(fmt.Sprintf("failed to restore %s", table))
	}
	return nil
}

func hasTable(tx *gorm.DB, table string) bool {
	return tx != nil && tx.Migrator().HasTable(table)
}

// RunCascadeArchive propagates an archive operation through an entity's
// Policy graph: CascadeTables get `ArchiveScopedRows` and ChildEntities
// recurse with the same `Cause`. The caller is responsible for archiving the
// root row itself; this helper only handles descendants. Use it when writing
// a new repository archive — the legacy per-repo cascade functions stay
// working in parallel and migrate opt-in.
//
// Name `RunCascadeArchive` avoids collision with the `CascadeArchive`
// RelationPolicy constant declared in policy.go.
func RunCascadeArchive(tx *gorm.DB, entity string, rootID int64, tenantID uuid.UUID,
	archivedAt time.Time, deletedBy *string, cause Cause) error {
	if tx == nil || rootID <= 0 || entity == "" {
		return nil
	}
	policy, ok := Policies[entity]
	if !ok {
		return nil
	}

	for _, ct := range policy.CascadeTables {
		if err := ArchiveScopedRows(tx, ct.Table, tenantID, archivedAt, deletedBy, cause,
			ct.ScopeColumn+" = ?", rootID); err != nil {
			return err
		}
	}

	for _, child := range policy.ChildEntities {
		if !hasTable(tx, child.Table) {
			continue
		}
		ids, err := ListScopedIDs(tx, child.Table, "id", tenantID,
			child.ScopeColumn+" = ? AND deleted_at IS NULL", rootID)
		if err != nil {
			return err
		}
		for _, id := range ids {
			if err := RunCascadeArchive(tx, child.Table, id, tenantID, archivedAt, deletedBy, cause); err != nil {
				return err
			}
		}
		if len(ids) > 0 {
			if err := ArchiveScopedRows(tx, child.Table, tenantID, archivedAt, deletedBy, cause,
				"id IN ?", ids); err != nil {
				return err
			}
		}
	}
	return nil
}

// RunCascadeRestore is the inverse of RunCascadeArchive: walks the Policy
// graph and restores rows archived by the same `Cause`. Rows archived by a
// different Cause stay archived — restoring a project does NOT revive a
// separately archived field.
func RunCascadeRestore(tx *gorm.DB, entity string, rootID int64, tenantID uuid.UUID,
	restoredAt time.Time, cause Cause) error {
	if tx == nil || rootID <= 0 || entity == "" {
		return nil
	}
	policy, ok := Policies[entity]
	if !ok {
		return nil
	}

	for _, ct := range policy.CascadeTables {
		if err := RestoreScopedRows(tx, ct.Table, tenantID, restoredAt, cause,
			ct.ScopeColumn+" = ?", rootID); err != nil {
			return err
		}
	}

	for _, child := range policy.ChildEntities {
		if !hasTable(tx, child.Table) {
			continue
		}
		query := tx.Table(child.Table).
			Where(child.ScopeColumn+" = ? AND deleted_at IS NOT NULL", rootID)
		query = ApplyCauseScope(query, child.Table, cause)
		if tenantID != uuid.Nil && hasColumn(tx, child.Table, "tenant_id") {
			query = query.Where("tenant_id = ?", tenantID)
		}
		var ids []int64
		if err := query.Pluck("id", &ids).Error; err != nil {
			return domainerr.Internal(fmt.Sprintf("failed to list archived %s children", child.Table))
		}
		for _, id := range ids {
			if err := RunCascadeRestore(tx, child.Table, id, tenantID, restoredAt, cause); err != nil {
				return err
			}
		}
		if len(ids) > 0 {
			if err := RestoreScopedRows(tx, child.Table, tenantID, restoredAt, cause, "id IN ?", ids); err != nil {
				return err
			}
		}
	}
	return nil
}

// WouldOrphanActiveChildren returns the count of active rows that reference
// `rootID` across all CascadeTables and ChildEntities of `entity`. Used by
// `BlockIfActiveChildren` archive flows to surface a precise count before
// refusing to archive. Walks one level (does not recurse) — the invariant is
// "are there active refs to me?", not "is my full subtree clean?".
func WouldOrphanActiveChildren(tx *gorm.DB, entity string, rootID int64,
	tenantID uuid.UUID) (int64, error) {
	if tx == nil || rootID <= 0 || entity == "" {
		return 0, nil
	}
	policy, ok := Policies[entity]
	if !ok {
		return 0, nil
	}
	var total int64
	walk := func(table, column string) error {
		if !hasTable(tx, table) || !hasColumn(tx, table, column) || !hasColumn(tx, table, "deleted_at") {
			return nil
		}
		var n int64
		q := tx.Table(table).Where(column+" = ? AND deleted_at IS NULL", rootID)
		if tenantID != uuid.Nil && hasColumn(tx, table, "tenant_id") {
			q = q.Where("tenant_id = ?", tenantID)
		}
		if err := q.Count(&n).Error; err != nil {
			return domainerr.Internal(fmt.Sprintf("failed to count %s references", table))
		}
		total += n
		return nil
	}
	for _, ct := range policy.CascadeTables {
		if err := walk(ct.Table, ct.ScopeColumn); err != nil {
			return 0, err
		}
	}
	for _, child := range policy.ChildEntities {
		if err := walk(child.Table, child.ScopeColumn); err != nil {
			return 0, err
		}
	}
	return total, nil
}
