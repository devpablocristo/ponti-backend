package nexus

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSubmitWithActionBinding(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/requests" {
			t.Errorf("expected /v1/requests, got %s", r.URL.Path)
		}
		if r.Header.Get("X-API-Key") != "test-key" {
			t.Errorf("expected X-API-Key=test-key, got %s", r.Header.Get("X-API-Key"))
		}
		if r.Header.Get("X-Org-ID") != "tenant-1" {
			t.Errorf("expected X-Org-ID=tenant-1, got %s", r.Header.Get("X-Org-ID"))
		}
		if r.Header.Get("Idempotency-Key") != "idem-123" {
			t.Errorf("expected Idempotency-Key=idem-123, got %s", r.Header.Get("Idempotency-Key"))
		}

		var body SubmitRequestBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body.ActionType != "workorder.create" {
			t.Errorf("expected action_type=workorder.create, got %s", body.ActionType)
		}
		if body.ActionBinding == nil {
			t.Fatal("expected action_binding in body")
		}
		if body.ActionBinding.SchemaVersion != ToolIntentSchemaVersion {
			t.Errorf("expected schema_version=%s, got %s", ToolIntentSchemaVersion, body.ActionBinding.SchemaVersion)
		}
		if body.ActionBinding.PayloadHash != "sha256:abc" {
			t.Errorf("expected payload_hash=sha256:abc, got %s", body.ActionBinding.PayloadHash)
		}
		if body.ActionBinding.ToolInvocationID != "inv-1" {
			t.Errorf("expected tool_invocation_id=inv-1, got %s", body.ActionBinding.ToolInvocationID)
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(SubmitResponse{
			RequestID:   "req-abc",
			Decision:    DecisionRequireApproval,
			RiskLevel:   "high",
			Status:      StatusPendingApproval,
			BindingHash: "hash-1",
			Approval:    &ApprovalRef{ID: "appr-1", ExpiresAt: "2026-06-10T13:00:00Z"},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "test-key", 0)
	resp, err := c.Submit(context.Background(), SubmitRequestBody{
		RequesterType: "service",
		RequesterID:   "ponti-backend",
		ActionType:    "workorder.create",
		TargetSystem:  "ponti",
		ActionBinding: &ToolIntent{
			SchemaVersion:    ToolIntentSchemaVersion,
			OrgID:            "tenant-1",
			ActorID:          "user:abc",
			ActorType:        "user",
			ProductSurface:   "ponti",
			RunID:            "run-1",
			ToolInvocationID: "inv-1",
			ConnectorID:      "ponti-core",
			CapabilityID:     "workorder.create",
			Operation:        "create",
			TargetSystem:     "ponti",
			TargetResource:   "work_orders",
			PayloadHash:      "sha256:abc",
			IdempotencyKey:   "idem-123",
		},
	}, WithTenantID("tenant-1"), WithIdempotencyKey("idem-123"))
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if resp.RequestID != "req-abc" {
		t.Errorf("expected request_id=req-abc, got %s", resp.RequestID)
	}
	if resp.BindingHash != "hash-1" {
		t.Errorf("expected binding_hash=hash-1, got %s", resp.BindingHash)
	}
	if resp.Approval == nil || resp.Approval.ID != "appr-1" {
		t.Errorf("expected approval id=appr-1, got %+v", resp.Approval)
	}
}

func TestApproveOnBehalfOf(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/approvals/appr-1/approve" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("X-On-Behalf-Of") != "user:pablo" {
			t.Errorf("expected X-On-Behalf-Of=user:pablo, got %s", r.Header.Get("X-On-Behalf-Of"))
		}
		if r.Header.Get("X-Org-ID") != "tenant-1" {
			t.Errorf("expected X-Org-ID=tenant-1, got %s", r.Header.Get("X-Org-ID"))
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["decided_by"] != "user:pablo" {
			t.Errorf("expected decided_by=user:pablo, got %v", body["decided_by"])
		}
		if body["note"] != "lgtm" {
			t.Errorf("expected note=lgtm, got %v", body["note"])
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"approved"}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "test-key", 0)
	st, raw, err := c.Approve(context.Background(), "appr-1", "user:pablo", "lgtm", WithTenantID("tenant-1"))
	if err != nil {
		t.Fatalf("approve: %v", err)
	}
	if st != http.StatusOK {
		t.Errorf("expected 200, got %d", st)
	}
	if len(raw) == 0 {
		t.Error("expected non-empty response body")
	}
}

