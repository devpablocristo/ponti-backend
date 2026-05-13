package actor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/devpablocristo/core/security/go/contextkeys"
	"github.com/gin-gonic/gin"

	domain "github.com/devpablocristo/ponti-backend/internal/actor/usecases/domain"
)

type actorHandlerUseCasesStub struct {
	listFilters     []domain.ListFilters
	listPages       []int
	listPerPages    []int
	archiveCalls    []int64
	restoreCalls    []int64
	hardDeleteCalls []int64
	mergeCalls      []domain.MergeRequest
}

func (s *actorHandlerUseCasesStub) CreateActor(context.Context, *domain.Actor) (int64, error) {
	return 0, nil
}

func (s *actorHandlerUseCasesStub) ListActors(_ context.Context, filters domain.ListFilters, page int, perPage int) ([]domain.Actor, int64, error) {
	s.listFilters = append(s.listFilters, filters)
	s.listPages = append(s.listPages, page)
	s.listPerPages = append(s.listPerPages, perPage)
	return []domain.Actor{{ID: 1, DisplayName: "Actor 1"}}, 1, nil
}

func (s *actorHandlerUseCasesStub) GetActor(context.Context, int64) (*domain.Actor, error) {
	return &domain.Actor{ID: 1, DisplayName: "Actor 1"}, nil
}

func (s *actorHandlerUseCasesStub) UpdateActor(context.Context, *domain.Actor) error {
	return nil
}

func (s *actorHandlerUseCasesStub) ArchiveActor(_ context.Context, id int64) error {
	s.archiveCalls = append(s.archiveCalls, id)
	return nil
}

func (s *actorHandlerUseCasesStub) RestoreActor(_ context.Context, id int64) error {
	s.restoreCalls = append(s.restoreCalls, id)
	return nil
}

func (s *actorHandlerUseCasesStub) HardDeleteActor(_ context.Context, id int64) error {
	s.hardDeleteCalls = append(s.hardDeleteCalls, id)
	return nil
}

func (s *actorHandlerUseCasesStub) AddRole(context.Context, int64, string) error {
	return nil
}

func (s *actorHandlerUseCasesStub) AddAlias(context.Context, int64, domain.ActorAlias) (int64, error) {
	return 0, nil
}

func (s *actorHandlerUseCasesStub) ListDuplicateCandidates(context.Context) ([]domain.DuplicateCandidate, error) {
	return nil, nil
}

func (s *actorHandlerUseCasesStub) MergeActors(_ context.Context, req domain.MergeRequest) (*domain.MergeImpact, error) {
	s.mergeCalls = append(s.mergeCalls, req)
	return &domain.MergeImpact{
		TargetActorID:  req.TargetActorID,
		SourceActorIDs: req.SourceActorIDs,
		Confirmed:      req.Confirm,
	}, nil
}

func newActorHandlerContext(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), ctxkeys.Actor, "tester@example.com"))
	ctx.Request = req
	return ctx, rec
}

func TestHandler_ListActors_ParsesFiltersAndPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := &actorHandlerUseCasesStub{}
	h := &Handler{ucs: stub}
	ctx, _ := newActorHandlerContext(http.MethodGet, "/api/v1/actors?status=all&role=inversor&q=agro&page=2&per_page=25", "")

	h.ListActors(ctx)

	if ctx.Writer.Status() != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, ctx.Writer.Status())
	}
	if len(stub.listFilters) != 1 {
		t.Fatalf("expected one ListActors call, got %#v", stub.listFilters)
	}
	got := stub.listFilters[0]
	if got.Status != "all" || got.Role != "inversor" || got.Query != "agro" {
		t.Fatalf("unexpected filters: %#v", got)
	}
	if stub.listPages[0] != 2 || stub.listPerPages[0] != 25 {
		t.Fatalf("expected page 2/per_page 25, got %d/%d", stub.listPages[0], stub.listPerPages[0])
	}
}

func TestHandler_ListArchivedActors_ForcesArchivedStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := &actorHandlerUseCasesStub{}
	h := &Handler{ucs: stub}
	ctx, _ := newActorHandlerContext(http.MethodGet, "/api/v1/actors/archived?status=active&role=cliente&q=agro", "")

	h.ListArchivedActors(ctx)

	if ctx.Writer.Status() != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, ctx.Writer.Status())
	}
	if len(stub.listFilters) != 1 {
		t.Fatalf("expected one ListActors call, got %#v", stub.listFilters)
	}
	got := stub.listFilters[0]
	if got.Status != "archived" || got.Role != "cliente" || got.Query != "agro" {
		t.Fatalf("unexpected filters: %#v", got)
	}
}

func TestHandler_ActorArchiveRestoreHardDeleteRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name  string
		run   func(*Handler, *gin.Context)
		calls func(*actorHandlerUseCasesStub) []int64
	}{
		{
			name:  "archive",
			run:   (*Handler).ArchiveActor,
			calls: func(s *actorHandlerUseCasesStub) []int64 { return s.archiveCalls },
		},
		{
			name:  "restore",
			run:   (*Handler).RestoreActor,
			calls: func(s *actorHandlerUseCasesStub) []int64 { return s.restoreCalls },
		},
		{
			name:  "hard delete",
			run:   (*Handler).HardDeleteActor,
			calls: func(s *actorHandlerUseCasesStub) []int64 { return s.hardDeleteCalls },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := &actorHandlerUseCasesStub{}
			h := &Handler{ucs: stub}
			ctx, _ := newActorHandlerContext(http.MethodPost, "/api/v1/actors/42", "")
			ctx.Params = gin.Params{{Key: "actor_id", Value: "42"}}

			tt.run(h, ctx)

			if ctx.Writer.Status() != http.StatusNoContent {
				t.Fatalf("expected status %d, got %d", http.StatusNoContent, ctx.Writer.Status())
			}
			calls := tt.calls(stub)
			if len(calls) != 1 || calls[0] != 42 {
				t.Fatalf("expected action to be called with id 42, got %#v", calls)
			}
		})
	}
}

func TestHandler_MergeActors_UsesAuthenticatedActor(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := &actorHandlerUseCasesStub{}
	h := &Handler{ucs: stub}
	ctx, _ := newActorHandlerContext(http.MethodPost, "/api/v1/actors/merge", `{
		"target_actor_id": 10,
		"source_actor_ids": [11, 12],
		"reason": "manual review",
		"confirm": true
	}`)

	h.MergeActors(ctx)

	if ctx.Writer.Status() != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, ctx.Writer.Status())
	}
	if len(stub.mergeCalls) != 1 {
		t.Fatalf("expected one merge call, got %#v", stub.mergeCalls)
	}
	got := stub.mergeCalls[0]
	if got.TargetActorID != 10 || len(got.SourceActorIDs) != 2 || got.SourceActorIDs[0] != 11 || got.SourceActorIDs[1] != 12 {
		t.Fatalf("unexpected merge ids: %#v", got)
	}
	if got.Reason != "manual review" || !got.Confirm || got.MergedBy != "tester@example.com" {
		t.Fatalf("unexpected merge metadata: %#v", got)
	}
}
