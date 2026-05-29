package workorder

import (
	"context"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	types "github.com/devpablocristo/ponti-backend/internal/shared/types"
	domain "github.com/devpablocristo/ponti-backend/internal/work-order/usecases/domain"
)

type listTestGormEngine struct {
	client *gorm.DB
}

func (e *listTestGormEngine) Client() *gorm.DB {
	return e.client
}

func newListWorkOrdersTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	statements := []string{
		`CREATE TABLE projects (
			id INTEGER PRIMARY KEY,
			deleted_at DATETIME
		);`,
		`ATTACH DATABASE ':memory:' AS v4_report;`,
		`CREATE TABLE v4_report.workorder_list (
			id INTEGER,
			number TEXT,
			project_id INTEGER,
			field_id INTEGER,
			date DATETIME,
			sequence_day INTEGER,
			is_digital BOOLEAN,
			status TEXT,
			supply_name TEXT
		);`,
		`CREATE TABLE supplies (
			id INTEGER PRIMARY KEY,
			name TEXT,
			deleted_at DATETIME
		);`,
		`CREATE TABLE workorder_items (
			id INTEGER PRIMARY KEY,
			workorder_id INTEGER,
			supply_id INTEGER,
			deleted_at DATETIME
		);`,
		`CREATE TABLE work_order_draft_items (
			id INTEGER PRIMARY KEY,
			draft_id INTEGER,
			supply_id INTEGER,
			deleted_at DATETIME
		);`,
		`INSERT INTO projects (id, deleted_at) VALUES (30, NULL);`,
		`INSERT INTO supplies (id, name, deleted_at) VALUES
			(100, '2-4D', NULL),
			(101, 'Glifosato', NULL);`,
		`INSERT INTO workorder_items (id, workorder_id, supply_id, deleted_at) VALUES
			(1, 10, 100, NULL),
			(2, 12, 101, NULL);`,
		`INSERT INTO work_order_draft_items (id, draft_id, supply_id, deleted_at) VALUES
			(1, 20, 100, NULL),
			(2, 20, 101, '2026-04-25T00:00:00Z');`,
		`INSERT INTO v4_report.workorder_list (id, number, project_id, field_id, date, sequence_day, is_digital, status, supply_name) VALUES
			(10, '2000', 30, 40, '2026-04-23T00:00:00Z', 1, false, 'published', '2-4D'),
			(11, '1862', 30, 40, '2026-03-29T00:00:00Z', 1, false, 'published', NULL),
			(12, '1706', 30, 40, '2026-04-23T00:00:00Z', 2, false, 'published', 'Glifosato'),
			(-20, 'D-2000', 30, 40, '2026-04-24T00:00:00Z', 0, true, 'draft', '2-4D');`,
	}

	for _, stmt := range statements {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("exec %q: %v", stmt, err)
		}
	}

	return db
}

func TestRepository_ListWorkOrders_OrdersByLatestDateFirst(t *testing.T) {
	db := newListWorkOrdersTestDB(t)
	repo := NewRepository(&listTestGormEngine{client: db})

	projectID := int64(30)
	rows, pageInfo, err := repo.ListWorkOrders(
		context.Background(),
		domain.WorkOrderFilter{ProjectID: &projectID},
		types.Input{Page: 1, PageSize: 10},
	)
	if err != nil {
		t.Fatalf("list work orders: %v", err)
	}

	if pageInfo.Total != 4 {
		t.Fatalf("expected total 4, got %d", pageInfo.Total)
	}
	if len(rows) != 4 {
		t.Fatalf("expected 4 rows, got %d", len(rows))
	}
	if rows[0].ID != -20 || rows[0].Number != "D-2000" {
		t.Fatalf("expected newest digital draft first, got id=%d number=%q", rows[0].ID, rows[0].Number)
	}
	if rows[1].ID != 12 || rows[1].Number != "1706" {
		t.Fatalf("expected latest date work order first with highest sequence, got id=%d number=%q", rows[0].ID, rows[0].Number)
	}
	if rows[2].ID != 10 || rows[2].Number != "2000" {
		t.Fatalf("expected same-date lower sequence second, got id=%d number=%q", rows[1].ID, rows[1].Number)
	}
	if rows[3].ID != 11 || rows[3].Number != "1862" {
		t.Fatalf("expected oldest date last, got id=%d number=%q", rows[2].ID, rows[2].Number)
	}
}

func TestRepository_ListWorkOrders_FiltersDigitalDrafts(t *testing.T) {
	db := newListWorkOrdersTestDB(t)
	repo := NewRepository(&listTestGormEngine{client: db})

	projectID := int64(30)
	isDigital := true
	status := "draft"
	rows, pageInfo, err := repo.ListWorkOrders(
		context.Background(),
		domain.WorkOrderFilter{
			ProjectID: &projectID,
			IsDigital: &isDigital,
			Status:    &status,
		},
		types.Input{Page: 1, PageSize: 10},
	)
	if err != nil {
		t.Fatalf("list work orders: %v", err)
	}

	if pageInfo.Total != 1 {
		t.Fatalf("expected total 1, got %d", pageInfo.Total)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0].ID != -20 || !rows[0].IsDigital || rows[0].Status != "draft" {
		t.Fatalf("expected digital draft row, got id=%d is_digital=%v status=%q", rows[0].ID, rows[0].IsDigital, rows[0].Status)
	}
}

func TestRepository_ListWorkOrders_FiltersSupplyForPublishedAndDigitalDrafts(t *testing.T) {
	db := newListWorkOrdersTestDB(t)
	repo := NewRepository(&listTestGormEngine{client: db})

	projectID := int64(30)
	supplyID := int64(100)
	rows, pageInfo, err := repo.ListWorkOrders(
		context.Background(),
		domain.WorkOrderFilter{
			ProjectID: &projectID,
			SupplyID:  &supplyID,
		},
		types.Input{Page: 1, PageSize: 10},
	)
	if err != nil {
		t.Fatalf("list work orders: %v", err)
	}

	if pageInfo.Total != 2 {
		t.Fatalf("expected total 2, got %d", pageInfo.Total)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[0].ID != -20 || rows[1].ID != 10 {
		t.Fatalf("expected digital draft and published order for supply filter, got ids=%d,%d", rows[0].ID, rows[1].ID)
	}
}

func TestRepository_ListWorkOrderFilterRows_ReturnsAllRowsWithoutPagination(t *testing.T) {
	db := newListWorkOrdersTestDB(t)
	repo := NewRepository(&listTestGormEngine{client: db})

	projectID := int64(30)
	rows, err := repo.ListWorkOrderFilterRows(
		context.Background(),
		domain.WorkOrderFilter{ProjectID: &projectID},
	)
	if err != nil {
		t.Fatalf("list work order filter rows: %v", err)
	}

	if len(rows) != 4 {
		t.Fatalf("expected all 4 rows for filter source, got %d", len(rows))
	}
	if rows[0].ID != -20 || rows[3].ID != 11 {
		t.Fatalf("expected filter rows to preserve work order ordering, got first=%d last=%d", rows[0].ID, rows[3].ID)
	}
}
