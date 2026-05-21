package actor

import (
	"context"
	"testing"
	"time"

	contextkeys "github.com/devpablocristo/platform/security/go/contextkeys"
	domain "github.com/devpablocristo/ponti-backend/internal/actor/usecases/domain"
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type actorTenantGormEngine struct {
	client *gorm.DB
}

func (e actorTenantGormEngine) Client() *gorm.DB {
	return e.client
}

func actorTenantContext(tenantID uuid.UUID) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextkeys.Actor, "tenant-user@example.com")
	ctx = context.WithValue(ctx, contextkeys.OrgID, tenantID)
	ctx = context.WithValue(ctx, contextkeys.Role, "tenant_admin")
	ctx = context.WithValue(ctx, contextkeys.Scopes, []string{"actors.read", "actors.write", "actors.archive", "actors.merge"})
	return ctx
}

func setupActorTenantDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.Exec(`
		CREATE TABLE actors (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			actor_kind TEXT NOT NULL,
			display_name TEXT NOT NULL,
			normalized_name TEXT NOT NULL,
			primary_email TEXT,
			primary_phone TEXT,
			notes TEXT,
			archived_at DATETIME,
			merged_into_actor_id INTEGER,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT
		);
		CREATE TABLE actor_roles (
			actor_id INTEGER NOT NULL,
			role TEXT NOT NULL,
			created_at DATETIME,
			archived_at DATETIME,
			PRIMARY KEY (actor_id, role)
		);
		CREATE TABLE actor_aliases (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			actor_id INTEGER NOT NULL,
			alias TEXT NOT NULL,
			normalized_alias TEXT NOT NULL,
			source TEXT,
			created_at DATETIME,
			archived_at DATETIME
		);
		CREATE TABLE actor_identifiers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT NOT NULL,
			actor_id INTEGER NOT NULL,
			country TEXT NOT NULL,
			identifier_type TEXT NOT NULL,
			identifier_value TEXT NOT NULL,
			normalized_identifier_value TEXT NOT NULL,
			is_primary BOOLEAN NOT NULL DEFAULT false,
			created_at DATETIME
		);
		CREATE TABLE actor_person_profiles (
			actor_id INTEGER PRIMARY KEY,
			first_name TEXT,
			last_name TEXT,
			birth_date DATETIME,
			document_type TEXT,
			document_number TEXT,
			normalized_document_number TEXT
		);
		CREATE TABLE actor_organization_profiles (
			actor_id INTEGER PRIMARY KEY,
			legal_name TEXT,
			normalized_legal_name TEXT,
			trade_name TEXT,
			normalized_trade_name TEXT,
			legal_entity_type TEXT,
			tax_condition TEXT,
			fiscal_address TEXT
		);
	`).Error; err != nil {
		t.Fatalf("create schema: %v", err)
	}

	return db
}

