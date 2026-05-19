package crop

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	domain "github.com/devpablocristo/ponti-backend/internal/crop/usecases/domain"
)

type cropHandlerUseCasesStub struct {
	listPage    int
	listPerPage int
	getCalls    []int64
	deleteCalls []int64
	updateCalls []domain.Crop
	createCalls []domain.Crop
}

func (s *cropHandlerUseCasesStub) CreateCrop(_ context.Context, crop *domain.Crop) (int64, error) {
	s.createCalls = append(s.createCalls, *crop)
	return 99, nil
}

func (s *cropHandlerUseCasesStub) ListCrops(_ context.Context, page int, perPage int) ([]domain.Crop, int64, error) {
	s.listPage = page
	s.listPerPage = perPage
	return []domain.Crop{{ID: 1, Name: "Soja"}}, 1, nil
}

func (s *cropHandlerUseCasesStub) ListArchivedCrops(_ context.Context, page int, perPage int) ([]domain.Crop, int64, error) {
	s.listPage = page
	s.listPerPage = perPage
	return []domain.Crop{{ID: 1, Name: "Soja"}}, 1, nil
}

func (s *cropHandlerUseCasesStub) GetCrop(_ context.Context, id int64) (*domain.Crop, error) {
	s.getCalls = append(s.getCalls, id)
	return &domain.Crop{ID: id, Name: "Soja"}, nil
}

func (s *cropHandlerUseCasesStub) UpdateCrop(_ context.Context, crop *domain.Crop) error {
	s.updateCalls = append(s.updateCalls, *crop)
	return nil
}

func (s *cropHandlerUseCasesStub) ArchiveCrop(_ context.Context, id int64) error {
	s.deleteCalls = append(s.deleteCalls, id)
	return nil
}

func (s *cropHandlerUseCasesStub) RestoreCrop(_ context.Context, id int64) error {
	s.deleteCalls = append(s.deleteCalls, id)
	return nil
}

func (s *cropHandlerUseCasesStub) HardDeleteCrop(_ context.Context, id int64) error {
	s.deleteCalls = append(s.deleteCalls, id)
	return nil
}

func (s *cropHandlerUseCasesStub) DeleteCrop(_ context.Context, id int64) error {
	s.deleteCalls = append(s.deleteCalls, id)
	return nil
}

func newCropHandlerContext(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx.Request = req
	return ctx, rec
}

func TestHandler_CropCRUDRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := &cropHandlerUseCasesStub{}
	h := &Handler{ucs: stub}

	listCtx, _ := newCropHandlerContext(http.MethodGet, "/api/v1/crops?page=2&per_page=25", "")
	h.ListCrops(listCtx)
	if listCtx.Writer.Status() != http.StatusOK || stub.listPage != 2 || stub.listPerPage != 25 {
		t.Fatalf("expected list status OK with page 2/per_page 25, got status %d page %d per_page %d", listCtx.Writer.Status(), stub.listPage, stub.listPerPage)
	}

	getCtx, _ := newCropHandlerContext(http.MethodGet, "/api/v1/crops/42", "")
	getCtx.Params = gin.Params{{Key: "crop_id", Value: "42"}}
	h.GetCrop(getCtx)
	if getCtx.Writer.Status() != http.StatusOK || len(stub.getCalls) != 1 || stub.getCalls[0] != 42 {
		t.Fatalf("expected get id 42, status %d calls %#v", getCtx.Writer.Status(), stub.getCalls)
	}

	createCtx, _ := newCropHandlerContext(http.MethodPost, "/api/v1/crops", `{"name":"Soja"}`)
	h.CreateCrop(createCtx)
	if createCtx.Writer.Status() != http.StatusCreated || len(stub.createCalls) != 1 || stub.createCalls[0].Name != "Soja" {
		t.Fatalf("expected create call, status %d calls %#v", createCtx.Writer.Status(), stub.createCalls)
	}

	updateCtx, _ := newCropHandlerContext(http.MethodPut, "/api/v1/crops/42", `{"name":"Maiz"}`)
	updateCtx.Params = gin.Params{{Key: "crop_id", Value: "42"}}
	h.UpdateCrop(updateCtx)
	if updateCtx.Writer.Status() != http.StatusNoContent || len(stub.updateCalls) != 1 || stub.updateCalls[0].ID != 42 || stub.updateCalls[0].Name != "Maiz" {
		t.Fatalf("expected update id 42, status %d calls %#v", updateCtx.Writer.Status(), stub.updateCalls)
	}

	deleteCtx, _ := newCropHandlerContext(http.MethodDelete, "/api/v1/crops/42", "")
	deleteCtx.Params = gin.Params{{Key: "crop_id", Value: "42"}}
	h.DeleteCrop(deleteCtx)
	if deleteCtx.Writer.Status() != http.StatusNoContent || len(stub.deleteCalls) != 1 || stub.deleteCalls[0] != 42 {
		t.Fatalf("expected delete id 42, status %d calls %#v", deleteCtx.Writer.Status(), stub.deleteCalls)
	}
}
