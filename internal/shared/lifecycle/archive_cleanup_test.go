package lifecycle

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newArchiveCleanupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	stmts := []string{
		`CREATE TABLE archive_batches (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT, root_entity TEXT NOT NULL, root_id INTEGER NOT NULL,
			action TEXT, reason TEXT, created_by TEXT, created_at DATETIME
		)`,
		archiveCleanupEntityTableSQL("customers", "actor_id INTEGER"),
		archiveCleanupEntityTableSQL("projects", "customer_id INTEGER"),
		archiveCleanupEntityTableSQL("fields", "project_id INTEGER"),
		archiveCleanupEntityTableSQL("lots", "field_id INTEGER"),
		archiveCleanupEntityTableSQL("workorders", "project_id INTEGER, field_id INTEGER, lot_id INTEGER"),
		archiveCleanupEntityTableSQL("labors", "project_id INTEGER"),
		archiveCleanupEntityTableSQL("supplies", "project_id INTEGER"),
		archiveCleanupEntityTableSQL("supply_movements", "project_id INTEGER, supply_id INTEGER"),
		archiveCleanupEntityTableSQL("stocks", "project_id INTEGER, supply_id INTEGER"),
		archiveCleanupEntityTableSQL("workorder_items", "workorder_id INTEGER"),
		archiveCleanupEntityTableSQL("workorder_investor_splits", "workorder_id INTEGER, investor_id INTEGER"),
		archiveCleanupEntityTableSQL("managers", ""),
		archiveCleanupEntityTableSQL("investors", ""),
		archiveCleanupEntityTableSQL("actors", ""),
		`CREATE TABLE project_managers (
			id INTEGER PRIMARY KEY, tenant_id TEXT, project_id INTEGER, manager_id INTEGER,
			deleted_at DATETIME, archive_batch_id INTEGER, archive_origin_entity TEXT,
			archive_origin_id INTEGER, archive_reason TEXT
		)`,
		`CREATE TABLE project_investors (
			id INTEGER PRIMARY KEY, tenant_id TEXT, project_id INTEGER, investor_id INTEGER,
			deleted_at DATETIME, archive_batch_id INTEGER, archive_origin_entity TEXT,
			archive_origin_id INTEGER, archive_reason TEXT
		)`,
		`CREATE TABLE field_investors (
			id INTEGER PRIMARY KEY, tenant_id TEXT, field_id INTEGER, investor_id INTEGER,
			deleted_at DATETIME, archive_batch_id INTEGER, archive_origin_entity TEXT,
			archive_origin_id INTEGER, archive_reason TEXT
		)`,
		`CREATE TABLE legacy_actor_map (
			id INTEGER PRIMARY KEY, tenant_id TEXT, source_table TEXT, source_id INTEGER, actor_id INTEGER
		)`,
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("schema: %v", err)
		}
	}
	return db
}

func archiveCleanupEntityTableSQL(table, extraColumns string) string {
	if extraColumns != "" {
		extraColumns += ","
	}
	return `CREATE TABLE ` + table + ` (
		id INTEGER PRIMARY KEY, tenant_id TEXT, ` + extraColumns + `
		deleted_at DATETIME, deleted_by TEXT, updated_at DATETIME,
		archive_batch_id INTEGER, archive_origin_entity TEXT,
		archive_origin_id INTEGER, archive_reason TEXT
	)`
}

func TestArchiveCleanupDryRunDoesNotMutate(t *testing.T) {
	db := newArchiveCleanupTestDB(t)
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000101")
	now := time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC)
	if err := db.Exec(`
		INSERT INTO customers (id, tenant_id, deleted_at) VALUES (1, ?, ?);
		INSERT INTO projects (id, tenant_id, customer_id) VALUES (10, ?, 1);
	`, tenantID.String(), now, tenantID.String()).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}

	report, err := RunArchiveCleanup(t.Context(), db, ArchiveCleanupOptions{TenantID: tenantID, Now: now})
	if err != nil {
		t.Fatalf("RunArchiveCleanup dry-run: %v", err)
	}
	if report.Mode != "dry-run" {
		t.Fatalf("expected dry-run report, got %q", report.Mode)
	}

	var active int64
	if err := db.Raw(`SELECT COUNT(*) FROM projects WHERE id = 10 AND deleted_at IS NULL`).Scan(&active).Error; err != nil {
		t.Fatalf("count project: %v", err)
	}
	if active != 1 {
		t.Fatalf("dry-run mutated project")
	}
	if !reportHasAction(report, "IA-1", operationArchiveChild, "projects") {
		t.Fatalf("expected IA-1 archive action in report: %+v", report.Actions)
	}
}

func TestArchiveCleanupApplyArchivesChildWithParentCause(t *testing.T) {
	db := newArchiveCleanupTestDB(t)
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000102")
	now := time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC)
	reason := "customer archived"
	createdBy := "tester"
	cause, err := RootCause(db, tenantID, "customers", 1, &reason, &createdBy)
	if err != nil {
		t.Fatalf("RootCause: %v", err)
	}
	if err := db.Exec(`
		INSERT INTO customers (id, tenant_id, deleted_at, archive_batch_id, archive_origin_entity, archive_origin_id, archive_reason)
			VALUES (1, ?, ?, ?, ?, ?, ?);
		INSERT INTO projects (id, tenant_id, customer_id) VALUES (10, ?, 1);
	`, tenantID.String(), now, cause.BatchID, cause.OriginEntity, cause.OriginID, reason, tenantID.String()).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}

	report, err := RunArchiveCleanup(t.Context(), db, ArchiveCleanupOptions{Apply: true, TenantID: tenantID, Now: now})
	if err != nil {
		t.Fatalf("RunArchiveCleanup apply: %v", err)
	}
	if hasAutoRemediableViolations(report.Checks) {
		t.Fatalf("expected clean post-checks, got %+v", report.Checks)
	}

	var row struct {
		DeletedAt           *time.Time
		ArchiveBatchID      int64
		ArchiveOriginEntity string
		ArchiveOriginID     int64
	}
	if err := db.Table("projects").
		Select("deleted_at, archive_batch_id, archive_origin_entity, archive_origin_id").
		Where("id = 10").
		Scan(&row).Error; err != nil {
		t.Fatalf("read project: %v", err)
	}
	if row.DeletedAt == nil {
		t.Fatalf("project should be archived")
	}
	if row.ArchiveBatchID != cause.BatchID || row.ArchiveOriginEntity != "customers" || row.ArchiveOriginID != 1 {
		t.Fatalf("project did not inherit parent cause: %+v want batch=%d customers:1", row, cause.BatchID)
	}
}

