package registry

import (
	"context"
	"testing"

	ctxkeys "github.com/devpablocristo/platform/security/go/contextkeys"
	"github.com/devpablocristo/platform/errors/go/domainerr"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type stubEngine struct{ db *gorm.DB }

func (s stubEngine) Client() *gorm.DB { return s.db }

func newAliasDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	stmts := []string{
		`CREATE TABLE actors (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT,
			party_type TEXT,
			display_name TEXT,
			raw_name TEXT,
			deleted_at DATETIME);`,
		`CREATE TABLE actor_keys (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			actor_id INTEGER,
			tenant_id TEXT,
			key_type TEXT,
			key_value TEXT,
			active BOOLEAN,
			source TEXT);`,
	}
	for _, s := range stmts {
		if err := db.Exec(s).Error; err != nil {
			t.Fatalf("schema: %v", err)
		}
	}
	return db
}

func countAliases(t *testing.T, db *gorm.DB, actorID int64) int64 {
	t.Helper()
	var n int64
	if err := db.Raw(`SELECT count(*) FROM actor_keys WHERE actor_id = ? AND key_type = 'ALIAS' AND active`, actorID).Scan(&n).Error; err != nil {
		t.Fatalf("count aliases: %v", err)
	}
	return n
}

// SetAliases debe rechazar (404) reescribir alias de un actor de OTRO tenant, incluso con
// TENANT_ENFORCEMENT apagado — el scoping de ownership es incondicional. Antes del fix esto
// era un IDOR de escritura cross-tenant.
func TestSetAliases_CrossTenantDenied(t *testing.T) {
	db := newAliasDB(t)
	tenantA := uuid.New()
	tenantB := uuid.New()
	if err := db.Exec(`INSERT INTO actors (id, tenant_id, party_type, display_name, raw_name) VALUES (1, ?, 'company', 'A', 'A')`, tenantA.String()).Error; err != nil {
		t.Fatalf("seed actor: %v", err)
	}

	repo := NewRepository(stubEngine{db: db})
	ctxB := context.WithValue(context.Background(), ctxkeys.OrgID, tenantB)

	err := repo.SetAliases(ctxB, 1, []string{"hacker-alias"})
	if !domainerr.IsNotFound(err) {
		t.Fatalf("want NotFound for cross-tenant actor, got %v", err)
	}
	if n := countAliases(t, db, 1); n != 0 {
		t.Fatalf("cross-tenant write leaked: %d alias rows", n)
	}
}

// El dueño (mismo tenant) sí puede setear alias.
func TestSetAliases_OwnerSucceeds(t *testing.T) {
	db := newAliasDB(t)
	tenantA := uuid.New()
	if err := db.Exec(`INSERT INTO actors (id, tenant_id, party_type, display_name, raw_name) VALUES (1, ?, 'company', 'A', 'A')`, tenantA.String()).Error; err != nil {
		t.Fatalf("seed actor: %v", err)
	}

	repo := NewRepository(stubEngine{db: db})
	ctxA := context.WithValue(context.Background(), ctxkeys.OrgID, tenantA)

	if err := repo.SetAliases(ctxA, 1, []string{"alias-1"}); err != nil {
		t.Fatalf("owner SetAliases: %v", err)
	}
	if n := countAliases(t, db, 1); n != 1 {
		t.Fatalf("want 1 alias for owner, got %d", n)
	}
}
