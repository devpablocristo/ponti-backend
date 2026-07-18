package governance_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	"github.com/google/uuid"

	"github.com/devpablocristo/ponti-backend/internal/governance"
	nexusclient "github.com/devpablocristo/ponti-backend/internal/nexus"
)

type stubRepo struct {
	rows              map[string]governance.RequestRecord // key: nexus_request_id
	createCalls       int
	updateCalls       int
	lastUpdates       map[string]any
	saveEvidenceCalls int
}

func newStubRepo() *stubRepo {
	return &stubRepo{rows: map[string]governance.RequestRecord{}}
}

func (s *stubRepo) GetByNexusRequestID(_ context.Context, tenantID uuid.UUID, nexusRequestID string) (governance.RequestRecord, error) {
	row, ok := s.rows[nexusRequestID]
	if !ok || row.TenantID != tenantID {
		return governance.RequestRecord{}, notFoundErr()
	}
	return row, nil
}

func (s *stubRepo) Create(_ context.Context, row governance.RequestRecord) error {
	s.createCalls++
	s.rows[row.NexusRequestID] = row
	return nil
}

func (s *stubRepo) UpdateByNexusRequestID(_ context.Context, tenantID uuid.UUID, nexusRequestID string, updates map[string]any) error {
	s.updateCalls++
	s.lastUpdates = updates
	row := s.rows[nexusRequestID]
	if status, ok := updates["status"].(string); ok {
		row.Status = status
	}
	if approvalID, ok := updates["approval_id"].(string); ok {
		row.ApprovalID = approvalID
	}
	if decidedBy, ok := updates["decided_by"].(string); ok {
		row.DecidedBy = decidedBy
	}
	row.TenantID = tenantID
	row.NexusRequestID = nexusRequestID
	s.rows[nexusRequestID] = row
	return nil
}

func (s *stubRepo) ListHistory(_ context.Context, _ uuid.UUID, _ int) ([]governance.RequestRecord, error) {
	return nil, nil
}

func (s *stubRepo) GetEvidence(_ context.Context, _ uuid.UUID, _ string) (governance.EvidenceRecord, error) {
	return governance.EvidenceRecord{}, notFoundErr()
}

func (s *stubRepo) SaveEvidence(_ context.Context, _ governance.EvidenceRecord) error {
	s.saveEvidenceCalls++
	return nil
}

type stubNexus struct {
	getCalls        int
	getResponse     nexusclient.Request
	getStatus       int
	approvals       []nexusclient.Approval
	approvalQueries []string
	approveCalls    int
	rejectCalls     int
}

func (s *stubNexus) Get(_ context.Context, _ string, _ ...nexusclient.RequestOption) (nexusclient.Request, int, error) {
	s.getCalls++
	if s.getStatus == 0 {
		s.getStatus = http.StatusOK
	}
	return s.getResponse, s.getStatus, nil
}

func (s *stubNexus) ListRequests(_ context.Context, _ string, _ ...nexusclient.RequestOption) ([]nexusclient.Request, error) {
	return nil, nil
}

func (s *stubNexus) ListPendingApprovals(_ context.Context, query string, _ ...nexusclient.RequestOption) ([]nexusclient.Approval, error) {
	s.approvalQueries = append(s.approvalQueries, query)
	return s.approvals, nil
}

func (s *stubNexus) Approve(_ context.Context, _, _, _ string, _ ...nexusclient.RequestOption) (int, []byte, error) {
	s.approveCalls++
	return http.StatusOK, []byte(`{"status":"approved"}`), nil
}

func (s *stubNexus) Reject(_ context.Context, _, _, _ string, _ ...nexusclient.RequestOption) (int, []byte, error) {
	s.rejectCalls++
	return http.StatusOK, []byte(`{"status":"rejected"}`), nil
}

func (s *stubNexus) GetEvidence(_ context.Context, _ string, _ ...nexusclient.RequestOption) ([]byte, int, error) {
	return nil, http.StatusNotFound, nil
}

type stubExecutor struct {
	calls   int
	lastRow governance.RequestRecord
}

func (s *stubExecutor) ExecuteApproved(_ context.Context, row governance.RequestRecord) error {
	s.calls++
	s.lastRow = row
	return nil
}

