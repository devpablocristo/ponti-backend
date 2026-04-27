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
			status TEXT
		);`,
		`INSERT INTO projects (id, deleted_at) VALUES (30, NULL);`,
		`INSERT INTO v4_report.workorder_list (id, number, project_id, field_id, date, sequence_day, is_digital, status) VALUES
			(10, '2000', 30, 40, '2026-04-23T00:00:00Z', 1, false, 'published'),
			(11, '1862', 30, 40, '2026-03-29T00:00:00Z', 1, false, 'published'),
			(12, '1706', 30, 40, '2026-04-23T00:00:00Z', 2, false, 'published'),
			(-20, 'D-20', 30, 40, '2026-04-24T00:00:00Z', 0, true, 'draft');`,
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
	if rows[0].ID != -20 || rows[0].Number != "D-20" {
		t.Fatalf("expected digital draft first by latest date, got id=%d number=%q", rows[0].ID, rows[0].Number)
	}
	if rows[1].ID != 12 || rows[1].Number != "1706" {
		t.Fatalf("expected latest date work order first with highest sequence, got id=%d number=%q", rows[1].ID, rows[1].Number)
	}
	if rows[2].ID != 10 || rows[2].Number != "2000" {
		t.Fatalf("expected same-date lower sequence second, got id=%d number=%q", rows[2].ID, rows[2].Number)
	}
	if rows[3].ID != 11 || rows[3].Number != "1862" {
		t.Fatalf("expected oldest date last, got id=%d number=%q", rows[3].ID, rows[3].Number)
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
