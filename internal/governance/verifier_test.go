package governance_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	"github.com/google/uuid"

	"github.com/devpablocristo/ponti-backend/internal/governance"
	nexusclient "github.com/devpablocristo/ponti-backend/internal/nexus"
)

// verifierNexusServer es un Nexus falso (httptest + cliente HTTP real) que
// sirve GET /v1/requests/{id} y cuenta llamadas para validar el cache.
type verifierNexusServer struct {
	mu       sync.Mutex
	requests map[string]nexusclient.Request
	getCalls int
}

func (f *verifierNexusServer) client(t *testing.T) (*nexusclient.Client, func()) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || !strings.HasPrefix(r.URL.Path, "/v1/requests/") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		f.mu.Lock()
		f.getCalls++
		req, ok := f.requests[strings.TrimPrefix(r.URL.Path, "/v1/requests/")]
		f.mu.Unlock()
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"code":"not_found","message":"request not found"}`))
			return
		}
		_ = json.NewEncoder(w).Encode(req)
	}))
	return nexusclient.NewClient(srv.URL, "test-key", 0), srv.Close
}

func (f *verifierNexusServer) calls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.getCalls
}

func (f *verifierNexusServer) set(req nexusclient.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.requests == nil {
		f.requests = map[string]nexusclient.Request{}
	}
	f.requests[req.ID] = req
}

func approvedNexusRequest(id string, tenantID uuid.UUID, actionType string) nexusclient.Request {
	return nexusclient.Request{
		ID:         id,
		OrgID:      tenantID.String(),
		ActionType: actionType,
		Status:     nexusclient.StatusApproved,
	}
}

func TestVerifyApprovedAcceptsApprovedRequestAndCachesResult(t *testing.T) {
	t.Parallel()
	tenantID := uuid.New()
	fake := &verifierNexusServer{}
	fake.set(approvedNexusRequest("req-1", tenantID, nexusclient.ActionTypeInsightResolve))
	client, closeFn := fake.client(t)
	defer closeFn()
	verifier := governance.NewVerifier(client)

	if err := verifier.VerifyApproved(context.Background(), tenantID, "req-1", nexusclient.ActionTypeInsightResolve); err != nil {
		t.Fatalf("verify approved: %v", err)
	}
	if err := verifier.VerifyApproved(context.Background(), tenantID, "req-1", nexusclient.ActionTypeInsightResolve); err != nil {
		t.Fatalf("verify approved (cache hit): %v", err)
	}
	if fake.calls() != 1 {
		t.Fatalf("expected single nexus call thanks to cache, got %d", fake.calls())
	}
}

func TestVerifyApprovedAcceptsAllowedStatus(t *testing.T) {
	t.Parallel()
	tenantID := uuid.New()
	fake := &verifierNexusServer{}
	req := approvedNexusRequest("req-1", tenantID, nexusclient.ActionTypeStockCountApply)
	req.Status = nexusclient.StatusAllowed
	fake.set(req)
	client, closeFn := fake.client(t)
	defer closeFn()

	if err := governance.NewVerifier(client).VerifyApproved(context.Background(), tenantID, "req-1", nexusclient.ActionTypeStockCountApply); err != nil {
		t.Fatalf("verify allowed: %v", err)
	}
}

func TestVerifyApprovedRejectsPendingRequestWithoutCaching(t *testing.T) {
	t.Parallel()
	tenantID := uuid.New()
	fake := &verifierNexusServer{}
	req := approvedNexusRequest("req-1", tenantID, nexusclient.ActionTypeInsightResolve)
	req.Status = nexusclient.StatusPendingApproval
	fake.set(req)
	client, closeFn := fake.client(t)
	defer closeFn()
	verifier := governance.NewVerifier(client)

	err := verifier.VerifyApproved(context.Background(), tenantID, "req-1", nexusclient.ActionTypeInsightResolve)
	if !governance.IsNotApproved(err) {
		t.Fatalf("expected NotApprovedError for pending request, got %v", err)
	}

	// Los resultados negativos no se cachean: tras aprobarse en Nexus, la
	// siguiente verificación debe reconsultar y pasar.
	fake.set(approvedNexusRequest("req-1", tenantID, nexusclient.ActionTypeInsightResolve))
	if err := verifier.VerifyApproved(context.Background(), tenantID, "req-1", nexusclient.ActionTypeInsightResolve); err != nil {
		t.Fatalf("verify after approval: %v", err)
	}
	if fake.calls() != 2 {
		t.Fatalf("expected second nexus call (no negative cache), got %d", fake.calls())
	}
}

func TestVerifyApprovedRejectsActionTypeMismatchUnlessLegacyInvoke(t *testing.T) {
	t.Parallel()
	tenantID := uuid.New()
	fake := &verifierNexusServer{}
	fake.set(approvedNexusRequest("req-mismatch", tenantID, nexusclient.ActionTypeStockAdjust))
	fake.set(approvedNexusRequest("req-legacy", tenantID, nexusclient.ActionTypeCapabilityInvoke))
	client, closeFn := fake.client(t)
	defer closeFn()
	verifier := governance.NewVerifier(client)

	err := verifier.VerifyApproved(context.Background(), tenantID, "req-mismatch", nexusclient.ActionTypeInsightResolve)
	if !governance.IsNotApproved(err) {
		t.Fatalf("expected NotApprovedError for action type mismatch, got %v", err)
	}
	// agent.capability.invoke es fallback legacy válido para cualquier tool.
	if err := verifier.VerifyApproved(context.Background(), tenantID, "req-legacy", nexusclient.ActionTypeInsightResolve); err != nil {
		t.Fatalf("expected legacy invoke fallback to pass, got %v", err)
	}
}

func TestVerifyApprovedRejectsForeignOrgRequest(t *testing.T) {
	t.Parallel()
	tenantID := uuid.New()
	fake := &verifierNexusServer{}
	fake.set(approvedNexusRequest("req-1", uuid.New(), nexusclient.ActionTypeInsightResolve))
	client, closeFn := fake.client(t)
	defer closeFn()

	err := governance.NewVerifier(client).VerifyApproved(context.Background(), tenantID, "req-1", nexusclient.ActionTypeInsightResolve)
	if !governance.IsNotApproved(err) {
		t.Fatalf("expected NotApprovedError for foreign org, got %v", err)
	}
}

func TestVerifyApprovedRejectsUnknownRequest(t *testing.T) {
	t.Parallel()
	fake := &verifierNexusServer{}
	client, closeFn := fake.client(t)
	defer closeFn()

	err := governance.NewVerifier(client).VerifyApproved(context.Background(), uuid.New(), "req-missing", nexusclient.ActionTypeInsightResolve)
	if !governance.IsNotApproved(err) {
		t.Fatalf("expected NotApprovedError for unknown request, got %v", err)
	}
}

func TestVerifyApprovedFailsClosedWhenNexusUnreachable(t *testing.T) {
	t.Parallel()
	fake := &verifierNexusServer{}
	client, closeFn := fake.client(t)
	closeFn() // Nexus caído.

	err := governance.NewVerifier(client).VerifyApproved(context.Background(), uuid.New(), "req-1", nexusclient.ActionTypeInsightResolve)
	if err == nil {
		t.Fatal("expected fail-closed error with nexus down")
	}
	if governance.IsNotApproved(err) {
		t.Fatalf("nexus outage must not map to 412, got %v", err)
	}
	if !domainerr.IsKind(err, domainerr.KindUpstreamError) {
		t.Fatalf("expected upstream error (502), got %v", err)
	}
}

func TestVerifyApprovedFailsClosedWithoutNexusClient(t *testing.T) {
	t.Parallel()
	err := governance.NewVerifier(nil).VerifyApproved(context.Background(), uuid.New(), "req-1", nexusclient.ActionTypeInsightResolve)
	if err == nil || governance.IsNotApproved(err) {
		t.Fatalf("expected fail-closed unavailable error, got %v", err)
	}
}

func TestVerifyApprovedRequiresRequestID(t *testing.T) {
	t.Parallel()
	err := governance.NewVerifier(nil).VerifyApproved(context.Background(), uuid.New(), "  ", nexusclient.ActionTypeInsightResolve)
	if !governance.IsNotApproved(err) {
		t.Fatalf("expected NotApprovedError for empty request id, got %v", err)
	}
}