func TestArchiveCleanupBackfillsLegacyParentAndUsesCauseForChild(t *testing.T) {
	db := newArchiveCleanupTestDB(t)
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000103")
	now := time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC)
	if err := db.Exec(`
		INSERT INTO customers (id, tenant_id, deleted_at) VALUES (1, ?, ?);
		INSERT INTO projects (id, tenant_id, customer_id) VALUES (10, ?, 1);
	`, tenantID.String(), now, tenantID.String()).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}

	report, err := RunArchiveCleanup(t.Context(), db, ArchiveCleanupOptions{Apply: true, TenantID: tenantID, Now: now})
	if err != nil {
		t.Fatalf("RunArchiveCleanup apply: %v", err)
	}
	if !reportHasAction(report, "IA-14", operationBackfill, "customers") {
		t.Fatalf("expected IA-14 backfill action: %+v", report.Actions)
	}

	var rows []struct {
		TableName       string
		ArchiveBatchID  int64
		ArchiveOrigin   string
		ArchiveOriginID int64
		ArchiveReason   string
	}
	if err := db.Raw(`
		SELECT 'customers' AS table_name, archive_batch_id, archive_origin_entity AS archive_origin,
			archive_origin_id, archive_reason FROM customers WHERE id = 1
		UNION ALL
		SELECT 'projects', archive_batch_id, archive_origin_entity, archive_origin_id, archive_reason
			FROM projects WHERE id = 10
	`).Scan(&rows).Error; err != nil {
		t.Fatalf("read rows: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected customer and project rows, got %+v", rows)
	}
	if rows[0].ArchiveBatchID == 0 || rows[0].ArchiveBatchID != rows[1].ArchiveBatchID {
		t.Fatalf("expected same backfilled batch on parent and child, got %+v", rows)
	}
	for _, row := range rows {
		if row.ArchiveOrigin != "customers" || row.ArchiveOriginID != 1 || row.ArchiveReason != legacyArchiveReason {
			t.Fatalf("unexpected archive metadata on %s: %+v", row.TableName, row)
		}
	}
}

func TestArchiveCleanupApplyIsIdempotent(t *testing.T) {
	db := newArchiveCleanupTestDB(t)
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000104")
	now := time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC)
	if err := db.Exec(`
		INSERT INTO customers (id, tenant_id, deleted_at) VALUES (1, ?, ?);
		INSERT INTO projects (id, tenant_id, customer_id) VALUES (10, ?, 1);
	`, tenantID.String(), now, tenantID.String()).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}

	if _, err := RunArchiveCleanup(t.Context(), db, ArchiveCleanupOptions{Apply: true, TenantID: tenantID, Now: now}); err != nil {
		t.Fatalf("first apply: %v", err)
	}
	report, err := RunArchiveCleanup(t.Context(), db, ArchiveCleanupOptions{Apply: true, TenantID: tenantID, Now: now})
	if err != nil {
		t.Fatalf("second apply: %v", err)
	}
	if len(report.Actions) != 0 {
		t.Fatalf("expected idempotent second apply with no actions, got %+v", report.Actions)
	}
	if hasAutoRemediableViolations(report.Checks) {
		t.Fatalf("expected clean post-checks, got %+v", report.Checks)
	}
}

func TestArchiveCleanupApplyBlocksManualReviewRefs(t *testing.T) {
	db := newArchiveCleanupTestDB(t)
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000105")
	now := time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC)
	if err := db.Exec(`
		INSERT INTO investors (id, tenant_id, deleted_at) VALUES (7, ?, ?);
		INSERT INTO project_investors (id, tenant_id, project_id, investor_id) VALUES (70, ?, 1, 7);
	`, tenantID.String(), now, tenantID.String()).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}

	report, err := RunArchiveCleanup(t.Context(), db, ArchiveCleanupOptions{Apply: true, TenantID: tenantID, Now: now})
	if !errors.Is(err, ErrArchiveCleanupManualReview) {
		t.Fatalf("expected manual-review error, got %v", err)
	}
	if len(report.Blockers) == 0 || report.Blockers[1].Rows != 1 {
		t.Fatalf("expected IA-11b blocker, got %+v", report.Blockers)
	}
	var batches int64
	if err := db.Raw(`SELECT COUNT(*) FROM archive_batches`).Scan(&batches).Error; err != nil {
		t.Fatalf("count batches: %v", err)
	}
	if batches != 0 {
		t.Fatalf("apply with blockers should not mutate data, created %d batches", batches)
	}
}

func reportHasAction(report ArchiveCleanupReport, checkID, operation, table string) bool {
	for _, action := range report.Actions {
		if action.CheckID == checkID && action.Operation == operation && action.Table == table {
			return true
		}
	}
	return false
}
