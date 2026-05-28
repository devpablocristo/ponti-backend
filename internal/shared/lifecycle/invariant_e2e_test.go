package lifecycle

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/devpablocristo/platform/errors/go/domainerr"
)

// e2eInvariantsDB seeds a minimal multi-entity schema spanning the parent →
// child → pivot relations of the projects Policy. Used to assert the
// hierarchical invariant ("no active child under archived parent") holds
// end-to-end across the helpers introduced in fases 9-10.
func e2eInvariantsDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	stmts := []string{
		`CREATE TABLE customers (
			id INTEGER PRIMARY KEY, deleted_at DATETIME, deleted_by TEXT,
			updated_at DATETIME, archive_batch_id INTEGER,
			archive_origin_entity TEXT, archive_origin_id INTEGER, archive_reason TEXT
		)`,
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
	if err := db.Exec(`
		INSERT INTO customers (id) VALUES (1);
		INSERT INTO projects (id, customer_id) VALUES (10, 1);
		INSERT INTO fields (id, project_id) VALUES (100, 10), (101, 10);
		INSERT INTO lots (id, field_id) VALUES (1000, 100), (1001, 100), (1002, 101);
	`).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}
	return db
}

// TestE2E_ArchiveProject_CascadesToFieldsAndLots: after archiving a project
// using the centralized RunCascadeArchive helper, no descendant field/lot
// can stay active. The invariant query IA-2 / IA-3 must return 0 rows.
func TestE2E_ArchiveProject_CascadesToFieldsAndLots(t *testing.T) {
	t.Parallel()
	db := e2eInvariantsDB(t)

	deletedBy := "tester"
	now := time.Now().UTC().Truncate(time.Microsecond)
	cause, err := RootCause(db, uuid.Nil, "projects", 10, nil, &deletedBy)
	if err != nil {
		t.Fatalf("RootCause: %v", err)
	}
	if err := db.Exec(`UPDATE projects SET deleted_at = ?, deleted_by = ?, updated_at = ?, archive_batch_id = ?, archive_origin_entity = ?, archive_origin_id = ?
		WHERE id = 10`, now, deletedBy, now, cause.BatchID, cause.OriginEntity, cause.OriginID).Error; err != nil {
		t.Fatalf("archive project: %v", err)
	}
	if err := RunCascadeArchive(db, "projects", 10, uuid.Nil, now, &deletedBy, cause); err != nil {
		t.Fatalf("RunCascadeArchive: %v", err)
	}

	// IA-2 equivalent: no active fields under archived project.
	var nFields int64
	if err := db.Raw(`SELECT COUNT(*) FROM fields f
		JOIN projects p ON f.project_id = p.id
		WHERE f.deleted_at IS NULL AND p.deleted_at IS NOT NULL`).Scan(&nFields).Error; err != nil {
		t.Fatalf("IA-2: %v", err)
	}
	if nFields != 0 {
		t.Fatalf("IA-2 violation: %d active fields under archived project", nFields)
	}

	// IA-3 equivalent: no active lots under archived field.
	var nLots int64
	if err := db.Raw(`SELECT COUNT(*) FROM lots l
		JOIN fields f ON l.field_id = f.id
		WHERE l.deleted_at IS NULL AND f.deleted_at IS NOT NULL`).Scan(&nLots).Error; err != nil {
		t.Fatalf("IA-3: %v", err)
	}
	if nLots != 0 {
		t.Fatalf("IA-3 violation: %d active lots under archived field", nLots)
	}
}

// TestE2E_RequireActive_RejectsArchivedReference: after archiving a field,
// `RequireActive` (the helper used by every `assertXReferencesActive` in the
// system) must reject a new reference to it.
func TestE2E_RequireActive_RejectsArchivedReference(t *testing.T) {
	t.Parallel()
	db := e2eInvariantsDB(t)

	if err := db.Exec(`UPDATE fields SET deleted_at = ? WHERE id = 100`, time.Now()).Error; err != nil {
		t.Fatalf("archive field: %v", err)
	}
	err := RequireActive(db, "fields", "field", 100)
	if err == nil {
		t.Fatalf("expected RequireActive to reject archived field")
	}
	if !domainerr.IsConflict(err) {
		t.Fatalf("expected Conflict, got %T %v", err, err)
	}
	if !strings.Contains(err.Error(), "field is archived") {
		t.Fatalf("expected message %q, got %q", "field is archived", err.Error())
	}
}