func TestActorRepositoryTenantIsolation(t *testing.T) {
	db := setupActorTenantDB(t)
	repo := NewRepository(actorTenantGormEngine{client: db})

	tenantA := uuid.New()
	tenantB := uuid.New()
	now := time.Now().UTC()

	if err := db.Exec(`
		INSERT INTO actors (
			id, tenant_id, actor_kind, display_name, normalized_name,
			archived_at, created_at, updated_at, deleted_at
		) VALUES
			(1, ?, 'organization', 'Actor A', 'actor a', NULL, ?, ?, NULL),
			(2, ?, 'organization', 'Actor B', 'actor b', NULL, ?, ?, NULL),
			(3, ?, 'organization', 'Actor B archived', 'actor b archived', ?, ?, ?, NULL);
		INSERT INTO actor_roles (actor_id, role, created_at, archived_at) VALUES
			(1, 'cliente', ?, NULL),
			(2, 'cliente', ?, NULL),
			(3, 'cliente', ?, NULL);
		INSERT INTO actor_aliases (id, tenant_id, actor_id, alias, normalized_alias, created_at, archived_at) VALUES
			(1, ?, 1, 'Alias A', 'alias a', ?, NULL),
			(2, ?, 2, 'Alias B', 'alias b', ?, NULL);
		INSERT INTO actor_identifiers (
			id, tenant_id, actor_id, country, identifier_type, identifier_value,
			normalized_identifier_value, is_primary, created_at
		) VALUES
			(1, ?, 1, 'AR', 'tax_id', '20-11111111-1', '20111111111', true, ?),
			(2, ?, 2, 'AR', 'tax_id', '20-22222222-2', '20222222222', true, ?);
	`, tenantA.String(), now, now,
		tenantB.String(), now, now,
		tenantB.String(), now, now, now,
		now, now, now,
		tenantA.String(), now,
		tenantB.String(), now,
		tenantA.String(), now,
		tenantB.String(), now,
	).Error; err != nil {
		t.Fatalf("seed actors: %v", err)
	}

	ctxA := actorTenantContext(tenantA)

	actors, total, err := repo.ListActors(ctxA, domain.ListFilters{Status: "all"}, 1, 50)
	if err != nil {
		t.Fatalf("list actors: %v", err)
	}
	if total != 1 || len(actors) != 1 || actors[0].ID != 1 {
		t.Fatalf("expected only tenant A actors, total=%d actors=%#v", total, actors)
	}

	if _, err := repo.GetActor(ctxA, 2); err == nil {
		t.Fatalf("expected get cross-tenant actor to fail")
	}

	if err := repo.UpdateActor(ctxA, &domain.Actor{
		ID:          2,
		ActorKind:   domain.KindOrganization,
		DisplayName: "cross tenant update",
		Roles:       []string{"cliente"},
		Base:        shareddomain.Base{UpdatedAt: now},
	}); err == nil {
		t.Fatalf("expected update cross-tenant actor to fail")
	}
	var displayName string
	if err := db.Raw(`SELECT display_name FROM actors WHERE id = 2`).Scan(&displayName).Error; err != nil {
		t.Fatalf("read actor 2: %v", err)
	}
	if displayName != "Actor B" {
		t.Fatalf("cross-tenant update changed actor 2 to %q", displayName)
	}

	if err := repo.ArchiveActor(ctxA, 2); err == nil {
		t.Fatalf("expected archive cross-tenant actor to fail")
	}
	var archivedCount int64
	if err := db.Raw(`SELECT COUNT(*) FROM actors WHERE id = 2 AND archived_at IS NOT NULL`).Scan(&archivedCount).Error; err != nil {
		t.Fatalf("check archive side effect: %v", err)
	}
	if archivedCount != 0 {
		t.Fatalf("cross-tenant archive modified actor 2")
	}

	if err := repo.RestoreActor(ctxA, 3); err == nil {
		t.Fatalf("expected restore cross-tenant actor to fail")
	}
	if err := db.Raw(`SELECT COUNT(*) FROM actors WHERE id = 3 AND archived_at IS NOT NULL`).Scan(&archivedCount).Error; err != nil {
		t.Fatalf("check restore side effect: %v", err)
	}
	if archivedCount != 1 {
		t.Fatalf("cross-tenant restore modified actor 3")
	}

	if err := repo.AddRole(ctxA, 2, "inversor"); err == nil {
		t.Fatalf("expected add role on cross-tenant actor to fail")
	}
	if _, err := repo.AddAlias(ctxA, 2, domain.ActorAlias{Alias: "Cross"}); err == nil {
		t.Fatalf("expected add alias on cross-tenant actor to fail")
	}

	if _, err := repo.MergeActors(ctxA, domain.MergeRequest{
		TargetActorID:  1,
		SourceActorIDs: []int64{2},
		Reason:         "cross tenant",
		Confirm:        false,
		MergedBy:       "tenant-user@example.com",
	}); err == nil {
		t.Fatalf("expected cross-tenant merge simulation to fail")
	}

	if err := repo.HardDeleteActor(ctxA, 2); err == nil {
		t.Fatalf("expected hard-delete cross-tenant actor to fail")
	}
	var exists int64
	if err := db.Raw(`SELECT COUNT(*) FROM actors WHERE id = 2`).Scan(&exists).Error; err != nil {
		t.Fatalf("check hard delete side effect: %v", err)
	}
	if exists != 1 {
		t.Fatalf("cross-tenant hard delete removed actor 2")
	}

	createdID, err := repo.CreateActor(ctxA, &domain.Actor{
		ActorKind:   domain.KindNaturalPerson,
		DisplayName: "Nuevo Actor",
		Roles:       []string{"responsable"},
		Aliases:     []domain.ActorAlias{{Alias: "N Actor"}},
		Identifiers: []domain.ActorIdentifier{{
			Country:         "AR",
			IdentifierType:  "national_id",
			IdentifierValue: "12345678",
			IsPrimary:       true,
		}},
		Base: shareddomain.Base{CreatedAt: now, UpdatedAt: now},
	})
	if err != nil {
		t.Fatalf("create tenant actor: %v", err)
	}
	var createdTenant string
	if err := db.Raw(`SELECT tenant_id FROM actors WHERE id = ?`, createdID).Scan(&createdTenant).Error; err != nil {
		t.Fatalf("read created actor tenant: %v", err)
	}
	if createdTenant != tenantA.String() {
		t.Fatalf("created actor tenant_id = %q, want %q", createdTenant, tenantA.String())
	}
}
