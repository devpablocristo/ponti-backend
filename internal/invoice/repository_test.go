package invoice

import (
	"context"
	"testing"
	"time"

	"github.com/devpablocristo/ponti-backend/internal/invoice/repository/models"
	domain "github.com/devpablocristo/ponti-backend/internal/invoice/usecases/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type testGormEngine struct {
	client *gorm.DB
}

func (e *testGormEngine) Client() *gorm.DB { return e.client }

func newInvoiceTestRepo(t *testing.T) *Repository {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	stmts := []string{
		`CREATE TABLE invoices (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			work_order_id INTEGER NOT NULL,
			investor_id INTEGER NULL,
			number TEXT NOT NULL,
			company TEXT NOT NULL,
			date DATETIME NOT NULL,
			status TEXT NOT NULL,
			created_at DATETIME NULL,
			updated_at DATETIME NULL,
			deleted_at DATETIME NULL,
			created_by TEXT NULL,
			updated_by TEXT NULL,
			deleted_by TEXT NULL
		);`,
		`CREATE TABLE workorders (
			id INTEGER PRIMARY KEY,
			investor_id INTEGER NOT NULL,
			deleted_at DATETIME NULL
		);`,
		`CREATE TABLE workorder_investor_splits (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workorder_id INTEGER NOT NULL,
			investor_id INTEGER NOT NULL,
			percentage NUMERIC NOT NULL,
			payment_status TEXT NOT NULL,
			deleted_at DATETIME NULL
		);`,
	}

	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("exec schema: %v", err)
		}
	}

	return NewRepository(&testGormEngine{client: db})
}

func TestRepository_GetByWorkOrderAndInvestor_PrefersInvestorSpecificInvoice(t *testing.T) {
	repo := newInvoiceTestRepo(t)
	ctx := context.Background()
	db := repo.db.Client()
	now := time.Now()

	assert.NoError(t, db.Create(&models.Invoice{
		WorkOrderID: 10,
		InvestorID:  0,
		Number:      "LEG-1",
		Company:     "Legacy SA",
		Date:        now,
		Status:      "Pendiente",
	}).Error)
	assert.NoError(t, db.Create(&models.Invoice{
		WorkOrderID: 10,
		InvestorID:  7,
		Number:      "INV-7",
		Company:     "Laguna Blanca",
		Date:        now,
		Status:      "Pagada",
	}).Error)

	item, err := repo.GetByWorkOrderAndInvestor(ctx, 10, 7)
	assert.NoError(t, err)
	assert.Equal(t, int64(7), item.InvestorID)
	assert.Equal(t, "INV-7", item.Number)
}

func TestRepository_GetByWorkOrderAndInvestor_FallsBackToLegacyInvoice(t *testing.T) {
	repo := newInvoiceTestRepo(t)
	ctx := context.Background()
	db := repo.db.Client()
	now := time.Now()

	assert.NoError(t, db.Exec(
		`INSERT INTO invoices (work_order_id, investor_id, number, company, date, status) VALUES (?, NULL, ?, ?, ?, ?)`,
		22, "LEG-22", "Legacy SA", now, "Pendiente",
	).Error)

	item, err := repo.GetByWorkOrderAndInvestor(ctx, 22, 9)
	assert.NoError(t, err)
	if assert.NotNil(t, item) {
		assert.Equal(t, int64(0), item.InvestorID)
		assert.Equal(t, "LEG-22", item.Number)
	}
}

func TestRepository_InvestorBelongsToWorkOrder_UsesSplitWhenExists(t *testing.T) {
	repo := newInvoiceTestRepo(t)
	ctx := context.Background()
	db := repo.db.Client()

	assert.NoError(t, db.Exec(`INSERT INTO workorders (id, investor_id) VALUES (?, ?)`, 30, 100).Error)
	assert.NoError(t, db.Exec(
		`INSERT INTO workorder_investor_splits (workorder_id, investor_id, percentage, payment_status) VALUES (?, ?, ?, ?)`,
		30, 7, decimal.NewFromInt(50), "Pendiente",
	).Error)

	ok, err := repo.InvestorBelongsToWorkOrder(ctx, 30, 7)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = repo.InvestorBelongsToWorkOrder(ctx, 30, 100)
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestRepository_InvestorBelongsToWorkOrder_UsesWorkOrderInvestorWhenNoSplits(t *testing.T) {
	repo := newInvoiceTestRepo(t)
	ctx := context.Background()
	db := repo.db.Client()

	assert.NoError(t, db.Exec(`INSERT INTO workorders (id, investor_id) VALUES (?, ?)`, 40, 11).Error)

	ok, err := repo.InvestorBelongsToWorkOrder(ctx, 40, 11)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = repo.InvestorBelongsToWorkOrder(ctx, 40, 12)
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestRepository_Update_UpgradesLegacyInvoiceToInvestorSpecific(t *testing.T) {
	repo := newInvoiceTestRepo(t)
	ctx := context.Background()
	db := repo.db.Client()
	now := time.Now()

	assert.NoError(t, db.Exec(
		`INSERT INTO invoices (work_order_id, investor_id, number, company, date, status) VALUES (?, NULL, ?, ?, ?, ?)`,
		50, "LEG-50", "Legacy SA", now, "Pendiente",
	).Error)

	err := repo.Update(ctx, &domain.Invoice{
		WorkOrderID: 50,
		InvestorID:  8,
		Number:      "INV-50",
		Company:     "Nueva SA",
		Date:        now.Add(24 * time.Hour),
		Status:      "Pagada",
	})
	assert.NoError(t, err)

	var row models.Invoice
	assert.NoError(t, db.Where("work_order_id = ?", 50).First(&row).Error)
	assert.Equal(t, int64(8), row.InvestorID)
	assert.Equal(t, "INV-50", row.Number)
	assert.Equal(t, "Nueva SA", row.Company)
	assert.Equal(t, "Pagada", row.Status)
}

func TestRepository_Delete_RemovesLegacyInvoiceForInvestorFallback(t *testing.T) {
	repo := newInvoiceTestRepo(t)
	ctx := context.Background()
	db := repo.db.Client()
	now := time.Now()

	assert.NoError(t, db.Exec(
		`INSERT INTO invoices (work_order_id, investor_id, number, company, date, status) VALUES (?, NULL, ?, ?, ?, ?)`,
		60, "LEG-60", "Legacy SA", now, "Pendiente",
	).Error)

	err := repo.Delete(ctx, 60, 9)
	assert.NoError(t, err)

	var count int64
	assert.NoError(t, db.Model(&models.Invoice{}).Where("work_order_id = ?", 60).Count(&count).Error)
	assert.Equal(t, int64(0), count)
}
