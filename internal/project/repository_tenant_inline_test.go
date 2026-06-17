package project

import (
	"context"
	"testing"

	ctxkeys "github.com/devpablocristo/platform/security/go/contextkeys"
	campmod "github.com/devpablocristo/ponti-backend/internal/campaign/repository/models"
	cusmod "github.com/devpablocristo/ponti-backend/internal/customer/repository/models"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestEnsureCustomerAndCampaignStampRequestTenant(t *testing.T) {
	t.Setenv("TENANT_ENFORCEMENT", "true")

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	stmts := []string{
		`CREATE TABLE customers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			tenant_id TEXT DEFAULT 'default-tenant',
			actor_id INTEGER,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT);`,
		`CREATE TABLE campaigns (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			tenant_id TEXT DEFAULT 'default-tenant',
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			created_by TEXT,
			updated_by TEXT,
			deleted_by TEXT);`,
	}
	for _, s := range stmts {
		if err := db.Exec(s).Error; err != nil {
			t.Fatalf("schema: %v", err)
		}
	}

	tenantID := uuid.New()
	ctx := context.WithValue(context.Background(), ctxkeys.OrgID, tenantID)
	tx := db.WithContext(ctx).Begin()
	defer tx.Rollback()

	customerID, err := ensureCustomer(tx, &cusmod.Customer{Name: "Cliente Tenant Real"})
	if err != nil {
		t.Fatalf("ensureCustomer: %v", err)
	}
	campaignID, err := ensureCampaign(tx, &campmod.Campaign{Name: "Campaña Tenant Real"})
	if err != nil {
		t.Fatalf("ensureCampaign: %v", err)
	}

	var customerTenant string
	if err := tx.Raw("SELECT tenant_id FROM customers WHERE id = ?", customerID).Scan(&customerTenant).Error; err != nil {
		t.Fatalf("load customer tenant: %v", err)
	}
	if customerTenant != tenantID.String() {
		t.Fatalf("customer tenant = %q, want %q", customerTenant, tenantID.String())
	}

	var campaignTenant string
	if err := tx.Raw("SELECT tenant_id FROM campaigns WHERE id = ?", campaignID).Scan(&campaignTenant).Error; err != nil {
		t.Fatalf("load campaign tenant: %v", err)
	}
	if campaignTenant != tenantID.String() {
		t.Fatalf("campaign tenant = %q, want %q", campaignTenant, tenantID.String())
	}
}
