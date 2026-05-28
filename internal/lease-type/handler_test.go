package leasetype

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	domain "github.com/devpablocristo/ponti-backend/internal/lease-type/usecases/domain"
)

type leaseTypeHandlerUseCasesStub struct {
	listPage    int
	listPerPage int
	getCalls    []int64
	deleteCalls []int64
	updateCalls []domain.LeaseType
	createCalls []domain.LeaseType
}

func (s *leaseTypeHandlerUseCasesStub) CreateLeaseType(_ context.Context, item *domain.LeaseType) (int64, error) {
	s.createCalls = append(s.createCalls, *item)
	return 99, nil
}

func (s *leaseTypeHandlerUseCasesStub) ListLeaseTypes(_ context.Context, page int, perPage int) ([]domain.LeaseType, int64, error) {
	s.listPage = page
	s.listPerPage = perPage
	return []domain.LeaseType{{ID: 1, Name: "Propio"}}, 1, nil
}

func (s *leaseTypeHandlerUseCasesStub) ListArchivedLeaseTypes(_ context.Context, page int, perPage int) ([]domain.LeaseType, int64, error) {
	s.listPage = page
	s.listPerPage = perPage
	return []domain.LeaseType{{ID: 1, Name: "Propio"}}, 1, nil
}

func (s *leaseTypeHandlerUseCasesStub) GetLeaseType(_ context.Context, id int64) (*domain.LeaseType, error) {
	s.getCalls = append(s.getCalls, id)
	return &domain.LeaseType{ID: id, Name: "Propio"}, nil
}

func (s *leaseTypeHandlerUseCasesStub) UpdateLeaseType(_ context.Context, item *domain.LeaseType) error {
	s.updateCalls = append(s.updateCalls, *item)
	return nil
}

func (s *leaseTypeHandlerUseCasesStub) ArchiveLeaseType(_ context.Context, id int64) error {
	s.deleteCalls = append(s.deleteCalls, id)
	return nil
}

func (s *leaseTypeHandlerUseCasesStub) RestoreLeaseType(_ context.Context, id int64) error {
	s.deleteCalls = append(s.deleteCalls, id)
	return nil
}

func (s *leaseTypeHandlerUseCasesStub) HardDeleteLeaseType(_ context.Context, id int64) error {
	s.deleteCalls = append(s.deleteCalls, id)
	return nil
}

func newLeaseTypeHandlerContext(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx.Request = req
	return ctx, rec
}

func TestHandler_LeaseTypeCRUDRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := &leaseTypeHandlerUseCasesStub{}
	h := &Handler{ucs: stub}

	listCtx, _ := newLeaseTypeHandlerContext(http.MethodGet, "/api/v1/lease-types?page=2&per_page=25", "")
	h.ListLeaseTypes(listCtx)
	if listCtx.Writer.Status() != http.StatusOK || stub.listPage != 2 || stub.listPerPage != 25 {
		t.Fatalf("expected list status OK with page 2/per_page 25, got status %d page %d per_page %d", listCtx.Writer.Status(), stub.listPage, stub.listPerPage)
	}

	getCtx, _ := newLeaseTypeHandlerContext(http.MethodGet, "/api/v1/lease-types/42", "")
	getCtx.Params = gin.Params{{Key: "lease_type_id", Value: "42"}}
	h.GetLeaseType(getCtx)
	if getCtx.Writer.Status() != http.StatusOK || len(stub.getCalls) != 1 || stub.getCalls[0] != 42 {
		t.Fatalf("expected get id 42, status %d calls %#v", getCtx.Writer.Status(), stub.getCalls)
	}

	createCtx, _ := newLeaseTypeHandlerContext(http.MethodPost, "/api/v1/lease-types", `{"name":"Propio"}`)
	h.CreateLeaseType(createCtx)
	if createCtx.Writer.Status() != http.StatusCreated || len(stub.createCalls) != 1 || stub.createCalls[0].Name != "Propio" {
		t.Fatalf("expected create call, status %d calls %#v", createCtx.Writer.Status(), stub.createCalls)
	}

	updateCtx, _ := newLeaseTypeHandlerContext(http.MethodPut, "/api/v1/lease-types/42", `{"name":"Arrendado"}`)
	updateCtx.Params = gin.Params{{Key: "lease_type_id", Value: "42"}}
	h.UpdateLeaseType(updateCtx)
	if updateCtx.Writer.Status() != http.StatusNoContent || len(stub.updateCalls) != 1 || stub.updateCalls[0].ID != 42 || stub.updateCalls[0].Name != "Arrendado" {
		t.Fatalf("expected update id 42, status %d calls %#v", updateCtx.Writer.Status(), stub.updateCalls)
	}

}