func notFoundErr() error {
	return domainerr.NotFound("governance request not found")
}

// fakeNexus es un servidor Nexus falso (httptest + cliente HTTP real) para
// probar el filtrado por org y los query params que Ponti manda de verdad.
type fakeNexus struct {
	mu              sync.Mutex
	requests        []nexusclient.Request
	approvals       []nexusclient.Approval
	evidence        map[string]string // request_id -> pack JSON crudo
	honorFilters    bool              // false simula un Nexus viejo que ignora query params
	requestQueries  []string
	approvalQueries []string
	decidePaths     []string
}

func (f *fakeNexus) client(t *testing.T) *nexusclient.Client {
	t.Helper()
	writeJSON := func(w http.ResponseWriter, status int, v any) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(v)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/requests", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		f.requestQueries = append(f.requestQueries, r.URL.RawQuery)
		data := append([]nexusclient.Request(nil), f.requests...)
		f.mu.Unlock()
		writeJSON(w, http.StatusOK, map[string]any{"data": data})
	})
	mux.HandleFunc("GET /v1/requests/{id}", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		for _, req := range f.requests {
			if req.ID == r.PathValue("id") {
				writeJSON(w, http.StatusOK, req)
				return
			}
		}
		writeJSON(w, http.StatusNotFound, map[string]any{"code": "NOT_FOUND", "message": "request not found"})
	})
	mux.HandleFunc("GET /v1/requests/{id}/evidence", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		pack, ok := f.evidence[r.PathValue("id")]
		f.mu.Unlock()
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]any{"code": "NOT_FOUND", "message": "evidence not found"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(pack))
	})
	mux.HandleFunc("GET /v1/approvals/pending", func(w http.ResponseWriter, r *http.Request) {
		requestID := r.URL.Query().Get("request_id")
		f.mu.Lock()
		f.approvalQueries = append(f.approvalQueries, r.URL.RawQuery)
		items := make([]nexusclient.Approval, 0, len(f.approvals))
		for _, a := range f.approvals {
			if f.honorFilters && requestID != "" && a.RequestID != requestID {
				continue
			}
			items = append(items, a)
		}
		f.mu.Unlock()
		writeJSON(w, http.StatusOK, map[string]any{"data": items})
	})
	mux.HandleFunc("POST /v1/approvals/{id}/{action}", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		f.decidePaths = append(f.decidePaths, r.URL.Path)
		f.mu.Unlock()
		writeJSON(w, http.StatusOK, map[string]any{"status": "approved"})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return nexusclient.NewClient(srv.URL, "test-key", 0)
}

// signCallback replica el algoritmo del publisher de Nexus
// (nexus/internal/callbacks/outbox.go) para generar vectores de test.
func signCallback(token, timestamp string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(token))
	_, _ = mac.Write([]byte(timestamp))
	_, _ = mac.Write([]byte("."))
	_, _ = mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestVerifyCallbackSignature_ValidVector(t *testing.T) {
	t.Parallel()
	svc := governance.NewService(newStubRepo(), nil, governance.Config{CallbackToken: "nexus-callback-secret"}, nil)

	timestamp := "2026-06-10T12:00:00.000000000Z"
	payload := []byte(`{"event":"approval_pending","approval_id":"appr-1","org_id":"3f2504e0-4f89-11d3-9a0c-0305e82c3301","request_id":"req-1","action_type":"workorder.create","risk_level":"high","created_at":"2026-06-10T11:59:00Z"}`)

	// Vector fijo precalculado con HMAC-SHA256(token, timestamp + "." + payload).
	fixed := "sha256=0fcc591dda200b29985f84c9da0479fafa3e73e5a8380eb5da49958ab4659f0b"
	if got := signCallback("nexus-callback-secret", timestamp, payload); got != fixed {
		t.Fatalf("signCallback vector mismatch: %s", got)
	}
	if !svc.VerifyCallbackSignature(timestamp, payload, fixed) {
		t.Fatal("expected valid signature to verify")
	}
}

