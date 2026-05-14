package lifecycle

import (
	"fmt"
	"time"

	"github.com/devpablocristo/core/errors/go/domainerr"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ArchiveBatch struct {
	ID         int64      `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID   *uuid.UUID `gorm:"column:tenant_id;type:uuid"`
	RootEntity string     `gorm:"column:root_entity"`
	RootID     int64      `gorm:"column:root_id"`
	Action     string     `gorm:"column:action"`
	Reason     *string    `gorm:"column:reason"`
	CreatedBy  *string    `gorm:"column:created_by"`
	CreatedAt  time.Time  `gorm:"column:created_at"`
}

func (ArchiveBatch) TableName() string {
	return "archive_batches"
}

type Cause struct {
	BatchID      int64
	OriginEntity string
	OriginID     int64
	Reason       *string
}

type RowState struct {
	ArchiveBatchID      *int64
	ArchiveOriginEntity *string
	ArchiveOriginID     *int64
	ArchiveReason       *string
}

func CreateArchiveBatch(tx *gorm.DB, tenantID uuid.UUID, rootEntity string, rootID int64, reason *string, createdBy *string) (*ArchiveBatch, error) {
	if tx == nil {
		return nil, domainerr.Internal("failed to create archive batch")
	}
	if !tx.Migrator().HasTable("archive_batches") {
		return &ArchiveBatch{RootEntity: rootEntity, RootID: rootID, Action: "archive", Reason: reason, CreatedBy: createdBy, CreatedAt: time.Now()}, nil
	}
	var scopedTenant *uuid.UUID
	if tenantID != uuid.Nil {
		scopedTenant = &tenantID
	}
	row := &ArchiveBatch{
		TenantID:   scopedTenant,
		RootEntity: rootEntity,
		RootID:     rootID,
		Action:     "archive",
		Reason:     reason,
		CreatedBy:  createdBy,
		CreatedAt:  time.Now(),
	}
	if err := tx.Create(row).Error; err != nil {
		return nil, domainerr.Internal("failed to create archive batch")
	}
	return row, nil
}

func CauseFromBatch(batch *ArchiveBatch) Cause {
	if batch == nil {
		return Cause{}
	}
	return Cause{
		BatchID:      batch.ID,
		OriginEntity: batch.RootEntity,
		OriginID:     batch.RootID,
		Reason:       batch.Reason,
	}
}

func CauseFromRow(row RowState, fallbackEntity string, fallbackID int64) Cause {
	cause := Cause{OriginEntity: fallbackEntity, OriginID: fallbackID}
	if row.ArchiveBatchID != nil {
		cause.BatchID = *row.ArchiveBatchID
	}
	if row.ArchiveOriginEntity != nil && *row.ArchiveOriginEntity != "" {
		cause.OriginEntity = *row.ArchiveOriginEntity
	}
	if row.ArchiveOriginID != nil && *row.ArchiveOriginID > 0 {
		cause.OriginID = *row.ArchiveOriginID
	}
	cause.Reason = row.ArchiveReason
	return cause
}

func ArchiveUpdates(tx *gorm.DB, table string, archivedAt time.Time, deletedBy *string, cause Cause) map[string]any {
	updates := map[string]any{"deleted_at": archivedAt}
	if hasColumn(tx, table, "deleted_by") {
		updates["deleted_by"] = deletedBy
	}
	if hasColumn(tx, table, "updated_at") {
		updates["updated_at"] = archivedAt
	}
	if cause.BatchID > 0 && hasColumn(tx, table, "archive_batch_id") {
		updates["archive_batch_id"] = cause.BatchID
	}
	if cause.OriginEntity != "" && hasColumn(tx, table, "archive_origin_entity") {
		updates["archive_origin_entity"] = cause.OriginEntity
	}
	if cause.OriginID > 0 && hasColumn(tx, table, "archive_origin_id") {
		updates["archive_origin_id"] = cause.OriginID
	}
	if hasColumn(tx, table, "archive_reason") {
		updates["archive_reason"] = cause.Reason
	}
	return updates
}

func RestoreUpdates(tx *gorm.DB, table string, restoredAt time.Time) map[string]any {
	updates := map[string]any{"deleted_at": nil}
	if hasColumn(tx, table, "deleted_by") {
		updates["deleted_by"] = nil
	}
	if hasColumn(tx, table, "updated_at") {
		updates["updated_at"] = restoredAt
	}
	if hasColumn(tx, table, "archive_batch_id") {
		updates["archive_batch_id"] = nil
	}
	if hasColumn(tx, table, "archive_origin_entity") {
		updates["archive_origin_entity"] = nil
	}
	if hasColumn(tx, table, "archive_origin_id") {
		updates["archive_origin_id"] = nil
	}
	if hasColumn(tx, table, "archive_reason") {
		updates["archive_reason"] = nil
	}
	return updates
}

func ApplyCauseScope(tx *gorm.DB, table string, cause Cause) *gorm.DB {
	if cause.BatchID <= 0 || !hasColumn(tx, table, "archive_batch_id") {
		return tx.Where("1 = 0")
	}
	tx = tx.Where("archive_batch_id = ?", cause.BatchID)
	if cause.OriginEntity != "" && hasColumn(tx, table, "archive_origin_entity") {
		tx = tx.Where("archive_origin_entity = ?", cause.OriginEntity)
	}
	if cause.OriginID > 0 && hasColumn(tx, table, "archive_origin_id") {
		tx = tx.Where("archive_origin_id = ?", cause.OriginID)
	}
	return tx
}

func ReadRowState(tx *gorm.DB, table string, id int64) (RowState, error) {
	var row RowState
	if tx == nil || id == 0 {
		return row, nil
	}
	if !hasColumn(tx, table, "archive_batch_id") {
		return row, nil
	}
	err := tx.Table(table).
		Select("archive_batch_id, archive_origin_entity, archive_origin_id, archive_reason").
		Where("id = ?", id).
		Scan(&row).Error
	if err != nil {
		return RowState{}, domainerr.Internal("failed to read archive metadata")
	}
	return row, nil
}

func IsArchived(tx *gorm.DB, table string, id int64) (bool, error) {
	if tx == nil || table == "" || id == 0 {
		return false, domainerr.Internal("failed to read lifecycle state")
	}
	var row struct {
		DeletedAt *time.Time `gorm:"column:deleted_at"`
	}
	if err := tx.Table(table).Select("deleted_at").Where("id = ?", id).Scan(&row).Error; err != nil {
		return false, domainerr.Internal("failed to read lifecycle state")
	}
	return row.DeletedAt != nil, nil
}

func RequireArchived(tx *gorm.DB, table string, label string, id int64) error {
	archived, err := IsArchived(tx, table, id)
	if err != nil {
		return err
	}
	if !archived {
		if label == "" {
			label = table
		}
		return domainerr.Conflict(fmt.Sprintf("%s must be archived before hard delete", label))
	}
	return nil
}

func RootCause(tx *gorm.DB, tenantID uuid.UUID, rootEntity string, rootID int64, reason *string, deletedBy *string) (Cause, error) {
	batch, err := CreateArchiveBatch(tx, tenantID, rootEntity, rootID, reason, deletedBy)
	if err != nil {
		return Cause{}, err
	}
	return CauseFromBatch(batch), nil
}

func hasColumn(tx *gorm.DB, table string, column string) bool {
	return tx != nil && tx.Migrator().HasColumn(table, column)
}
