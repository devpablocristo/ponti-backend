package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/devpablocristo/platform/security/go/contextkeys"
	"github.com/devpablocristo/ponti-backend/internal/shared/authz"
)

var aiTestTenantID = uuid.MustParse("00000000-0000-0000-0000-000000000601")

func newAIHandlerTestContext(scopes []string, tenantID uuid.UUID, projectID string) *gin.Context {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/chat", nil)
	if projectID != "" {
		req.Header.Set("X-PROJECT-ID", projectID)
	}
	ctx := req.Context()
	ctx = context.WithValue(ctx, ctxkeys.Actor, "ai-user")
	ctx = context.WithValue(ctx, ctxkeys.OrgID, tenantID)
	ctx = context.WithValue(ctx, ctxkeys.Role, "tenant_admin")
	ctx = context.WithValue(ctx, ctxkeys.Scopes, scopes)
	c.Request = req.WithContext(ctx)
	return c
}

func newAIHandlerTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Exec(`CREATE TABLE projects (
		id INTEGER PRIMARY KEY,
		tenant_id TEXT NOT NULL,
		deleted_at DATETIME NULL
	)`).Error; err != nil {
		t.Fatalf("create projects table: %v", err)
	}
	return db
}

// Nota: el permiso explícito `ai.use` fue removido al deprecar ponti-ai y
// migrar a Companion. Hoy basta con que el usuario haya pasado el middleware
// de auth + tenant scope (api.read/api.write). El test asociado se eliminó.

func TestExtractIDsRequiresProjectHeader(t *testing.T) {
	handler := NewHandler(nil, nil, nil, nil, nil, false)
	c := newAIHandlerTestContext([]string{authz.PermissionAIUse}, aiTestTenantID, "")

	if _, _, _, err := handler.extractIDs(c); err == nil {
		t.Fatal("expected missing project id to be rejected")
	}
}

func TestExtractIDsRejectsProjectFromAnotherTenantWhenScoped(t *testing.T) {
	db := newAIHandlerTestDB(t)
	otherTenantID := uuid.MustParse("00000000-0000-0000-0000-000000000602")
	if err := db.Exec(`INSERT INTO projects (id, tenant_id, deleted_at) VALUES (?, ?, NULL)`, 10, otherTenantID.String()).Error; err != nil {
		t.Fatalf("insert project: %v", err)
	}
	handler := NewHandler(nil, nil, nil, nil, db, true)
	c := newAIHandlerTestContext([]string{authz.PermissionAIUse}, aiTestTenantID, "10")

	if _, _, _, err := handler.extractIDs(c); err == nil {
		t.Fatal("expected cross-tenant project to be rejected")
	}
}

func TestExtractIDsAcceptsProjectForCurrentTenantWhenScoped(t *testing.T) {
	db := newAIHandlerTestDB(t)
	if err := db.Exec(`INSERT INTO projects (id, tenant_id, deleted_at) VALUES (?, ?, NULL)`, 10, aiTestTenantID.String()).Error; err != nil {
		t.Fatalf("insert project: %v", err)
	}
	handler := NewHandler(nil, nil, nil, nil, db, true)
	c := newAIHandlerTestContext([]string{authz.PermissionAIUse}, aiTestTenantID, "10")

	userID, tenantID, projectID, err := handler.extractIDs(c)
	if err != nil {
		t.Fatalf("expected ids to be accepted: %v", err)
	}
	if userID != "ai-user" || tenantID != aiTestTenantID.String() || projectID != "10" {
		t.Fatalf("unexpected ids user=%q tenant=%q project=%q", userID, tenantID, projectID)
	}
}