func TestVerifyCallbackSignature_Invalid(t *testing.T) {
	t.Parallel()
	svc := governance.NewService(newStubRepo(), nil, governance.Config{CallbackToken: "nexus-callback-secret"}, nil)
	timestamp := "2026-06-10T12:00:00.000000000Z"
	payload := []byte(`{"event":"approval_pending","request_id":"req-1"}`)
	valid := signCallback("nexus-callback-secret", timestamp, payload)

	if svc.VerifyCallbackSignature(timestamp, []byte(`{"event":"approval_pending","request_id":"req-TAMPERED"}`), valid) {
		t.Error("tampered payload must not verify")
	}
	if svc.VerifyCallbackSignature("2026-06-10T12:00:01Z", payload, valid) {
		t.Error("tampered timestamp must not verify")
	}
	if svc.VerifyCallbackSignature(timestamp, payload, signCallback("other-token", timestamp, payload)) {
		t.Error("signature with wrong token must not verify")
	}
	if svc.VerifyCallbackSignature(timestamp, payload, "") {
		t.Error("empty signature must not verify")
	}

	unconfigured := governance.NewService(newStubRepo(), nil, governance.Config{}, nil)
	if unconfigured.VerifyCallbackSignature(timestamp, payload, valid) {
		t.Error("service without token must not verify")
	}
}

func TestVerifyCallbackTimestamp_Freshness(t *testing.T) {
	t.Parallel()
	svc := governance.NewService(newStubRepo(), nil, governance.Config{CallbackToken: "tok"}, nil)
	now := time.Now().UTC()

	cases := []struct {
		name      string
		timestamp string
		want      bool
	}{
		{"fresh RFC3339Nano", now.Format(time.RFC3339Nano), true},
		{"fresh RFC3339 sin nanos", now.Format(time.RFC3339), true},
		{"dentro de la ventana", now.Add(-4 * time.Minute).Format(time.RFC3339Nano), true},
		{"stale mas de 5 minutos", now.Add(-10 * time.Minute).Format(time.RFC3339Nano), false},
		{"futuro mas de 5 minutos", now.Add(10 * time.Minute).Format(time.RFC3339Nano), false},
		{"vacio", "", false},
		{"basura", "not-a-timestamp", false},
	}
	for _, tc := range cases {
		if got := svc.VerifyCallbackTimestamp(tc.timestamp); got != tc.want {
			t.Errorf("%s: VerifyCallbackTimestamp(%q) = %v, want %v", tc.name, tc.timestamp, got, tc.want)
		}
	}
}

func TestHandleCallback_PendingCreatesRowFromNexus(t *testing.T) {
	t.Parallel()
	tenantID := uuid.New()
	repo := newStubRepo()
	nx := &stubNexus{getResponse: nexusclient.Request{
		ID:          "req-1",
		OrgID:       tenantID.String(),
		RequesterID: "agent:ponti-ops",
		ActionType:  "workorder.create",
		Status:      nexusclient.StatusPendingApproval,
		Decision:    nexusclient.DecisionRequireApproval,
		RiskLevel:   "high",
		BindingHash: "hash-1",
		Reason:      "draft listo",
	}}
	svc := governance.NewService(repo, nx, governance.Config{CallbackToken: "tok"}, nil)

	err := svc.HandleCallback(context.Background(), governance.CallbackEvent{
		Event:      governance.EventApprovalPending,
		ApprovalID: "appr-1",
		OrgID:      tenantID.String(),
		RequestID:  "req-1",
		ActionType: "workorder.create",
		RiskLevel:  "high",
	})
	if err != nil {
		t.Fatalf("HandleCallback: %v", err)
	}
	if nx.getCalls != 1 {
		t.Fatalf("nexus get calls = %d, want 1", nx.getCalls)
	}
	if repo.createCalls != 1 {
		t.Fatalf("create calls = %d, want 1", repo.createCalls)
	}
	row := repo.rows["req-1"]
	if row.Origin != "agent" {
		t.Errorf("origin = %q, want agent", row.Origin)
	}
	if row.TenantID != tenantID {
		t.Errorf("tenant mismatch: %s", row.TenantID)
	}
	if row.ApprovalID != "appr-1" || row.Status != nexusclient.StatusPendingApproval {
		t.Errorf("approval state not applied: %+v", row)
	}
	if row.BindingHash != "hash-1" || row.RequesterID != "agent:ponti-ops" {
		t.Errorf("nexus details not hydrated: %+v", row)
	}
}

