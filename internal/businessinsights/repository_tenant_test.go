package businessinsights

import (
	"context"
	"testing"
	"time"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/devpablocristo/ponti-backend/internal/businessinsights/repository/models"
)

func setupBusinessInsightsTenantDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.CandidateModel{}, &models.ReadModel{}); err != nil {
		t.Fatalf("migrate business insights models: %v", err)
	}
	return db
}

func TestBusinessInsightsRepositoryTenantIsolation(t *testing.T) {
	db := setupBusinessInsightsTenantDB(t)
	repo := NewRepository(db)

	tenantA := uuid.New()
	tenantB := uuid.New()
	now := time.Date(2026, 5, 12, 12, 0, 0, 0, time.UTC)

	candidateA, _, err := repo.Upsert(context.Background(), CandidateUpsert{
		TenantID:    tenantA.String(),
		Kind:        "insight",
		EventType:   "ponti.stock.negative",
		EntityType:  "supply",
		EntityID:    "supply-a",
		Fingerprint: "tenant-a:fingerprint",
		Severity:    "warning",
		Title:       "A",
		Body:        "Tenant A",
		Actor:       "tester",
		Now:         now,
	})
	if err != nil {
		t.Fatalf("upsert tenant A candidate: %v", err)
	}
	candidateB, _, err := repo.Upsert(context.Background(), CandidateUpsert{
		TenantID:    tenantB.String(),
		Kind:        "insight",
		EventType:   "ponti.stock.negative",
		EntityType:  "supply",
		EntityID:    "supply-b",
		Fingerprint: "tenant-b:fingerprint",
		Severity:    "warning",
		Title:       "B",
		Body:        "Tenant B",
		Actor:       "tester",
		Now:         now,
	})
	if err != nil {
		t.Fatalf("upsert tenant B candidate: %v", err)
	}

	list, err := repo.ListByTenantForUser(context.Background(), tenantA.String(), "user-a", ListOptions{IncludeResolved: true})
	if err != nil {
		t.Fatalf("list tenant A candidates: %v", err)
	}
	if len(list) != 1 || list[0].ID != candidateA.ID {
		t.Fatalf("expected tenant A candidate only, got %#v", list)
	}

	if err := repo.MarkRead(context.Background(), tenantA.String(), candidateB.ID, "user-a", now); err == nil {
		t.Fatalf("expected MarkRead to reject cross-tenant candidate")
	}
	var reads int64
	if err := db.Model(&models.ReadModel{}).Count(&reads).Error; err != nil {
		t.Fatalf("count reads: %v", err)
	}
	if reads != 0 {
		t.Fatalf("cross-tenant MarkRead created %d read rows", reads)
	}

	if err := repo.ResolveByID(context.Background(), tenantA.String(), candidateB.ID, "user-a", now); !domainerr.IsKind(err, domainerr.KindNotFound) {
		t.Fatalf("expected ResolveByID cross-tenant not found, got %v", err)
	}
	if err := repo.ReopenByID(context.Background(), tenantA.String(), candidateB.ID, "user-a", now); !domainerr.IsKind(err, domainerr.KindNotFound) {
		t.Fatalf("expected ReopenByID cross-tenant not found, got %v", err)
	}
	if err := repo.MarkNotified(context.Background(), tenantA.String(), candidateB.ID, now); err == nil {
		t.Fatalf("expected MarkNotified to reject cross-tenant candidate")
	}

	if err := repo.ResolveByID(context.Background(), tenantA.String(), candidateA.ID, "user-a", now); err != nil {
		t.Fatalf("resolve tenant A candidate: %v", err)
	}
	var status string
	if err := db.Raw(`SELECT status FROM business_insight_candidates WHERE id = ?`, candidateA.ID).Scan(&status).Error; err != nil {
		t.Fatalf("read tenant A status: %v", err)
	}
	if status != "resolved" {
		t.Fatalf("expected tenant A candidate resolved, got %q", status)
	}

	if err := db.Raw(`SELECT status FROM business_insight_candidates WHERE id = ?`, candidateB.ID).Scan(&status).Error; err != nil {
		t.Fatalf("read tenant B status: %v", err)
	}
	if status != "new" {
		t.Fatalf("cross-tenant actions changed tenant B status to %q", status)
	}
}
