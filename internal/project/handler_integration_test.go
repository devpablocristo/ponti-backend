package project

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	domainField "github.com/alphacodinggroup/ponti-backend/internal/field/usecases/domain"
	domain "github.com/alphacodinggroup/ponti-backend/internal/project/usecases/domain"
)

type fakeUseCases struct {
	updateCalled bool
	updatedProj  *domain.Project
	updateErr    error
}

func (f *fakeUseCases) CreateProject(context.Context, *domain.Project) (int64, error) {
	return 0, nil
}
func (f *fakeUseCases) GetProjects(context.Context, string, int64, int64, int, int) ([]domain.Project, decimal.Decimal, int64, error) {
	return nil, decimal.Zero, 0, nil
}
func (f *fakeUseCases) ListArchivedProjects(context.Context, int, int) ([]domain.Project, decimal.Decimal, int64, error) {
	return nil, decimal.Zero, 0, nil
}
func (f *fakeUseCases) ListProjects(context.Context, int, int) ([]domain.ListedProject, int64, error) {
	return nil, 0, nil
}
func (f *fakeUseCases) ListProjectsByCustomerID(context.Context, int64, int, int) ([]domain.ListedProject, int64, error) {
	return nil, 0, nil
}
func (f *fakeUseCases) ListProjectsByName(context.Context, string, int, int) ([]domain.ListedProject, int64, error) {
	return nil, 0, nil
}
func (f *fakeUseCases) GetFieldsByProjectID(context.Context, int64) ([]domainField.Field, error) {
	return nil, nil
}
func (f *fakeUseCases) GetProject(context.Context, int64) (*domain.Project, error) {
	return nil, nil
}
func (f *fakeUseCases) UpdateProject(_ context.Context, p *domain.Project) error {
	f.updateCalled = true
	f.updatedProj = p
	return f.updateErr
}
func (f *fakeUseCases) DeleteProject(context.Context, int64) error     { return nil }
func (f *fakeUseCases) RestoreProject(context.Context, int64) error    { return nil }
func (f *fakeUseCases) HardDeleteProject(context.Context, int64) error { return nil }

type fakeGinEngine struct{ r *gin.Engine }

func (f *fakeGinEngine) GetRouter() *gin.Engine           { return f.r }
func (f *fakeGinEngine) RunServer(context.Context) error  { return nil }

type fakeConfig struct{}

func (fakeConfig) APIVersion() string { return "v1" }
func (fakeConfig) APIBaseURL() string { return "/api/v1" }

type fakeMiddlewares struct{}

func (fakeMiddlewares) GetGlobal() []gin.HandlerFunc     { return nil }
func (fakeMiddlewares) GetValidation() []gin.HandlerFunc { return nil }
func (fakeMiddlewares) GetProtected() []gin.HandlerFunc  { return nil }

func setupProjectRouter(ucs *fakeUseCases) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	h := NewHandler(
		ucs,
		&fakeGinEngine{r: r},
		fakeConfig{},
		fakeMiddlewares{},
	)
	h.Routes()
	return r
}

func TestUpdateProject_AllowsFrontendPayloadWithEmptyLeaseTypeDecimals(t *testing.T) {
	ucs := &fakeUseCases{}
	router := setupProjectRouter(ucs)

	payload := `{
		"name":"DEPOSITO",
		"updated_at":"2026-02-14T12:00:00Z",
		"customer":{"id":25,"name":"SOALEN SRL 25%"},
		"campaign":{"id":3,"name":"2025-2026"},
		"admin_cost":100,
		"planned_cost":200,
		"managers":[{"id":1,"name":"RESP"}],
		"investors":[{"id":1,"name":"INV","percentage":100}],
		"admin_cost_investors":[{"id":1,"name":"INV","percentage":100}],
		"fields":[
			{
				"id":10,
				"name":"DEPOSITO",
				"lease_type_id":2,
				"lease_type_percent":"",
				"lease_type_value":"",
				"investors":[{"id":1,"name":"INV","percentage":100}],
				"lots":[
					{
						"id":1,
						"name":"1",
						"hectares":10,
						"previous_crop_id":1,
						"previous_crop_name":"Soja",
						"current_crop_id":1,
						"current_crop_name":"Soja",
						"season":"Verano"
					}
				]
			}
		]
	}`

	req := httptest.NewRequest(http.MethodPut, "/api/v1/projects/25", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d. body=%s", rr.Code, rr.Body.String())
	}
	if !ucs.updateCalled {
		t.Fatal("expected UpdateProject to be called")
	}
	if ucs.updatedProj == nil {
		t.Fatal("expected updated project to be captured")
	}
	if ucs.updatedProj.ID != 25 {
		t.Fatalf("expected project id=25, got %d", ucs.updatedProj.ID)
	}
	if len(ucs.updatedProj.Fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(ucs.updatedProj.Fields))
	}
	if ucs.updatedProj.Fields[0].LeaseTypePercent != nil {
		t.Fatalf("expected lease_type_percent nil, got %+v", ucs.updatedProj.Fields[0].LeaseTypePercent)
	}
	if ucs.updatedProj.Fields[0].LeaseTypeValue != nil {
		t.Fatalf("expected lease_type_value nil, got %+v", ucs.updatedProj.Fields[0].LeaseTypeValue)
	}
}