func TestHandleCallback_PendingIsIdempotent(t *testing.T) {
	t.Parallel()
	tenantID := uuid.New()
	repo := newStubRepo()
	repo.rows["req-1"] = governance.RequestRecord{
		TenantID:       tenantID,
		NexusRequestID: "req-1",
		ActionType:     "workorder.create",
		Origin:         "agent",
		Status:         nexusclient.StatusPendingApproval,
		ApprovalID:     "appr-1",
	}
	svc := governance.NewService(repo, &stubNexus{}, governance.Config{CallbackToken: "tok"}, nil)

	err := svc.HandleCallback(context.Background(), governance.CallbackEvent{
		Event:      governance.EventApprovalPending,
		ApprovalID: "appr-1",
		OrgID:      tenantID.String(),
		RequestID:  "req-1",
	})
	if err != nil {
		t.Fatalf("HandleCallback: %v", err)
	}
	if repo.updateCalls != 0 || repo.createCalls != 0 {
		t.Fatalf("expected no effect on replay: updates=%d creates=%d", repo.updateCalls, repo.createCalls)
	}
}

func TestHandleCallback_PendingDoesNotDowngradeTerminalRow(t *testing.T) {
	t.Parallel()
	terminal := []string{
		nexusclient.StatusApproved,
		nexusclient.StatusRejected,
		nexusclient.StatusExpired,
		nexusclient.StatusExecuted,
		nexusclient.StatusFailed,
		nexusclient.StatusCancelled,
	}
	for _, status := range terminal {
		t.Run(status, func(t *testing.T) {
			t.Parallel()
			tenantID := uuid.New()
			repo := newStubRepo()
			repo.rows["req-1"] = governance.RequestRecord{
				TenantID:       tenantID,
				NexusRequestID: "req-1",
				ActionType:     "workorder.create",
				Origin:         "agent",
				Status:         status,
				ApprovalID:     "appr-1",
				DecidedBy:      "user:pablo",
			}
			svc := governance.NewService(repo, &stubNexus{}, governance.Config{CallbackToken: "tok"}, nil)

			// Callback pending tardío/replayed con otro approval_id: el guard de
			// idempotencia no aplica y solo el guard de estado terminal lo frena.
			err := svc.HandleCallback(context.Background(), governance.CallbackEvent{
				Event:      governance.EventApprovalPending,
				ApprovalID: "appr-2",
				OrgID:      tenantID.String(),
				RequestID:  "req-1",
			})
			if err != nil {
				t.Fatalf("expected ack without effect, got %v", err)
			}
			if repo.updateCalls != 0 {
				t.Fatalf("terminal row must not be updated, got %d updates", repo.updateCalls)
			}
			if row := repo.rows["req-1"]; row.Status != status {
				t.Fatalf("status downgraded to %q, want %q", row.Status, status)
			}
		})
	}
}

func TestHandleCallback_ResolvedStatusMappingAndExecutorHook(t *testing.T) {
	t.Parallel()
	cases := []struct {
		decision      string
		wantStatus    string
		wantExecCalls int
	}{
		{"approved", nexusclient.StatusApproved, 1},
		{"rejected", nexusclient.StatusRejected, 0},
		{"expired", nexusclient.StatusExpired, 0},
	}
	for _, tc := range cases {
		t.Run(tc.decision, func(t *testing.T) {
			t.Parallel()
			tenantID := uuid.New()
			repo := newStubRepo()
			repo.rows["req-1"] = governance.RequestRecord{
				TenantID:       tenantID,
				NexusRequestID: "req-1",
				ActionType:     "workorder.create",
				Origin:         "agent",
				Status:         nexusclient.StatusPendingApproval,
				ApprovalID:     "appr-1",
			}
			exec := &stubExecutor{}
			svc := governance.NewService(repo, &stubNexus{}, governance.Config{CallbackToken: "tok"}, exec)

			err := svc.HandleCallback(context.Background(), governance.CallbackEvent{
				Event:      governance.EventApprovalResolved,
				ApprovalID: "appr-1",
				OrgID:      tenantID.String(),
				RequestID:  "req-1",
				Decision:   tc.decision,
				DecidedBy:  "user:pablo",
			})
			if err != nil {
				t.Fatalf("HandleCallback: %v", err)
			}
			row := repo.rows["req-1"]
			if row.Status != tc.wantStatus {
				t.Errorf("status = %q, want %q", row.Status, tc.wantStatus)
			}
			if row.DecidedBy != "user:pablo" {
				t.Errorf("decided_by = %q", row.DecidedBy)
			}
			if exec.calls != tc.wantExecCalls {
				t.Errorf("executor calls = %d, want %d", exec.calls, tc.wantExecCalls)
			}
		})
	}
}

