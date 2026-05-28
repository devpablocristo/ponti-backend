package lifecycle

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newCascadeTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	// Simplified schema mirroring the parent→child→pivot relations used by
	// the `projects` Policy. tenant_id is omitted; archive metadata columns
	// match what ArchiveUpdates/RestoreUpdates write.
	stmts := []string{
		`CREATE TABLE projects (
			id INTEGER PRIMARY KEY, customer_id INTEGER, campaign_id INTEGER,
			deleted_at DATETIME, deleted_by TEXT, updated_at DATETIME,
			archive_batch_id INTEGER, archive_origin_entity TEXT,
			archive_origin_id INTEGER, archive_reason TEXT
		)`,
		`CREATE TABLE fields (
			id INTEGER PRIMARY KEY, project_id INTEGER,
			deleted_at DATETIME, deleted_by TEXT, updated_at DATETIME,
			archive_batch_id INTEGER, archive_origin_entity TEXT,
			archive_origin_id INTEGER, archive_reason TEXT
		)`,
		`CREATE TABLE lots (
			id INTEGER PRIMARY KEY, field_id INTEGER,
			deleted_at DATETIME, deleted_by TEXT, updated_at DATETIME,
			archive_batch_id INTEGER, archive_origin_entity TEXT,
			archive_origin_id INTEGER, archive_reason TEXT
		)`,
		`CREATE TABLE project_managers (
			project_id INTEGER, manager_id INTEGER,
			deleted_at DATETIME, deleted_by TEXT, updated_at DATETIME,
			archive_batch_id INTEGER, archive_origin_entity TEXT,
			archive_origin_id INTEGER, archive_reason TEXT
		)`,
		`CREATE TABLE field_investors (
			field_id INTEGER, investor_id INTEGER,
			deleted_at DATETIME, deleted_by TEXT, updated_at DATETIME,
			archive_batch_id INTEGER, archive_origin_entity TEXT,
			archive_origin_id INTEGER, archive_reason TEXT
		)`,
		// Required by ArchiveUpdates path
		`CREATE TABLE archive_batches (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT, root_entity TEXT NOT NULL, root_id INTEGER NOT NULL,
			action TEXT, reason TEXT, created_by TEXT, created_at DATETIME
		)`,
	}
	for _, s := range stmts {
		if err := db.Exec(s).Error; err != nil {
			t.Fatalf("schema: %v", err)
		}
	}
	return db
}

func seedProjectTree(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.Exec(`
		INSERT INTO projects (id, customer_id, campaign_id) VALUES (1, 10, 20);
		INSERT INTO fields (id, project_id) VALUES (1, 1), (2, 1);
		INSERT INTO lots (id, field_id) VALUES (1, 1), (2, 1), (3, 2);
		INSERT INTO project_managers (project_id, manager_id) VALUES (1, 100);
		INSERT INTO field_investors (field_id, investor_id) VALUES (1, 200), (2, 201);
	`).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}
}

func TestRunCascadeArchive_PropagatesToChildrenAndPivots(t *testing.T) {
	t.Parallel()
	db := newCascadeTestDB(t)
	seedProjectTree(t, db)

	deletedBy := "tester"
	archivedAt := time.Now().UTC().Truncate(time.Microsecond)
	cause, err := RootCause(db, uuid.Nil, "projects", 1, nil, &deletedBy)
	if err != nil {
		t.Fatalf("RootCause: %v", err)
	}

	if err := RunCascadeArchive(db, "projects", 1, uuid.Nil, archivedAt, &deletedBy, cause); err != nil {
		t.Fatalf("RunCascadeArchive: %v", err)
	}

	// Every descendant must now have deleted_at NOT NULL and share the same
	// archive_batch_id.
	checks := []struct {
		name, query string
	}{
		{"fields archived", "SELECT COUNT(*) FROM fields WHERE deleted_at IS NULL"},
		{"lots archived", "SELECT COUNT(*) FROM lots WHERE deleted_at IS NULL"},
		{"project_managers archived", "SELECT COUNT(*) FROM project_managers WHERE deleted_at IS NULL"},
		{"field_investors archived", "SELECT COUNT(*) FROM field_investors WHERE deleted_at IS NULL"},
	}
	for _, c := range checks {
		var n int64
		if err := db.Raw(c.query).Scan(&n).Error; err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if n != 0 {
			t.Fatalf("expected 0 active rows after cascade for %s, got %d", c.name, n)
		}
	}

	// All descendants share the same batch_id.
	var batches []int64
	if err := db.Raw(`SELECT DISTINCT archive_batch_id FROM fields WHERE archive_batch_id IS NOT NULL
		UNION SELECT DISTINCT archive_batch_id FROM lots WHERE archive_batch_id IS NOT NULL
		UNION SELECT DISTINCT archive_batch_id FROM project_managers WHERE archive_batch_id IS NOT NULL
		UNION SELECT DISTINCT archive_batch_id FROM field_investors WHERE archive_batch_id IS NOT NULL`).
		Scan(&batches).Error; err != nil {
		t.Fatalf("batches: %v", err)
	}
	if len(batches) != 1 {
		t.Fatalf("expected single batch across descendants, got %v", batches)
	}
}

