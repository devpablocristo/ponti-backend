package category

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	domain "github.com/devpablocristo/ponti-backend/internal/category/usecases/domain"
)

type categoryHandlerUseCasesStub struct {
	listPage    int
	listPerPage int
	listTypeID  *int64
	getCalls    []int64
	deleteCalls []int64
	updateCalls []domain.Category
	createCalls []domain.Category
}

func (s *categoryHandlerUseCasesStub) CreateCategory(_ context.Context, category *domain.Category) (int64, error) {
	s.createCalls = append(s.createCalls, *category)
	return 99, nil
}

func (s *categoryHandlerUseCasesStub) ListCategories(_ context.Context, filters domain.ListFilters, page int, perPage int) ([]domain.Category, int64, error) {
	s.listPage = page
	s.listPerPage = perPage
	s.listTypeID = filters.TypeID
	return []domain.Category{{ID: 1, Name: "Semilla", TypeID: 2}}, 1, nil
}

func (s *categoryHandlerUseCasesStub) ListArchivedCategories(_ context.Context, page int, perPage int) ([]domain.Category, int64, error) {
	s.listPage = page
	s.listPerPage = perPage
	return []domain.Category{{ID: 1, Name: "Semilla", TypeID: 2}}, 1, nil
}

func (s *categoryHandlerUseCasesStub) GetCategory(_ context.Context, id int64) (*domain.Category, error) {
	s.getCalls = append(s.getCalls, id)
	return &domain.Category{ID: id, Name: "Semilla", TypeID: 2}, nil
}

func (s *categoryHandlerUseCasesStub) UpdateCategory(_ context.Context, category *domain.Category) error {
	s.updateCalls = append(s.updateCalls, *category)
	return nil
}

func (s *categoryHandlerUseCasesStub) ArchiveCategory(_ context.Context, id int64) error {
	s.deleteCalls = append(s.deleteCalls, id)
	return nil
}

func (s *categoryHandlerUseCasesStub) RestoreCategory(_ context.Context, id int64) error {
	s.deleteCalls = append(s.deleteCalls, id)
	return nil
}

func (s *categoryHandlerUseCasesStub) HardDeleteCategory(_ context.Context, id int64) error {
	s.deleteCalls = append(s.deleteCalls, id)
	return nil
}


func newCategoryHandlerContext(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx.Request = req
	return ctx, rec
}

func TestHandler_CategoryCRUDRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := &categoryHandlerUseCasesStub{}
	h := &Handler{ucs: stub}

	listCtx, _ := newCategoryHandlerContext(http.MethodGet, "/api/v1/categories?page=2&per_page=25", "")
	h.ListCategories(listCtx)
	if listCtx.Writer.Status() != http.StatusOK || stub.listPage != 2 || stub.listPerPage != 25 {
		t.Fatalf("expected list status OK with page 2/per_page 25, got status %d page %d per_page %d", listCtx.Writer.Status(), stub.listPage, stub.listPerPage)
	}

	getCtx, _ := newCategoryHandlerContext(http.MethodGet, "/api/v1/categories/42", "")
	getCtx.Params = gin.Params{{Key: "category_id", Value: "42"}}
	h.GetCategory(getCtx)
	if getCtx.Writer.Status() != http.StatusOK || len(stub.getCalls) != 1 || stub.getCalls[0] != 42 {
		t.Fatalf("expected get id 42, status %d calls %#v", getCtx.Writer.Status(), stub.getCalls)
	}

	createCtx, _ := newCategoryHandlerContext(http.MethodPost, "/api/v1/categories", `{"name":"Semilla","type_id":2}`)
	h.CreateCategory(createCtx)
	if createCtx.Writer.Status() != http.StatusCreated || len(stub.createCalls) != 1 || stub.createCalls[0].Name != "Semilla" || stub.createCalls[0].TypeID != 2 {
		t.Fatalf("expected create call, status %d calls %#v", createCtx.Writer.Status(), stub.createCalls)
	}

	updateCtx, _ := newCategoryHandlerContext(http.MethodPut, "/api/v1/categories/42", `{"name":"Fertilizantes","type_id":3}`)
	updateCtx.Params = gin.Params{{Key: "category_id", Value: "42"}}
	h.UpdateCategory(updateCtx)
	if updateCtx.Writer.Status() != http.StatusNoContent || len(stub.updateCalls) != 1 || stub.updateCalls[0].ID != 42 || stub.updateCalls[0].TypeID != 3 {
		t.Fatalf("expected update id 42, status %d calls %#v", updateCtx.Writer.Status(), stub.updateCalls)
	}

}