func TestHandleCallback_ResolvedIsIdempotent(t *testing.T) {
	t.Parallel()
	tenantID := uuid.New()
	repo := newStubRepo()
	repo.rows["req-1"] = governance.RequestRecord{
		TenantID:       tenantID,
		NexusRequestID: "req-1",
		ActionType:     "workorder.create",
		Origin:         "agent",
		Status:         nexusclient.StatusApproved,
		ApprovalID:     "appr-1",
		DecidedBy:      "user:pablo",
	}
	exec := &stubExecutor{}
	svc := governance.NewService(repo, &stubNexus{}, governance.Config{CallbackToken: "tok"}, exec)

	err := svc.HandleCallback(context.Background(), governance.CallbackEvent{
		Event:      governance.EventApprovalResolved,
		ApprovalID: "appr-1",
		OrgID:      tenantID.String(),
		RequestID:  "req-1",
		Decision:   "approved",
		DecidedBy:  "user:pablo",
	})
	if err != nil {
		t.Fatalf("HandleCallback: %v", err)
	}
	if repo.updateCalls != 0 {
		t.Fatalf("expected no update on replay, got %d", repo.updateCalls)
	}
	if exec.calls != 0 {
		t.Fatalf("executor must not run on replay, got %d calls", exec.calls)
	}
}

func TestHandleCallback_UnknownDecisionFails(t *testing.T) {
	t.Parallel()
	tenantID := uuid.New()
	svc := governance.NewService(newStubRepo(), &stubNexus{}, governance.Config{CallbackToken: "tok"}, nil)

	err := svc.HandleCallback(context.Background(), governance.CallbackEvent{
		Event:     governance.EventApprovalResolved,
		OrgID:     tenantID.String(),
		RequestID: "req-1",
		Decision:  "maybe",
	})
	if err == nil {
		t.Fatal("expected error for unknown decision")
	}
}

func TestHandleCallback_InvalidOrgIsAckedWithoutEffect(t *testing.T) {
	t.Parallel()
	repo := newStubRepo()
	svc := governance.NewService(repo, &stubNexus{}, governance.Config{CallbackToken: "tok"}, nil)

	err := svc.HandleCallback(context.Background(), governance.CallbackEvent{
		Event:     governance.EventApprovalPending,
		OrgID:     "not-a-uuid",
		RequestID: "req-1",
	})
	if err != nil {
		t.Fatalf("expected ack without effect, got %v", err)
	}
	if repo.createCalls != 0 || repo.updateCalls != 0 {
		t.Fatalf("expected no persistence: creates=%d updates=%d", repo.createCalls, repo.updateCalls)
	}
}