func TestRunCascadeRestore_OnlyRestoresSameCause(t *testing.T) {
	t.Parallel()
	db := newCascadeTestDB(t)
	seedProjectTree(t, db)

	deletedBy := "tester"
	now := time.Now().UTC().Truncate(time.Microsecond)
	cause, err := RootCause(db, uuid.Nil, "projects", 1, nil, &deletedBy)
	if err != nil {
		t.Fatalf("RootCause: %v", err)
	}

	if err := RunCascadeArchive(db, "projects", 1, uuid.Nil, now, &deletedBy, cause); err != nil {
		t.Fatalf("RunCascadeArchive: %v", err)
	}
	// Independently archive a lot with a different Cause (simulating
	// someone archiving lot 3 by hand before).
	otherCause := Cause{BatchID: 999, OriginEntity: "lots", OriginID: 3}
	if err := db.Exec(`UPDATE lots SET deleted_at = ?, archive_batch_id = ?,
		archive_origin_entity = ?, archive_origin_id = ?
		WHERE id = 3`, now, otherCause.BatchID, otherCause.OriginEntity, otherCause.OriginID).Error; err != nil {
		t.Fatalf("indep archive: %v", err)
	}

	// Restore with the project's Cause — should leave lot 3 archived.
	if err := RunCascadeRestore(db, "projects", 1, uuid.Nil, now, cause); err != nil {
		t.Fatalf("RunCascadeRestore: %v", err)
	}

	var active int64
	if err := db.Raw(`SELECT COUNT(*) FROM lots WHERE deleted_at IS NULL`).Scan(&active).Error; err != nil {
		t.Fatalf("count active: %v", err)
	}
	if active != 2 {
		t.Fatalf("expected 2 active lots after restore (lot 3 stayed archived), got %d", active)
	}

	var lot3Deleted *time.Time
	if err := db.Raw(`SELECT deleted_at FROM lots WHERE id = 3`).Scan(&lot3Deleted).Error; err != nil {
		t.Fatalf("lot 3: %v", err)
	}
	if lot3Deleted == nil {
		t.Fatalf("lot 3 should remain archived (different Cause)")
	}
}

func TestWouldOrphanActiveChildren_CountsPivotsAndChildren(t *testing.T) {
	t.Parallel()
	db := newCascadeTestDB(t)
	seedProjectTree(t, db)

	count, err := WouldOrphanActiveChildren(db, "projects", 1, uuid.Nil)
	if err != nil {
		t.Fatalf("WouldOrphan: %v", err)
	}
	// projects Policy direct children: 2 fields (ChildEntity) + 1
	// project_manager (CascadeTable). field_investors is a CascadeTable of
	// fields, not of projects, so it doesn't count here (helper is one level
	// deep by design). Other tables in projects Policy (workorders, labors,
	// supplies, stocks, drafts, supply_movements, admin_cost_investors,
	// project_dollar_values, crop_commercializations) are absent in this
	// simplified test schema and get skipped via hasTable.
	if count != 3 {
		t.Fatalf("expected 3 active references (2 fields + 1 manager), got %d", count)
	}
}
