package classtype

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	domain "github.com/devpablocristo/ponti-backend/internal/class-type/usecases/domain"
)

type classTypeHandlerUseCasesStub struct {
	listPage    int
	listPerPage int
	getCalls    []int64
	deleteCalls []int64
	updateCalls []domain.ClassType
	createCalls []domain.ClassType
}

func (s *classTypeHandlerUseCasesStub) CreateClassType(_ context.Context, item *domain.ClassType) (int64, error) {
	s.createCalls = append(s.createCalls, *item)
	return 99, nil
}

func (s *classTypeHandlerUseCasesStub) ListClassTypes(_ context.Context, page int, perPage int) ([]domain.ClassType, int64, error) {
	s.listPage = page
	s.listPerPage = perPage
	return []domain.ClassType{{ID: 1, Name: "Insumos"}}, 1, nil
}

func (s *classTypeHandlerUseCasesStub) GetClassType(_ context.Context, id int64) (*domain.ClassType, error) {
	s.getCalls = append(s.getCalls, id)
	return &domain.ClassType{ID: id, Name: "Insumos"}, nil
}

func (s *classTypeHandlerUseCasesStub) UpdateClassType(_ context.Context, item *domain.ClassType) error {
	s.updateCalls = append(s.updateCalls, *item)
	return nil
}

func (s *classTypeHandlerUseCasesStub) DeleteClassType(_ context.Context, id int64) error {
	s.deleteCalls = append(s.deleteCalls, id)
	return nil
}

func newClassTypeHandlerContext(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx.Request = req
	return ctx, rec
}

func TestHandler_ClassTypeCRUDRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := &classTypeHandlerUseCasesStub{}
	h := &Handler{ucs: stub}

	listCtx, _ := newClassTypeHandlerContext(http.MethodGet, "/api/v1/types?page=2&per_page=25", "")
	h.ListClassTypes(listCtx)
	if listCtx.Writer.Status() != http.StatusOK || stub.listPage != 2 || stub.listPerPage != 25 {
		t.Fatalf("expected list status OK with page 2/per_page 25, got status %d page %d per_page %d", listCtx.Writer.Status(), stub.listPage, stub.listPerPage)
	}

	getCtx, _ := newClassTypeHandlerContext(http.MethodGet, "/api/v1/types/42", "")
	getCtx.Params = gin.Params{{Key: "class_type_id", Value: "42"}}
	h.GetClassType(getCtx)
	if getCtx.Writer.Status() != http.StatusOK || len(stub.getCalls) != 1 || stub.getCalls[0] != 42 {
		t.Fatalf("expected get id 42, status %d calls %#v", getCtx.Writer.Status(), stub.getCalls)
	}

	createCtx, _ := newClassTypeHandlerContext(http.MethodPost, "/api/v1/types", `{"name":"Insumos"}`)
	h.CreateClassType(createCtx)
	if createCtx.Writer.Status() != http.StatusCreated || len(stub.createCalls) != 1 || stub.createCalls[0].Name != "Insumos" {
		t.Fatalf("expected create call, status %d calls %#v", createCtx.Writer.Status(), stub.createCalls)
	}

	updateCtx, _ := newClassTypeHandlerContext(http.MethodPut, "/api/v1/types/42", `{"name":"Labores"}`)
	updateCtx.Params = gin.Params{{Key: "class_type_id", Value: "42"}}
	h.UpdateClassType(updateCtx)
	if updateCtx.Writer.Status() != http.StatusNoContent || len(stub.updateCalls) != 1 || stub.updateCalls[0].ID != 42 || stub.updateCalls[0].Name != "Labores" {
		t.Fatalf("expected update id 42, status %d calls %#v", updateCtx.Writer.Status(), stub.updateCalls)
	}

	deleteCtx, _ := newClassTypeHandlerContext(http.MethodDelete, "/api/v1/types/42", "")
	deleteCtx.Params = gin.Params{{Key: "class_type_id", Value: "42"}}
	h.DeleteClassType(deleteCtx)
	if deleteCtx.Writer.Status() != http.StatusNoContent || len(stub.deleteCalls) != 1 || stub.deleteCalls[0] != 42 {
		t.Fatalf("expected delete id 42, status %d calls %#v", deleteCtx.Writer.Status(), stub.deleteCalls)
	}
}