func TestListPending_DropsForeignAndEmptyOrgItems(t *testing.T) {
	t.Parallel()
	tenantID := uuid.New()
	foreignOrg := uuid.New().String()
	f := &fakeNexus{
		requests: []nexusclient.Request{
			{ID: "req-own", OrgID: tenantID.String(), ActionType: "workorder.create", Status: nexusclient.StatusPendingApproval, RiskLevel: "high"},
			{ID: "req-foreign", OrgID: foreignOrg, ActionType: "workorder.create", Status: nexusclient.StatusPendingApproval, RiskLevel: "high"},
			{ID: "req-empty", OrgID: "", ActionType: "workorder.create", Status: nexusclient.StatusPendingApproval, RiskLevel: "high"},
		},
		approvals: []nexusclient.Approval{
			{ID: "appr-own", OrgID: tenantID.String(), RequestID: "req-own", ExpiresAt: "2026-06-10T13:00:00Z"},
			{ID: "appr-foreign", OrgID: foreignOrg, RequestID: "req-foreign"},
		},
	}
	svc := governance.NewService(newStubRepo(), f.client(t), governance.Config{CallbackToken: "tok"}, nil)

	items, err := svc.ListApprovals(context.Background(), tenantID, "pending", 50)
	if err != nil {
		t.Fatalf("ListApprovals: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("items = %d, want 1 (solo la org propia): %+v", len(items), items)
	}
	if items[0].RequestID != "req-own" || items[0].ApprovalID != "appr-own" {
		t.Errorf("unexpected item: %+v", items[0])
	}
	// Finding 5: el limit del inbox se propaga a ambos listados de Nexus.
	if len(f.requestQueries) != 1 || f.requestQueries[0] != "status=pending_approval&limit=50" {
		t.Errorf("requests query = %v", f.requestQueries)
	}
	if len(f.approvalQueries) != 1 || f.approvalQueries[0] != "limit=50" {
		t.Errorf("approvals query = %v", f.approvalQueries)
	}
}

func TestSummary_CountsOnlyOwnOrgPending(t *testing.T) {
	t.Parallel()
	tenantID := uuid.New()
	foreignOrg := uuid.New().String()
	expiringSoon := time.Now().UTC().Add(5 * time.Minute).Format(time.RFC3339)
	f := &fakeNexus{
		requests: []nexusclient.Request{
			{ID: "req-own", OrgID: tenantID.String(), ActionType: "workorder.create", Status: nexusclient.StatusPendingApproval},
			{ID: "req-foreign", OrgID: foreignOrg, ActionType: "workorder.create", Status: nexusclient.StatusPendingApproval},
			{ID: "req-empty", OrgID: "", ActionType: "workorder.create", Status: nexusclient.StatusPendingApproval},
		},
		approvals: []nexusclient.Approval{
			{ID: "appr-own", OrgID: tenantID.String(), RequestID: "req-own", ExpiresAt: expiringSoon},
		},
	}
	svc := governance.NewService(newStubRepo(), f.client(t), governance.Config{CallbackToken: "tok"}, nil)

	summary, err := svc.Summary(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("Summary: %v", err)
	}
	if summary.PendingCount != 1 || summary.ExpiringSoonCount != 1 {
		t.Errorf("summary = %+v, want pending=1 expiring=1", summary)
	}
	if len(f.approvalQueries) != 1 || f.approvalQueries[0] != "limit=200" {
		t.Errorf("approvals query = %v, want limit=200", f.approvalQueries)
	}
}

func TestGetApproval_ForeignOrEmptyOrgReturnsNotFound(t *testing.T) {
	t.Parallel()
	tenantID := uuid.New()
	f := &fakeNexus{
		requests: []nexusclient.Request{
			{ID: "req-foreign", OrgID: uuid.New().String(), ActionType: "workorder.create", Status: nexusclient.StatusPendingApproval},
			{ID: "req-empty", OrgID: "", ActionType: "workorder.create", Status: nexusclient.StatusPendingApproval},
		},
	}
	svc := governance.NewService(newStubRepo(), f.client(t), governance.Config{CallbackToken: "tok"}, nil)

	for _, requestID := range []string{"req-foreign", "req-empty"} {
		if _, err := svc.GetApproval(context.Background(), tenantID, requestID); !domainerr.IsNotFound(err) {
			t.Errorf("GetApproval(%s) err = %v, want not found", requestID, err)
		}
	}
}

func TestEvidence_ForeignOrEmptyOrgPackNotFoundAndNotCached(t *testing.T) {
	t.Parallel()
	tenantID := uuid.New()
	repo := newStubRepo()
	f := &fakeNexus{
		evidence: map[string]string{
			"req-own":     fmt.Sprintf(`{"request":{"id":"req-own","org_id":%q},"signature":"sig","signature_key_id":"key-1"}`, tenantID.String()),
			"req-foreign": fmt.Sprintf(`{"request":{"id":"req-foreign","org_id":%q},"signature":"sig","signature_key_id":"key-1"}`, uuid.New().String()),
			"req-empty":   `{"request":{"id":"req-empty"},"signature":"sig","signature_key_id":"key-1"}`,
		},
	}
	svc := governance.NewService(repo, f.client(t), governance.Config{CallbackToken: "tok"}, nil)

	for _, requestID := range []string{"req-foreign", "req-empty"} {
		if _, err := svc.Evidence(context.Background(), tenantID, requestID); !domainerr.IsNotFound(err) {
			t.Errorf("Evidence(%s) err = %v, want not found", requestID, err)
		}
	}
	if repo.saveEvidenceCalls != 0 {
		t.Fatalf("foreign packs must never be cached, got %d saves", repo.saveEvidenceCalls)
	}

	// Control: el pack de la org propia sí se sirve y se cachea.
	if _, err := svc.Evidence(context.Background(), tenantID, "req-own"); err != nil {
		t.Fatalf("Evidence(req-own): %v", err)
	}
	if repo.saveEvidenceCalls != 1 {
		t.Fatalf("own pack should be cached once, got %d saves", repo.saveEvidenceCalls)
	}
}

func TestDecide_ForeignOrEmptyOrgApprovalNotFound(t *testing.T) {
	t.Parallel()
	tenantID := uuid.New()
	f := &fakeNexus{
		approvals: []nexusclient.Approval{
			{ID: "appr-foreign", OrgID: uuid.New().String(), RequestID: "req-1"},
			{ID: "appr-empty", OrgID: "", RequestID: "req-1"},
		},
	}
	svc := governance.NewService(newStubRepo(), f.client(t), governance.Config{CallbackToken: "tok"}, nil)

	err := svc.Decide(context.Background(), tenantID, "req-1", "pablo", "approve", "")
	if !domainerr.IsNotFound(err) {
		t.Fatalf("Decide err = %v, want not found", err)
	}
	if len(f.decidePaths) != 0 {
		t.Fatalf("must not decide on foreign approvals, got %v", f.decidePaths)
	}
}

func TestDecide_PersistsOnlyApprovalIDOn200(t *testing.T) {
	t.Parallel()
	tenantID := uuid.New()
	repo := newStubRepo()
	repo.rows["req-1"] = governance.RequestRecord{
		TenantID:       tenantID,
		NexusRequestID: "req-1",
		ActionType:     "workorder.create",
		Origin:         "agent",
		Status:         nexusclient.StatusPendingApproval,
		ApprovalID:     "appr-1",
	}
	nx := &stubNexus{}
	svc := governance.NewService(repo, nx, governance.Config{CallbackToken: "tok"}, nil)

	if err := svc.Decide(context.Background(), tenantID, "req-1", "pablo", "approve", "lgtm"); err != nil {
		t.Fatalf("Decide: %v", err)
	}
	if nx.approveCalls != 1 {
		t.Fatalf("approve calls = %d, want 1", nx.approveCalls)
	}
	// El 200 de Nexus puede ser una aprobación parcial multi-approver: el status
	// terminal lo escribe únicamente el callback approval_resolved.
	if _, ok := repo.lastUpdates["status"]; ok {
		t.Errorf("Decide must not write status, got updates %v", repo.lastUpdates)
	}
	if _, ok := repo.lastUpdates["decided_by"]; ok {
		t.Errorf("Decide must not write decided_by, got updates %v", repo.lastUpdates)
	}
	if repo.lastUpdates["approval_id"] != "appr-1" {
		t.Errorf("approval_id not persisted: %v", repo.lastUpdates)
	}
	if row := repo.rows["req-1"]; row.Status != nexusclient.StatusPendingApproval {
		t.Errorf("mirror status = %q, want pending_approval", row.Status)
	}
}

func TestDecide_PartialApprovalThenResolvedCallbackRunsExecutorOnce(t *testing.T) {
	t.Parallel()
	tenantID := uuid.New()
	repo := newStubRepo()
	repo.rows["req-1"] = governance.RequestRecord{
		TenantID:       tenantID,
		NexusRequestID: "req-1",
		ActionType:     "workorder.create",
		Origin:         "agent",
		Status:         nexusclient.StatusPendingApproval,
		ApprovalID:     "appr-1",
	}
	exec := &stubExecutor{}
	svc := governance.NewService(repo, &stubNexus{}, governance.Config{CallbackToken: "tok"}, exec)

	// Aprobación parcial: Nexus responde 200 pero la request sigue pending.
	if err := svc.Decide(context.Background(), tenantID, "req-1", "pablo", "approve", ""); err != nil {
		t.Fatalf("Decide: %v", err)
	}
	if row := repo.rows["req-1"]; row.Status != nexusclient.StatusPendingApproval {
		t.Fatalf("mirror must stay pending after partial approval, got %q", row.Status)
	}
	if exec.calls != 0 {
		t.Fatalf("executor must not run on Decide, got %d calls", exec.calls)
	}

	// El callback approval_resolved genuino (mismo decided_by que el Decide)
	// es el único writer del status terminal y dispara el executor.
	resolved := governance.CallbackEvent{
		Event:      governance.EventApprovalResolved,
		ApprovalID: "appr-1",
		OrgID:      tenantID.String(),
		RequestID:  "req-1",
		Decision:   "approved",
		DecidedBy:  "user:pablo",
	}
	if err := svc.HandleCallback(context.Background(), resolved); err != nil {
		t.Fatalf("HandleCallback resolved: %v", err)
	}
	row := repo.rows["req-1"]
	if row.Status != nexusclient.StatusApproved || row.DecidedBy != "user:pablo" {
		t.Fatalf("resolved not applied: %+v", row)
	}
	if exec.calls != 1 {
		t.Fatalf("executor calls = %d, want 1", exec.calls)
	}

	// Retry del mismo callback: idempotente, sin re-ejecución.
	if err := svc.HandleCallback(context.Background(), resolved); err != nil {
		t.Fatalf("HandleCallback retry: %v", err)
	}
	if exec.calls != 1 {
		t.Fatalf("executor must run exactly once, got %d calls", exec.calls)
	}
}

func TestDecide_LooksUpApprovalByRequestIDQuery(t *testing.T) {
	t.Parallel()
	tenantID := uuid.New()
	f := &fakeNexus{
		honorFilters: true,
		approvals: []nexusclient.Approval{
			{ID: "appr-1", OrgID: tenantID.String(), RequestID: "req-1"},
			{ID: "appr-2", OrgID: tenantID.String(), RequestID: "req-2"},
		},
	}
	svc := governance.NewService(newStubRepo(), f.client(t), governance.Config{CallbackToken: "tok"}, nil)

	if err := svc.Decide(context.Background(), tenantID, "req-1", "pablo", "approve", ""); err != nil {
		t.Fatalf("Decide: %v", err)
	}
	if len(f.approvalQueries) != 1 || f.approvalQueries[0] != "request_id=req-1" {
		t.Errorf("approvals query = %v, want request_id=req-1", f.approvalQueries)
	}
	if len(f.decidePaths) != 1 || f.decidePaths[0] != "/v1/approvals/appr-1/approve" {
		t.Errorf("decide paths = %v, want approve de appr-1", f.decidePaths)
	}
}

func TestDecide_ClientSideFilterFallbackWhenNexusIgnoresParams(t *testing.T) {
	t.Parallel()
	tenantID := uuid.New()
	// Nexus viejo: ignora request_id/limit y devuelve la lista completa, con
	// approvals de otras orgs y de otros requests por delante.
	f := &fakeNexus{
		honorFilters: false,
		approvals: []nexusclient.Approval{
			{ID: "appr-foreign", OrgID: uuid.New().String(), RequestID: "req-1"},
			{ID: "appr-other", OrgID: tenantID.String(), RequestID: "req-9"},
			{ID: "appr-1", OrgID: tenantID.String(), RequestID: "req-1"},
		},
	}
	svc := governance.NewService(newStubRepo(), f.client(t), governance.Config{CallbackToken: "tok"}, nil)

	if err := svc.Decide(context.Background(), tenantID, "req-1", "pablo", "approve", ""); err != nil {
		t.Fatalf("Decide: %v", err)
	}
	if len(f.decidePaths) != 1 || f.decidePaths[0] != "/v1/approvals/appr-1/approve" {
		t.Errorf("decide paths = %v, want approve de appr-1", f.decidePaths)
	}
}
