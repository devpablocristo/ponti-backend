package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/devpablocristo/platform/security/go/contextkeys"

	"github.com/devpablocristo/ponti-backend/internal/admin/idp"
)

func newMeContextTestHandler(t *testing.T, db *gorm.DB) *Handler {
	t.Helper()
	uc := NewUseCases(NewRepository(db), &idp.NoopAdmin{})
	return &Handler{uc: uc}
}

func TestGetMeContextReturnsCurrentTenantMembershipsAndPermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newMeContextTestDB(t)

	userID := uuid.New()
	tenantID := uuid.New()
	roleID := uuid.New()
	permissionID := uuid.New()

	execMeContextSQL(t, db, `
		INSERT INTO users (id, username, idp_sub, idp_email, email)
		VALUES (?, ?, ?, ?, ?)
	`, userID.String(), "user", "idp-user-1", "user@example.com", "user@example.com")
	execMeContextSQL(t, db, `
		INSERT INTO auth_tenants (id, name)
		VALUES (?, ?)
	`, tenantID.String(), "default")
	execMeContextSQL(t, db, `
		INSERT INTO auth_roles (id, name)
		VALUES (?, ?)
	`, roleID.String(), "tenant_admin")
	execMeContextSQL(t, db, `
		INSERT INTO auth_memberships (id, tenant_id, user_id, role_id, status)
		VALUES (?, ?, ?, ?, ?)
	`, uuid.New().String(), tenantID.String(), userID.String(), roleID.String(), "active")
	execMeContextSQL(t, db, `
		INSERT INTO auth_permissions (id, name)
		VALUES (?, ?)
	`, permissionID.String(), "customers.read")
	execMeContextSQL(t, db, `
		INSERT INTO auth_role_permissions (role_id, permission_id)
		VALUES (?, ?)
	`, roleID.String(), permissionID.String())

	h := newMeContextTestHandler(t, db)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/context", nil)
	ctx := context.WithValue(req.Context(), ctxkeys.Actor, "idp-user-1")
	ctx = context.WithValue(ctx, ctxkeys.OrgID, tenantID)
	c.Request = req.WithContext(ctx)

	h.GetMeContext(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		User struct {
			ID       string `json:"id"`
			IDPSub   string `json:"idp_sub"`
			IDPEmail string `json:"idp_email"`
			Email    string `json:"email"`
		} `json:"user"`
		CurrentTenantID string `json:"current_tenant_id"`
		Tenants         []struct {
			ID          string   `json:"id"`
			Name        string   `json:"name"`
			Role        string   `json:"role"`
			Permissions []string `json:"permissions"`
			IsCurrent   bool     `json:"is_current"`
		} `json:"tenants"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.User.ID != userID.String() || payload.User.IDPSub != "idp-user-1" || payload.User.Email != "user@example.com" {
		t.Fatalf("unexpected user payload: %+v", payload.User)
	}
	if payload.CurrentTenantID != tenantID.String() {
		t.Fatalf("expected current tenant %s, got %s", tenantID, payload.CurrentTenantID)
	}
	if len(payload.Tenants) != 1 {
		t.Fatalf("expected one tenant, got %d", len(payload.Tenants))
	}
	tenant := payload.Tenants[0]
	if tenant.ID != tenantID.String() || tenant.Name != "default" || tenant.Role != "tenant_admin" || !tenant.IsCurrent {
		t.Fatalf("unexpected tenant payload: %+v", tenant)
	}
	if len(tenant.Permissions) != 1 || tenant.Permissions[0] != "customers.read" {
		t.Fatalf("unexpected permissions: %+v", tenant.Permissions)
	}
}

func TestGetMeContextRequiresActor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newMeContextTestHandler(t, newMeContextTestDB(t))
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/me/context", nil)

	h.GetMeContext(c)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func newMeContextTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	statements := []string{
		`CREATE TABLE users (
			id TEXT PRIMARY KEY,
			username TEXT NOT NULL DEFAULT '',
			idp_sub TEXT NOT NULL,
			idp_email TEXT NOT NULL,
			email TEXT NOT NULL
		)`,
		`CREATE TABLE auth_tenants (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL
		)`,
		`CREATE TABLE auth_roles (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL
		)`,
		`CREATE TABLE auth_memberships (
			id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			role_id TEXT NOT NULL,
			status TEXT NOT NULL
		)`,
		`CREATE TABLE auth_permissions (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL
		)`,
		`CREATE TABLE auth_role_permissions (
			role_id TEXT NOT NULL,
			permission_id TEXT NOT NULL
		)`,
	}
	for _, stmt := range statements {
		execMeContextSQL(t, db, stmt)
	}
	return db
}

func execMeContextSQL(t *testing.T, db *gorm.DB, query string, args ...any) {
	t.Helper()
	if err := db.Exec(query, args...).Error; err != nil {
		t.Fatalf("exec sql: %v", err)
	}
}
