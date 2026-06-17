package actors

import (
	"context"
	"database/sql"
	"testing"

	ctxkeys "github.com/devpablocristo/platform/security/go/contextkeys"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	domain "github.com/devpablocristo/ponti-backend/internal/actors/usecases/domain"
)

type actorsTestEngine struct {
	db *gorm.DB
}

func (e actorsTestEngine) Client() *gorm.DB { return e.db }

func newActorsRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	stmts := []string{
		`CREATE TABLE actors (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT, party_type TEXT, display_name TEXT, raw_name TEXT,
			status TEXT DEFAULT 'active',
			created_at DATETIME, updated_at DATETIME, deleted_at DATETIME,
			created_by TEXT, updated_by TEXT);`,
		`CREATE TABLE actor_roles (
			actor_id INTEGER NOT NULL, role TEXT NOT NULL, created_at DATETIME,
			PRIMARY KEY (actor_id, role));`,
		`CREATE TABLE actor_keys (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			actor_id INTEGER NOT NULL, tenant_id TEXT,
			key_type TEXT NOT NULL, key_value TEXT NOT NULL,
			active BOOLEAN NOT NULL DEFAULT 1, source TEXT NOT NULL DEFAULT 'direct',
			created_at DATETIME);`,
		`CREATE UNIQUE INDEX uq_actor_keys_active ON actor_keys (tenant_id, key_type, key_value) WHERE active;`,
		`CREATE TABLE customers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			tenant_id TEXT DEFAULT 'default-tenant',
			actor_id INTEGER,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			deleted_by TEXT);`,
		`CREATE UNIQUE INDEX uq_customers_tenant_name ON customers (tenant_id, name) WHERE deleted_at IS NULL;`,
		`CREATE TABLE projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			customer_id INTEGER NOT NULL,
			deleted_at DATETIME);`,
	}
	for _, s := range stmts {
		if err := db.Exec(s).Error; err != nil {
			t.Fatalf("schema: %v", err)
		}
	}
	return db
}

func actorsTenantContext(id uuid.UUID) context.Context {
	return context.WithValue(context.Background(), ctxkeys.OrgID, id)
}

func TestResolveCustomerCreatesLegacyCustomer(t *testing.T) {
	db := newActorsRepositoryTestDB(t)
	repo := NewRepository(actorsTestEngine{db: db})
	tenantID := uuid.New()
	taxID := "30111111111"

	res, err := repo.Resolve(actorsTenantContext(tenantID), domain.ResolveInput{
		Name:           "Cliente Nuevo SA",
		TaxID:          &taxID,
		Role:           "customer",
		AllowCreate:    true,
		RejectExisting: true,
	})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	var row struct {
		ID      int64
		Name    string
		ActorID int64
		Tenant  string
	}
	if err := db.Raw("SELECT id, name, actor_id, tenant_id AS tenant FROM customers WHERE actor_id = ?", res.Actor.ID).Scan(&row).Error; err != nil {
		t.Fatalf("load customer: %v", err)
	}
	if row.ID == 0 {
		t.Fatal("expected legacy customer to be created")
	}
	if row.Name != "Cliente Nuevo SA" {
		t.Fatalf("customer name = %q", row.Name)
	}
	if row.ActorID != res.Actor.ID {
		t.Fatalf("customer actor_id = %d, want %d", row.ActorID, res.Actor.ID)
	}
	if row.Tenant != tenantID.String() {
		t.Fatalf("customer tenant = %q, want %q", row.Tenant, tenantID.String())
	}
}

func TestSetRolesMaterializesAndArchivesCustomerRole(t *testing.T) {
	db := newActorsRepositoryTestDB(t)
	repo := NewRepository(actorsTestEngine{db: db})
	tenantID := uuid.New()
	ctx := actorsTenantContext(tenantID)

	res, err := repo.Resolve(ctx, domain.ResolveInput{
		Name:           "Proveedor Luego Cliente",
		Role:           "provider",
		AllowCreate:    true,
		RejectExisting: true,
	})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	if err := repo.SetRoles(ctx, res.Actor.ID, []string{"provider", "customer"}); err != nil {
		t.Fatalf("SetRoles add customer: %v", err)
	}
	var customerID int64
	if err := db.Raw("SELECT id FROM customers WHERE actor_id = ? AND deleted_at IS NULL", res.Actor.ID).Scan(&customerID).Error; err != nil {
		t.Fatalf("load customer id: %v", err)
	}
	if customerID == 0 {
		t.Fatal("expected active customer after adding customer role")
	}

	if err := db.Exec("INSERT INTO projects (customer_id) VALUES (?)", customerID).Error; err != nil {
		t.Fatalf("insert project: %v", err)
	}
	if err := repo.SetRoles(ctx, res.Actor.ID, []string{"provider"}); err == nil {
		t.Fatal("expected conflict when removing customer role with active projects")
	}

	if err := db.Exec("UPDATE projects SET deleted_at = CURRENT_TIMESTAMP WHERE customer_id = ?", customerID).Error; err != nil {
		t.Fatalf("archive project: %v", err)
	}
	if err := repo.SetRoles(ctx, res.Actor.ID, []string{"provider"}); err != nil {
		t.Fatalf("SetRoles remove customer: %v", err)
	}

	var archived sql.NullString
	if err := db.Raw("SELECT deleted_at FROM customers WHERE id = ?", customerID).Scan(&archived).Error; err != nil {
		t.Fatalf("load deleted_at: %v", err)
	}
	if !archived.Valid {
		t.Fatal("expected legacy customer to be archived after removing customer role")
	}
}

func TestUpdateAndRestoreSyncLegacyCustomer(t *testing.T) {
	db := newActorsRepositoryTestDB(t)
	repo := NewRepository(actorsTestEngine{db: db})
	tenantID := uuid.New()
	ctx := actorsTenantContext(tenantID)

	res, err := repo.Resolve(ctx, domain.ResolveInput{
		Name:           "Cliente Original",
		Role:           "customer",
		AllowCreate:    true,
		RejectExisting: true,
	})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	if err := repo.Update(ctx, &domain.Actor{ID: res.Actor.ID, DisplayName: "Cliente Renombrado", PartyType: "org"}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	var name string
	if err := db.Raw("SELECT name FROM customers WHERE actor_id = ?", res.Actor.ID).Scan(&name).Error; err != nil {
		t.Fatalf("load name: %v", err)
	}
	if name != "Cliente Renombrado" {
		t.Fatalf("customer name = %q", name)
	}

	if err := repo.Archive(ctx, res.Actor.ID); err != nil {
		t.Fatalf("Archive: %v", err)
	}
	var archived sql.NullString
	if err := db.Raw("SELECT deleted_at FROM customers WHERE actor_id = ?", res.Actor.ID).Scan(&archived).Error; err != nil {
		t.Fatalf("load archived customer: %v", err)
	}
	if !archived.Valid {
		t.Fatal("expected archived legacy customer")
	}

	if err := repo.Restore(ctx, res.Actor.ID); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if err := db.Raw("SELECT deleted_at FROM customers WHERE actor_id = ?", res.Actor.ID).Scan(&archived).Error; err != nil {
		t.Fatalf("load restored customer: %v", err)
	}
	if archived.Valid {
		t.Fatal("expected restored legacy customer")
	}
}
