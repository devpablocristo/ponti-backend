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
			field_id INTEGER
		);`,
		`INSERT INTO projects (id, deleted_at) VALUES (30, NULL);`,
		`INSERT INTO v4_report.workorder_list (id, number, project_id, field_id) VALUES
			(10, '2000', 30, 40),
			(11, '1862', 30, 40);`,
	}

	for _, stmt := range statements {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("exec %q: %v", stmt, err)
		}
	}

	return db
}

func TestRepository_ListWorkOrders_OrdersByLatestCreatedFirst(t *testing.T) {
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

	if pageInfo.Total != 2 {
		t.Fatalf("expected total 2, got %d", pageInfo.Total)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[0].ID != 11 || rows[0].Number != "1862" {
		t.Fatalf("expected latest created work order first, got id=%d number=%q", rows[0].ID, rows[0].Number)
	}
	if rows[1].ID != 10 || rows[1].Number != "2000" {
		t.Fatalf("expected older work order second, got id=%d number=%q", rows[1].ID, rows[1].Number)
	}
}
