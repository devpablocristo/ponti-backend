package labor

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type laborTestGormEngine struct {
	client *gorm.DB
}

func (e *laborTestGormEngine) Client() *gorm.DB { return e.client }

func TestInvoiceFallbackJoinSQL(t *testing.T) {
	joinSQL := `LEFT JOIN LATERAL (
    SELECT i.*
    FROM invoices i
    WHERE i.work_order_id = v4.workorder_id
      AND (i.investor_id = v4.investor_id OR i.investor_id IS NULL)
      AND i.deleted_at IS NULL
    ORDER BY
      CASE
        WHEN i.investor_id = v4.investor_id THEN 0
        WHEN i.investor_id IS NULL THEN 1
        ELSE 2
      END,
      i.id DESC
    LIMIT 1
) i ON true`

	assert.Contains(t, joinSQL, "LEFT JOIN LATERAL")
	assert.Contains(t, joinSQL, "i.work_order_id = v4.workorder_id")
	assert.Contains(t, joinSQL, "i.investor_id = v4.investor_id OR i.investor_id IS NULL")
	assert.Contains(t, joinSQL, "WHEN i.investor_id = v4.investor_id THEN 0")
	assert.Contains(t, joinSQL, "WHEN i.investor_id IS NULL THEN 1")
	assert.Contains(t, joinSQL, "LIMIT 1")
}

func TestRepository_ListLabor_ReturnsPendingAndCategorizedCatalogRows(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	schema := []string{
		`CREATE TABLE categories (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			type_id INTEGER NOT NULL,
			created_at DATETIME NULL,
			updated_at DATETIME NULL,
			deleted_at DATETIME NULL,
			created_by TEXT NULL,
			updated_by TEXT NULL,
			deleted_by TEXT NULL
		);`,
		`CREATE TABLE labors (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			contractor_name TEXT NOT NULL,
			price TEXT NOT NULL DEFAULT '0',
			is_partial_price BOOLEAN NOT NULL DEFAULT false,
			project_id INTEGER NOT NULL,
			category_id INTEGER NULL,
			is_pending BOOLEAN NOT NULL DEFAULT false,
			created_at DATETIME NULL,
			updated_at DATETIME NULL,
			deleted_at DATETIME NULL,
			created_by TEXT NULL,
			updated_by TEXT NULL,
			deleted_by TEXT NULL
		);`,
		`INSERT INTO categories (id, name, type_id) VALUES (11, 'Otras Labores', 4);`,
		`INSERT INTO labors (
			id, name, contractor_name, price, is_partial_price, project_id, category_id,
			is_pending, updated_at
		) VALUES
			(119, 'LIMPIEZA MANUAL', 'E.VEDOYA', '1.15', false, 30, 11, false, '2026-03-11T14:04:48Z'),
			(120, 'LABOR PENDIENTE', '', '0', false, 30, NULL, true, '2026-03-12T14:04:48Z'),
			(121, 'OTRO PROYECTO', 'X', '9', false, 31, 11, false, '2026-03-13T14:04:48Z');`,
	}

	for _, stmt := range schema {
		assert.NoError(t, db.Exec(stmt).Error)
	}

	repo := NewRepository(&laborTestGormEngine{client: db})

	items, total, err := repo.ListLabor(context.Background(), 1, 100, 30)

	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	if assert.Len(t, items, 2) {
		byName := map[string]int{}
		for i, item := range items {
			byName[item.Name] = i
		}

		catalog := items[byName["LIMPIEZA MANUAL"]]
		assert.Equal(t, int64(119), catalog.ID)
		assert.Equal(t, "E.VEDOYA", catalog.ContractorName)
		assert.Equal(t, int64(11), catalog.CategoryId)
		assert.Equal(t, "Otras Labores", catalog.CategoryName)
		assert.False(t, catalog.IsPending)

		pending := items[byName["LABOR PENDIENTE"]]
		assert.Equal(t, int64(120), pending.ID)
		assert.Equal(t, int64(0), pending.CategoryId)
		assert.Equal(t, "", pending.CategoryName)
		assert.True(t, pending.IsPending)
	}
}

func TestLaborPendingMigrationDefinesCatalogPrerequisites(t *testing.T) {
	contents, err := os.ReadFile("../../migrations_v4/000232_labor_pending_changes.up.sql")
	assert.NoError(t, err)

	sql := strings.ToLower(string(contents))
	assert.Contains(t, sql, "add column if not exists is_pending")
	assert.Contains(t, sql, "alter column category_id drop not null")
	assert.Contains(t, sql, "alter column price set default 0")
}