func TestUpdateProject_RejectsInvalidLeaseTypeDecimal(t *testing.T) {
	ucs := &fakeUseCases{}
	router := setupProjectRouter(ucs)

	payload := `{
		"name":"DEPOSITO",
		"updated_at":"2026-02-14T12:00:00Z",
		"customer":{"id":25,"name":"SOALEN SRL 25%"},
		"campaign":{"id":3,"name":"2025-2026"},
		"admin_cost":100,
		"planned_cost":200,
		"managers":[{"id":1,"name":"RESP"}],
		"investors":[{"id":1,"name":"INV","percentage":100}],
		"admin_cost_investors":[{"id":1,"name":"INV","percentage":100}],
		"fields":[
			{
				"id":10,
				"name":"DEPOSITO",
				"lease_type_id":2,
				"lease_type_percent":"abc",
				"lease_type_value":null,
				"investors":[{"id":1,"name":"INV","percentage":100}],
				"lots":[
					{
						"id":1,
						"name":"1",
						"hectares":10,
						"previous_crop_id":1,
						"previous_crop_name":"Soja",
						"current_crop_id":1,
						"current_crop_name":"Soja",
						"season":"Verano"
					}
				]
			}
		]
	}`

	req := httptest.NewRequest(http.MethodPut, "/api/v1/projects/25", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d. body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "payload inválido") {
		t.Fatalf("expected explicit payload inválido message, got body=%s", rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "lease_type_percent") {
		t.Fatalf("expected error details about lease_type_percent, got body=%s", rr.Body.String())
	}
	if ucs.updateCalled {
		t.Fatal("did not expect UpdateProject to be called on invalid payload")
	}
}

func TestUpdateProject_RejectsMissingFieldInvestorsWithExplicitMessage(t *testing.T) {
	ucs := &fakeUseCases{}
	router := setupProjectRouter(ucs)

	payload := `{
		"name":"DEPOSITO",
		"updated_at":"2026-02-14T12:00:00Z",
		"customer":{"id":25,"name":"SOALEN SRL 25%"},
		"campaign":{"id":3,"name":"2025-2026"},
		"admin_cost":100,
		"planned_cost":200,
		"managers":[{"id":1,"name":"RESP"}],
		"investors":[{"id":1,"name":"INV","percentage":100}],
		"admin_cost_investors":[{"id":1,"name":"INV","percentage":100}],
		"fields":[
			{
				"id":10,
				"name":"DEPOSITO",
				"lease_type_id":2,
				"lease_type_percent":"10",
				"lease_type_value":"0",
				"lots":[
					{
						"id":1,
						"name":"1",
						"hectares":10,
						"previous_crop_id":1,
						"previous_crop_name":"Soja",
						"current_crop_id":1,
						"current_crop_name":"Soja",
						"season":"Verano"
					}
				]
			}
		]
	}`

	req := httptest.NewRequest(http.MethodPut, "/api/v1/projects/25", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d. body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "fields[0].investors es requerido") {
		t.Fatalf("expected explicit missing field investors message, got body=%s", rr.Body.String())
	}
	if ucs.updateCalled {
		t.Fatal("did not expect UpdateProject to be called on invalid payload")
	}
}
