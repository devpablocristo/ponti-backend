package dollar

import (
	"context"
	"testing"

	contextkeys "github.com/devpablocristo/platform/security/go/contextkeys"
	domain "github.com/devpablocristo/ponti-backend/internal/dollar/usecases/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type dollarTenantGormEngine struct {
	client *gorm.DB
}

func (e dollarTenantGormEngine) Client() *gorm.DB {
	return e.client
}

func dollarTenantContext(tenantID uuid.UUID) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextkeys.Actor, "tenant-user@example.com")
	ctx = context.WithValue(ctx, contextkeys.OrgID, tenantID)
	ctx = context.WithValue(ctx, contextkeys.Role, "tenant_admin")
	ctx = context.WithValue(ctx, contextkeys.Scopes, []string{"projects.read", "projects.write"})
	return ctx
}

func setupDollarTenantDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Exec(`
		CREATE TABLE projects (
			id INTEGER PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			deleted_at DATETIME
		);
		CREATE TABLE project_dollar_values (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT,
			project_id INTEGER NOT NULL,
			year INTEGER NOT NULL,
			month TEXT NOT NULL,
			start_value NUMERIC NOT NULL,
			end_value NUMERIC NOT NULL,
			average_value NUMERIC NOT NULL,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		);
	`).Error; err != nil {
		t.Fatalf("create schema: %v", err)
	}

	return db
}

func TestDollarRepositoryProjectTenantIsolation(t *testing.T) {
	db := setupDollarTenantDB(t)
	repo := NewRepository(dollarTenantGormEngine{client: db})

	tenantA := uuid.New()
	tenantB := uuid.New()
	if err := db.Exec(`
		INSERT INTO projects (id, tenant_id) VALUES (10, ?), (20, ?);
		INSERT INTO project_dollar_values
			(id, tenant_id, project_id, year, month, start_value, end_value, average_value)
		VALUES
			(1, ?, 10, 2026, 'enero', 1, 2, 1.5),
			(2, ?, 20, 2026, 'enero', 3, 4, 3.5);
	`, tenantA.String(), tenantB.String(), tenantA.String(), tenantB.String()).Error; err != nil {
		t.Fatalf("seed dollar values: %v", err)
	}

	ctxA := dollarTenantContext(tenantA)
	if values, err := repo.ListByProject(ctxA, 10); err != nil || len(values) != 1 || values[0].ID != 1 {
		t.Fatalf("expected own project values, values=%#v err=%v", values, err)
	}
	if _, err := repo.ListByProject(ctxA, 20); err == nil {
		t.Fatalf("expected cross-tenant list to fail")
	}
	if _, err := repo.GetByComposite(ctxA, 20, 2026, "enero"); err == nil {
		t.Fatalf("expected cross-tenant composite lookup to fail")
	}
	if _, err := repo.Create(ctxA, &domain.DollarAverage{
		ProjectID:  20,
		Year:       2026,
		Month:      "febrero",
		StartValue: decimal.NewFromInt(1),
		EndValue:   decimal.NewFromInt(1),
		AvgValue:   decimal.NewFromInt(1),
	}); err == nil {
		t.Fatalf("expected cross-tenant create to fail")
	}
	if err := repo.Update(ctxA, &domain.DollarAverage{
		ID:         2,
		ProjectID:  20,
		Year:       2026,
		Month:      "enero",
		StartValue: decimal.NewFromInt(10),
		EndValue:   decimal.NewFromInt(10),
		AvgValue:   decimal.NewFromInt(10),
	}); err == nil {
		t.Fatalf("expected cross-tenant update to fail")
	}

	var avg string
	if err := db.Raw(`SELECT average_value FROM project_dollar_values WHERE id = 2`).Scan(&avg).Error; err != nil {
		t.Fatalf("read cross-tenant row: %v", err)
	}
	if avg != "3.5" && avg != "3.50" {
		t.Fatalf("cross-tenant update changed average_value to %q", avg)
	}
}

func TestDollarRepositoryRequiresTenantInStrictMode(t *testing.T) {
	t.Setenv("TENANT_STRICT_MODE", "true")

	db := setupDollarTenantDB(t)
	repo := NewRepository(dollarTenantGormEngine{client: db})

	if _, err := repo.ListByProject(context.Background(), 10); err == nil {
		t.Fatalf("expected strict list without tenant to fail")
	}
}
