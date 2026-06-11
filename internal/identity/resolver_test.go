package identity

import (
	"context"
	"testing"

	ctxkeys "github.com/devpablocristo/platform/security/go/contextkeys"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newIdentityTestDB(t *testing.T) *gorm.DB {
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
	}
	for _, s := range stmts {
		if err := db.Exec(s).Error; err != nil {
			t.Fatalf("schema: %v", err)
		}
	}
	return db
}

func ctxTenant(id uuid.UUID) context.Context {
	return context.WithValue(context.Background(), ctxkeys.OrgID, id)
}

func sp(s string) *string { return &s }

func roleCount(t *testing.T, db *gorm.DB, actorID int64) int {
	t.Helper()
	var n int64
	if err := db.Raw(`SELECT count(*) FROM actor_roles WHERE actor_id = ?`, actorID).Scan(&n).Error; err != nil {
		t.Fatalf("roleCount: %v", err)
	}
	return int(n)
}

func TestResolveOrCreateIdentity(t *testing.T) {
	db := newIdentityTestDB(t)
	tenantA := uuid.New()
	ctx := ctxTenant(tenantA)
	cuit := "30-11111111-1"

	// 1) alta nueva: customer Acme S.A. con CUIT
	r1, err := ResolveOrCreateIdentity(ctx, db, RoleCustomer, ResolveInput{RawName: "Acme S.A.", TaxID: sp(cuit)})
	if err != nil || r1.Reused || r1.ActorID == 0 {
		t.Fatalf("alta customer: %+v err=%v", r1, err)
	}

	// 2) mismo CUIT, nombre distinto, como provider → REUSA por TAX_ID (cross-rol)
	r2, err := ResolveOrCreateIdentity(ctx, db, RoleProvider, ResolveInput{RawName: "ACME SOCIEDAD ANONIMA SUCURSAL", TaxID: sp("30111111111")})
	if err != nil || !r2.Reused || r2.ActorID != r1.ActorID || r2.MatchedKey != "TAX_ID" {
		t.Fatalf("reuse por CUIT: %+v err=%v (esperaba reuse de %d por TAX_ID)", r2, err, r1.ActorID)
	}

	// 3) sin CUIT, nombre legal equivalente → REUSA por LEGAL_NAME (acme|SA)
	r3, err := ResolveOrCreateIdentity(ctx, db, RoleInvestor, ResolveInput{RawName: "Acme Sociedad Anónima"})
	if err != nil || !r3.Reused || r3.ActorID != r1.ActorID || r3.MatchedKey != "LEGAL_NAME" {
		t.Fatalf("reuse por nombre legal: %+v err=%v", r3, err)
	}

	// el ente acumuló 3 roles (customer+provider+investor) en UNA identidad
	if got := roleCount(t, db, r1.ActorID); got != 3 {
		t.Fatalf("roles acumulados = %d, want 3", got)
	}

	// 4) Acme SRL → forma jurídica distinta → NUEVA identidad
	r4, err := ResolveOrCreateIdentity(ctx, db, RoleCustomer, ResolveInput{RawName: "Acme SRL"})
	if err != nil || r4.Reused || r4.ActorID == r1.ActorID {
		t.Fatalf("Acme SRL debería ser nueva identidad: %+v err=%v", r4, err)
	}

	// 5) aislamiento por tenant: mismo nombre en tenant B → otra identidad
	r5, err := ResolveOrCreateIdentity(ctxTenant(uuid.New()), db, RoleCustomer, ResolveInput{RawName: "Acme SRL"})
	if err != nil || r5.ActorID == r4.ActorID {
		t.Fatalf("tenant B debería crear identidad propia: %+v err=%v", r5, err)
	}

	// 6) el índice único es la garantía: insertar a mano otra key activa igual → viola
	dup := db.Exec(
		`INSERT INTO actor_keys (actor_id, tenant_id, key_type, key_value, active, source) VALUES (?, ?, 'TAX_ID', '30111111111', true, 'direct')`,
		r1.ActorID, tenantA,
	).Error
	if dup == nil {
		t.Fatal("insert duplicado de clave activa debería violar el índice único")
	}
}
