package businessinsights

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	ctxkeys "github.com/devpablocristo/core/security/go/contextkeys"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type businessInsightsListRepoStub struct {
	tenantID string
	userID   string
	opts     ListOptions
}

func (s *businessInsightsListRepoStub) ListByTenantForUser(_ context.Context, tenantID, userID string, opts ListOptions) ([]CandidateView, error) {
	s.tenantID = tenantID
	s.userID = userID
	s.opts = opts
	now := time.Date(2026, 5, 12, 10, 0, 0, 0, time.UTC)
	return []CandidateView{{
		CandidateRecord: CandidateRecord{
			ID:              "candidate-1",
			Kind:            "insight",
			EventType:       "ponti.stock.negative",
			EntityType:      "supply",
			EntityID:        "7",
			Severity:        "warning",
			Status:          "new",
			Title:           "Stock negativo",
			Body:            "Revisar stock",
			OccurrenceCount: 1,
			FirstSeenAt:     now,
			LastSeenAt:      now,
		},
	}}, nil
}

func newBusinessInsightsHandlerContext(method, target string, tenantID uuid.UUID) (*gin.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(method, target, strings.NewReader(""))
	req = req.WithContext(context.WithValue(req.Context(), ctxkeys.Actor, "tester@example.com"))
	req = req.WithContext(context.WithValue(req.Context(), ctxkeys.OrgID, tenantID))
	ctx.Request = req
	return ctx, rec
}

func TestHandler_BusinessInsightsList_UsesTenantUserAndOptions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tenantID := uuid.New()
	repo := &businessInsightsListRepoStub{}
	h := &Handler{repo: repo, svc: NewService(nil, nil, nil, nil, Config{})}
	ctx, _ := newBusinessInsightsHandlerContext(http.MethodGet, "/api/v1/insights?limit=25&include_resolved=true", tenantID)

	h.List(ctx)

	if ctx.Writer.Status() != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, ctx.Writer.Status())
	}
	if repo.tenantID != tenantID.String() || repo.userID != "tester@example.com" {
		t.Fatalf("expected tenant/user %s/tester@example.com, got %s/%s", tenantID, repo.tenantID, repo.userID)
	}
	if repo.opts.Limit != 25 || !repo.opts.IncludeResolved {
		t.Fatalf("expected limit 25 include_resolved true, got %#v", repo.opts)
	}
}

func TestHandler_BusinessInsightsActions_ParseIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tenantID := uuid.New()
	h := &Handler{svc: NewService(nil, nil, nil, nil, Config{})}

	tests := []struct {
		name   string
		method string
		run    func(*Handler, *gin.Context)
	}{
		{name: "mark read", method: http.MethodPost, run: (*Handler).MarkRead},
		{name: "mark unread", method: http.MethodDelete, run: (*Handler).MarkUnread},
		{name: "resolve", method: http.MethodPost, run: (*Handler).Resolve},
		{name: "reopen", method: http.MethodDelete, run: (*Handler).Reopen},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, _ := newBusinessInsightsHandlerContext(tt.method, "/api/v1/insights/candidate-1", tenantID)
			ctx.Params = gin.Params{{Key: "id", Value: "candidate-1"}}

			tt.run(h, ctx)

			if ctx.Writer.Status() != http.StatusNoContent {
				t.Fatalf("expected status %d, got %d", http.StatusNoContent, ctx.Writer.Status())
			}
		})
	}
}
