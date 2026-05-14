package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	domain "github.com/devpablocristo/ponti-backend/internal/provider/usecases/domain"
)

func TestGetProvidersReturnsProviderList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	actorID := int64(99)
	ucs := &providerUseCasesStub{
		providers: []domain.Provider{
			{ID: 1, Name: "Proveedor A", ActorID: &actorID},
			{ID: 2, Name: "Proveedor B"},
		},
	}
	h := &Handler{ucs: ucs}

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/providers", nil)

	h.GetProviders(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var payload []struct {
		ID      int64  `json:"id"`
		Name    string `json:"name"`
		ActorID *int64 `json:"actor_id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload) != 2 {
		t.Fatalf("expected two providers, got %d", len(payload))
	}
	if payload[0].ID != 1 || payload[0].Name != "Proveedor A" || payload[0].ActorID == nil || *payload[0].ActorID != actorID {
		t.Fatalf("unexpected first provider: %+v", payload[0])
	}
	if payload[1].ID != 2 || payload[1].Name != "Proveedor B" || payload[1].ActorID != nil {
		t.Fatalf("unexpected second provider: %+v", payload[1])
	}
}

type providerUseCasesStub struct {
	providers []domain.Provider
}

func (s *providerUseCasesStub) GetProviders(context.Context) ([]domain.Provider, error) {
	return s.providers, nil
}

func (s *providerUseCasesStub) ListArchivedProviders(context.Context) ([]domain.Provider, error) {
	return s.providers, nil
}

func (s *providerUseCasesStub) GetProvider(_ context.Context, id int64) (*domain.Provider, error) {
	return &domain.Provider{ID: id, Name: "Proveedor A"}, nil
}

func (s *providerUseCasesStub) CreateProvider(context.Context, *domain.Provider) (int64, error) {
	return 99, nil
}

func (s *providerUseCasesStub) UpdateProvider(context.Context, *domain.Provider) error {
	return nil
}

func (s *providerUseCasesStub) ArchiveProvider(context.Context, int64) error {
	return nil
}

func (s *providerUseCasesStub) RestoreProvider(context.Context, int64) error {
	return nil
}

func (s *providerUseCasesStub) HardDeleteProvider(context.Context, int64) error {
	return nil
}

func (s *providerUseCasesStub) DeleteProvider(context.Context, int64) error {
	return nil
}