func TestRejectPropagatesConflict(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"code":"CONFLICT","message":"approver cannot approve their own request"}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "test-key", 0)
	st, raw, err := c.Reject(context.Background(), "appr-2", "user:pablo", "")
	if err != nil {
		t.Fatalf("reject: %v", err)
	}
	if st != http.StatusConflict {
		t.Errorf("expected 409, got %d", st)
	}
	if got := ParseErrorBody(raw); got != "approver cannot approve their own request" {
		t.Errorf("unexpected error message: %s", got)
	}
}

func TestListRequestsDecodesEnvelope(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/requests" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.RawQuery != "status=pending_approval&limit=50" {
			t.Errorf("unexpected query: %s", r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":[{"id":"req-1","action_type":"workorder.create","status":"pending_approval","risk_level":"high"}]}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "test-key", 0)
	out, err := c.ListRequests(context.Background(), "status=pending_approval&limit=50")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(out) != 1 || out[0].ID != "req-1" || out[0].Status != StatusPendingApproval {
		t.Errorf("unexpected items: %+v", out)
	}
}

func TestListPendingApprovalsPassesQuery(t *testing.T) {
	t.Parallel()
	var gotQueries []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/approvals/pending" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		gotQueries = append(gotQueries, r.URL.RawQuery)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":[{"id":"appr-1","request_id":"req-1","status":"pending","expires_at":"2026-06-10T13:00:00Z"}]}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "test-key", 0)
	out, err := c.ListPendingApprovals(context.Background(), "request_id=req-1&limit=50")
	if err != nil {
		t.Fatalf("list pending approvals: %v", err)
	}
	if len(out) != 1 || out[0].ID != "appr-1" || out[0].RequestID != "req-1" {
		t.Errorf("unexpected items: %+v", out)
	}
	// Sin query no debe agregarse "?" al path.
	if _, err := c.ListPendingApprovals(context.Background(), ""); err != nil {
		t.Fatalf("list pending approvals without query: %v", err)
	}
	if len(gotQueries) != 2 || gotQueries[0] != "request_id=req-1&limit=50" || gotQueries[1] != "" {
		t.Errorf("queries = %v, want [request_id=req-1&limit=50 \"\"]", gotQueries)
	}
}

func TestGetEvidenceReturnsRawPack(t *testing.T) {
	t.Parallel()
	pack := `{"request":{"id":"req-1"},"signature":"sig","signature_key_id":"key-1"}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/requests/req-1/evidence" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(pack))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "test-key", 0)
	raw, st, err := c.GetEvidence(context.Background(), "req-1")
	if err != nil {
		t.Fatalf("evidence: %v", err)
	}
	if st != http.StatusOK {
		t.Errorf("expected 200, got %d", st)
	}
	if string(raw) != pack {
		t.Errorf("expected raw pack passthrough, got %s", raw)
	}
}

func TestReportResultConflict(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/requests/req-9/result" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body ReportResultBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body.ErrorMessage != "boom" || body.Success {
			t.Errorf("unexpected body: %+v", body)
		}
		w.WriteHeader(http.StatusConflict)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "test-key", 0)
	st, err := c.ReportResult(context.Background(), "req-9", ReportResultBody{
		Success:      false,
		Result:       map[string]any{"detail": "x"},
		ErrorMessage: "boom",
	})
	if err != nil {
		t.Fatalf("report result: %v", err)
	}
	if st != http.StatusConflict {
		t.Errorf("expected 409, got %d", st)
	}
}