// TestE2E_DataAuditQueries_HoldOnCleanDataset: with the seeded clean dataset
// (no archives), every IA-X query must return zero rows. This is the
// regression guard: if the seed or the queries diverge in the future, this
// test fires.
func TestE2E_DataAuditQueries_HoldOnCleanDataset(t *testing.T) {
	t.Parallel()
	db := e2eInvariantsDB(t)

	queries := []struct {
		name, sql string
	}{
		{"IA-1 projects under archived customer", `SELECT COUNT(*) FROM projects p
			JOIN customers c ON p.customer_id = c.id
			WHERE p.deleted_at IS NULL AND c.deleted_at IS NOT NULL`},
		{"IA-2 fields under archived project", `SELECT COUNT(*) FROM fields f
			JOIN projects p ON f.project_id = p.id
			WHERE f.deleted_at IS NULL AND p.deleted_at IS NOT NULL`},
		{"IA-3 lots under archived field", `SELECT COUNT(*) FROM lots l
			JOIN fields f ON l.field_id = f.id
			WHERE l.deleted_at IS NULL AND f.deleted_at IS NOT NULL`},
	}
	for _, q := range queries {
		var n int64
		if err := db.Raw(q.sql).Scan(&n).Error; err != nil {
			t.Fatalf("%s: %v", q.name, err)
		}
		if n != 0 {
			t.Fatalf("%s: expected 0 on clean dataset, got %d", q.name, n)
		}
	}
}

// TestE2E_CascadeRestore_RestoresFullSubtreeUnderSameCause: archive + restore
// roundtrip via the helpers must leave the dataset identical to the start,
// preserving the invariant in both directions.
func TestE2E_CascadeRestore_RestoresFullSubtreeUnderSameCause(t *testing.T) {
	t.Parallel()
	db := e2eInvariantsDB(t)

	deletedBy := "tester"
	now := time.Now().UTC().Truncate(time.Microsecond)
	cause, err := RootCause(db, uuid.Nil, "projects", 10, nil, &deletedBy)
	if err != nil {
		t.Fatalf("RootCause: %v", err)
	}
	if err := db.Exec(`UPDATE projects SET deleted_at = ?, archive_batch_id = ?, archive_origin_entity = ?, archive_origin_id = ?
		WHERE id = 10`, now, cause.BatchID, cause.OriginEntity, cause.OriginID).Error; err != nil {
		t.Fatalf("archive root: %v", err)
	}
	if err := RunCascadeArchive(db, "projects", 10, uuid.Nil, now, &deletedBy, cause); err != nil {
		t.Fatalf("cascade archive: %v", err)
	}

	// Now restore: clear project's deleted_at and run the cascade restore.
	if err := db.Exec(`UPDATE projects SET deleted_at = NULL WHERE id = 10`).Error; err != nil {
		t.Fatalf("restore root: %v", err)
	}
	if err := RunCascadeRestore(db, "projects", 10, uuid.Nil, now, cause); err != nil {
		t.Fatalf("cascade restore: %v", err)
	}

	checks := []struct {
		name, sql string
		want      int64
	}{
		{"fields", `SELECT COUNT(*) FROM fields WHERE deleted_at IS NULL`, 2},
		{"lots", `SELECT COUNT(*) FROM lots WHERE deleted_at IS NULL`, 3},
	}
	for _, c := range checks {
		var n int64
		if err := db.Raw(c.sql).Scan(&n).Error; err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if n != c.want {
			t.Fatalf("%s: expected %d active, got %d", c.name, c.want, n)
		}
	}
}
