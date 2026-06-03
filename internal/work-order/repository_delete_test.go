package workorder

import (
	"context"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type deleteTestGormEngine struct {
	client *gorm.DB
}

func (e *deleteTestGormEngine) Client() *gorm.DB {
	return e.client
}

func newDeleteWorkOrderTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	statements := []string{
		`PRAGMA foreign_keys = ON;`,
		`CREATE TABLE workorders (
			id INTEGER PRIMARY KEY,
			deleted_at DATETIME
		);`,
		`CREATE TABLE workorder_items (
			id INTEGER PRIMARY KEY,
			workorder_id INTEGER,
			deleted_at DATETIME
		);`,
		`CREATE TABLE workorder_investor_splits (
			id INTEGER PRIMARY KEY,
			workorder_id INTEGER,
			deleted_at DATETIME
		);`,
		`CREATE TABLE work_order_drafts (
			id INTEGER PRIMARY KEY,
			published_work_order_id INTEGER,
			FOREIGN KEY (published_work_order_id) REFERENCES workorders(id) ON DELETE RESTRICT
		);`,
		`CREATE TABLE work_order_draft_items (
			id INTEGER PRIMARY KEY,
			draft_id INTEGER,
			FOREIGN KEY (draft_id) REFERENCES work_order_drafts(id) ON DELETE RESTRICT
		);`,
		`CREATE TABLE work_order_draft_investor_splits (
			id INTEGER PRIMARY KEY,
			draft_id INTEGER,
			FOREIGN KEY (draft_id) REFERENCES work_order_drafts(id) ON DELETE RESTRICT
		);`,
		`INSERT INTO workorders (id, deleted_at) VALUES (598, NULL);`,
		`INSERT INTO workorder_items (id, workorder_id, deleted_at) VALUES (1, 598, NULL);`,
		`INSERT INTO workorder_investor_splits (id, workorder_id, deleted_at) VALUES (1, 598, NULL);`,
		`INSERT INTO work_order_drafts (id, published_work_order_id) VALUES (10, 598);`,
		`INSERT INTO work_order_draft_items (id, draft_id) VALUES (20, 10);`,
		`INSERT INTO work_order_draft_investor_splits (id, draft_id) VALUES (30, 10);`,
	}

	for _, stmt := range statements {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("exec %q: %v", stmt, err)
		}
	}

	return db
}

func TestRepository_DeleteWorkOrderByID_RemovesPublishedDigitalDraftReference(t *testing.T) {
	db := newDeleteWorkOrderTestDB(t)
	repo := NewRepository(&deleteTestGormEngine{client: db})

	if err := repo.DeleteWorkOrderByID(context.Background(), 598); err != nil {
		t.Fatalf("delete work order: %v", err)
	}

	for _, table := range []string{
		"workorders",
		"workorder_items",
		"workorder_investor_splits",
		"work_order_drafts",
		"work_order_draft_items",
		"work_order_draft_investor_splits",
	} {
		var count int64
		if err := db.Table(table).Count(&count).Error; err != nil {
			t.Fatalf("count %s: %v", table, err)
		}
		if count != 0 {
			t.Fatalf("expected %s to be empty after hard delete, got %d rows", table, count)
		}
	}
}