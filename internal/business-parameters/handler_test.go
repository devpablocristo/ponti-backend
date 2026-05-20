package bparams

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	ctxkeys "github.com/devpablocristo/core/security/go/contextkeys"
	"github.com/gin-gonic/gin"

	domain "github.com/devpablocristo/ponti-backend/internal/business-parameters/usecases/domain"
)

type businessParametersHandlerUseCasesStub struct {
	getKeyCalls      []string
	getCategoryCalls []string
	getAllCalls      int
	createCalls      []domain.BusinessParameter
	updateCalls      []domain.BusinessParameter
	deleteCalls      []int64
}

func (s *businessParametersHandlerUseCasesStub) GetParameter(_ context.Context, key string) (*domain.BusinessParameter, error) {
	s.getKeyCalls = append(s.getKeyCalls, key)
	return &domain.BusinessParameter{ID: 1, Key: key, Value: "1", Type: "decimal", Category: "units"}, nil
}

func (s *businessParametersHandlerUseCasesStub) GetParametersByCategory(_ context.Context, category string) ([]domain.BusinessParameter, error) {
	s.getCategoryCalls = append(s.getCategoryCalls, category)
	return []domain.BusinessParameter{{ID: 1, Key: "kg", Value: "1", Type: "decimal", Category: category}}, nil
}

func (s *businessParametersHandlerUseCasesStub) GetAllParameters(context.Context) ([]domain.BusinessParameter, error) {
	s.getAllCalls++
	return []domain.BusinessParameter{{ID: 1, Key: "kg", Value: "1", Type: "decimal", Category: "units"}}, nil
}

func (s *businessParametersHandlerUseCasesStub) GetArchivedParameters(context.Context) ([]domain.BusinessParameter, error) {
	s.getAllCalls++
	return []domain.BusinessParameter{{ID: 1, Key: "kg", Value: "1", Type: "decimal", Category: "units"}}, nil
}

func (s *businessParametersHandlerUseCasesStub) CreateParameter(_ context.Context, param *domain.BusinessParameter) (int64, error) {
	s.createCalls = append(s.createCalls, *param)
	return 99, nil
}

func (s *businessParametersHandlerUseCasesStub) UpdateParameter(_ context.Context, param *domain.BusinessParameter) error {
	s.updateCalls = append(s.updateCalls, *param)
	return nil
}

func (s *businessParametersHandlerUseCasesStub) ArchiveParameter(_ context.Context, id int64) error {
	s.deleteCalls = append(s.deleteCalls, id)
	return nil
}

func (s *businessParametersHandlerUseCasesStub) RestoreParameter(_ context.Context, id int64) error {
	s.deleteCalls = append(s.deleteCalls, id)
	return nil
}

func (s *businessParametersHandlerUseCasesStub) HardDeleteParameter(_ context.Context, id int64) error {
	s.deleteCalls = append(s.deleteCalls, id)
	return nil
}


func newBusinessParametersHandlerContext(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), ctxkeys.Actor, "tester@example.com"))
	ctx.Request = req
	return ctx, rec
}

func TestHandler_BusinessParametersRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := &businessParametersHandlerUseCasesStub{}
	h := &Handler{ucs: stub}

	allCtx, _ := newBusinessParametersHandlerContext(http.MethodGet, "/api/v1/business-parameters", "")
	h.GetAllParameters(allCtx)
	if allCtx.Writer.Status() != http.StatusOK || stub.getAllCalls != 1 {
		t.Fatalf("expected get all, status %d calls %d", allCtx.Writer.Status(), stub.getAllCalls)
	}

	categoryCtx, _ := newBusinessParametersHandlerContext(http.MethodGet, "/api/v1/business-parameters/category/units", "")
	categoryCtx.Params = gin.Params{{Key: "category", Value: "units"}}
	h.GetParametersByCategory(categoryCtx)
	if categoryCtx.Writer.Status() != http.StatusOK || len(stub.getCategoryCalls) != 1 || stub.getCategoryCalls[0] != "units" {
		t.Fatalf("expected category units, status %d calls %#v", categoryCtx.Writer.Status(), stub.getCategoryCalls)
	}

	getCtx, _ := newBusinessParametersHandlerContext(http.MethodGet, "/api/v1/business-parameters/kg", "")
	getCtx.Params = gin.Params{{Key: "parameter_key", Value: "kg"}}
	h.GetParameter(getCtx)
	if getCtx.Writer.Status() != http.StatusOK || len(stub.getKeyCalls) != 1 || stub.getKeyCalls[0] != "kg" {
		t.Fatalf("expected key kg, status %d calls %#v", getCtx.Writer.Status(), stub.getKeyCalls)
	}

	body := `{"key":"kg","value":"1","type":"decimal","category":"units","description":"Kilogramo"}`
	createCtx, _ := newBusinessParametersHandlerContext(http.MethodPost, "/api/v1/business-parameters", body)
	h.CreateParameter(createCtx)
	if createCtx.Writer.Status() != http.StatusCreated || len(stub.createCalls) != 1 || stub.createCalls[0].CreatedBy == nil || *stub.createCalls[0].CreatedBy != "tester@example.com" {
		t.Fatalf("expected create with actor, status %d calls %#v", createCtx.Writer.Status(), stub.createCalls)
	}

	updateCtx, _ := newBusinessParametersHandlerContext(http.MethodPut, "/api/v1/business-parameters/42", body)
	updateCtx.Params = gin.Params{{Key: "parameter_id", Value: "42"}}
	h.UpdateParameter(updateCtx)
	if updateCtx.Writer.Status() != http.StatusNoContent || len(stub.updateCalls) != 1 || stub.updateCalls[0].ID != 42 || stub.updateCalls[0].UpdatedBy == nil {
		t.Fatalf("expected update id 42 with actor, status %d calls %#v", updateCtx.Writer.Status(), stub.updateCalls)
	}

}
