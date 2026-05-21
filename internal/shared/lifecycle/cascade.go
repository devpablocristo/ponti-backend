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
