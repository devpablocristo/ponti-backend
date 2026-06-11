package businessinsights_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	ctxkeys "github.com/devpablocristo/platform/security/go/contextkeys"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/devpablocristo/ponti-backend/internal/businessinsights"
)

type handlerRepoStub struct {
	views []businessinsights.CandidateView
}

func (s handlerRepoStub) ListByTenantForUser(_ context.Context, _ string, _ string, _ businessinsights.ListOptions) ([]businessinsights.CandidateView, error) {
	return s.views, nil
}

func (s handlerRepoStub) GetByIDForTenant(_ context.Context, _ string, candidateID, _ string) (businessinsights.CandidateView, error) {
	for _, view := range s.views {
		if view.ID == candidateID {
			return view, nil
		}
	}
	return businessinsights.CandidateView{}, domainerr.NotFound("insight not found")
}

type handlerTestEngine struct {
	router *gin.Engine
}

func (e handlerTestEngine) GetRouter() *gin.Engine { return e.router }

type handlerTestConfig struct{}

func (handlerTestConfig) APIBaseURL() string { return "/api/v1" }

type handlerTestMiddlewares struct {
	orgID  uuid.UUID
	actor  string
	scopes []string
}

func (m handlerTestMiddlewares) GetValidation() []gin.HandlerFunc {
	return []gin.HandlerFunc{func(c *gin.Context) {
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, ctxkeys.OrgID, m.orgID)
		ctx = context.WithValue(ctx, ctxkeys.Actor, m.actor)
		ctx = context.WithValue(ctx, ctxkeys.Scopes, m.scopes)
		c.Request = c.Request.WithContext(ctx)
		c.Set(string(ctxkeys.OrgID), m.orgID)
		c.Set(string(ctxkeys.Actor), m.actor)
		c.Next()
	}}
}

func TestHandlerSummaryReturnsEvidenceAndAggregates(t *testing.T) {
	t.Parallel()
	orgID := uuid.New()
	router := newBusinessInsightsTestRouter(orgID, []businessinsights.CandidateView{
		candidateView("cand-1", "stock", "critical", "new"),
		candidateView("cand-2", "integrity", "medium", "resolved"),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/insights/summary", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	summary := body["summary"].(map[string]any)
	if summary["total"].(float64) != 2 {
		t.Fatalf("unexpected summary: %#v", summary)
	}
	if summary["open"].(float64) != 1 || summary["resolved"].(float64) != 1 {
		t.Fatalf("unexpected open/resolved counts: %#v", summary)
	}
	evidence := body["evidence"].(map[string]any)
	if evidence["tenant_scope"] != orgID.String() {
		t.Fatalf("tenant evidence mismatch: %#v", evidence)
	}
	if evidence["source_ref"] != "ponti.businessinsights.summary" {
		t.Fatalf("source ref mismatch: %#v", evidence)
	}
}

func TestHandlerExplainReturnsTenantScopedEvidence(t *testing.T) {
	t.Parallel()
	orgID := uuid.New()
	router := newBusinessInsightsTestRouter(orgID, []businessinsights.CandidateView{
		candidateView("cand-1", "stock", "critical", "new"),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/insights/cand-1/explain", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	insight := body["insight"].(map[string]any)
	if insight["id"] != "cand-1" {
		t.Fatalf("unexpected insight: %#v", insight)
	}
	evidence := body["evidence"].(map[string]any)
	if evidence["tenant_scope"] != orgID.String() {
		t.Fatalf("tenant evidence mismatch: %#v", evidence)
	}
	entity := evidence["entity"].(map[string]any)
	if entity["type"] != "supply" || entity["id"] != "supply-cand-1" {
		t.Fatalf("entity evidence mismatch: %#v", entity)
	}
}

func newBusinessInsightsTestRouter(orgID uuid.UUID, views []businessinsights.CandidateView) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := businessinsights.NewHandler(
		handlerRepoStub{views: views},
		nil,
		handlerTestEngine{router: router},
		handlerTestConfig{},
		handlerTestMiddlewares{orgID: orgID, actor: "user-1", scopes: []string{"api.read"}},
	)
	handler.Routes()
	return router
}

func candidateView(id, kind, severity, status string) businessinsights.CandidateView {
	now := time.Date(2026, 6, 8, 12, 0, 0, 0, time.UTC)
	return businessinsights.CandidateView{
		CandidateRecord: businessinsights.CandidateRecord{
			ID:              id,
			TenantID:        uuid.NewString(),
			Kind:            kind,
			EventType:       "stock_negative",
			EntityType:      "supply",
			EntityID:        "supply-" + id,
			Fingerprint:     "fp-" + id,
			Severity:        severity,
			Status:          status,
			Title:           "Insight " + id,
			Body:            "Body " + id,
			Evidence:        map[string]any{"domain_source": "test"},
			OccurrenceCount: 1,
			FirstSeenAt:     now,
			LastSeenAt:      now,
		},
	}
}
