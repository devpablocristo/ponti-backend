package governance_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"testing"

	"github.com/google/uuid"

	"github.com/devpablocristo/ponti-backend/internal/governance"
	nexusclient "github.com/devpablocristo/ponti-backend/internal/nexus"
)

type stubExecutorNexus struct {
	mu                sync.Mutex
	reportCalls       int
	lastReport        nexusclient.ReportResultBody
	attestCalls       int
	lastAttestation   map[string]any
	getEvidenceCalls  int
	evidencePack      []byte
	evidenceHTTPState int
}

func (s *stubExecutorNexus) ReportResult(_ context.Context, _ string, body nexusclient.ReportResultBody, _ ...nexusclient.RequestOption) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reportCalls++
	s.lastReport = body
	return http.StatusNoContent, nil
}

func (s *stubExecutorNexus) Attest(_ context.Context, _ string, body any, _ ...nexusclient.RequestOption) (int, []byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.attestCalls++
	if m, ok := body.(map[string]any); ok {
		s.lastAttestation = m
	}
	return http.StatusCreated, nil, nil
}

func (s *stubExecutorNexus) GetEvidence(_ context.Context, _ string, _ ...nexusclient.RequestOption) ([]byte, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.getEvidenceCalls++
	status := s.evidenceHTTPState
	if status == 0 {
		status = http.StatusOK
	}
	return s.evidencePack, status, nil
}

func (s *stubExecutorNexus) snapshot() (int, nexusclient.ReportResultBody, int, map[string]any, int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.reportCalls, s.lastReport, s.attestCalls, s.lastAttestation, s.getEvidenceCalls
}

type stubDispatcher struct {
	calls       int
	lastTenant  uuid.UUID
	lastAction  string
	lastRequest string
	lastParams  map[string]any
	lastActor   string
	result      map[string]any
	err         error
}

func (s *stubDispatcher) DispatchApproved(_ context.Context, tenantID uuid.UUID, actionType, nexusRequestID string, params map[string]any, actor string) (map[string]any, error) {
	s.calls++
	s.lastTenant = tenantID
	s.lastAction = actionType
	s.lastRequest = nexusRequestID
	s.lastParams = params
	s.lastActor = actor
	return s.result, s.err
}

func evidencePackForTenant(tenantID uuid.UUID) []byte {
	pack, _ := json.Marshal(map[string]any{
		"request":   map[string]any{"org_id": tenantID.String()},
		"signature": "sig-1",
		"key_id":    "key-1",
	})
	return pack
}

func approvedRow(tenantID uuid.UUID, actionType string, params map[string]any) governance.RequestRecord {
	raw, _ := json.Marshal(params)
	return governance.RequestRecord{
		TenantID:       tenantID,
		NexusRequestID: "req-1",
		ActionType:     actionType,
		Status:         nexusclient.StatusApproved,
		ParamsJSON:     raw,
		DecidedBy:      "user:approver",
	}
}

func TestApprovedExecutorDispatchesAndReportsResult(t *testing.T) {
	t.Parallel()
	tenantID := uuid.New()
	repo := newStubRepo()
	row := approvedRow(tenantID, nexusclient.ActionTypeInsightResolve, map[string]any{"insight_id": "ins-1"})
	repo.rows[row.NexusRequestID] = row
	nexus := &stubExecutorNexus{evidencePack: evidencePackForTenant(tenantID)}
	dispatcher := &stubDispatcher{result: map[string]any{"draft_id": "d-1"}}
	executor := governance.NewApprovedExecutor(repo, nexus, governance.ExecutorConfig{
		GovernedWritesEnabled: true,
		AttestationHMACSecret: "attest-secret",
	})
	executor.SetDispatcher(dispatcher)

	if err := executor.ExecuteApproved(context.Background(), row); err != nil {
		t.Fatalf("execute approved: %v", err)
	}
	executor.Wait()

	if dispatcher.calls != 1 || dispatcher.lastAction != nexusclient.ActionTypeInsightResolve || dispatcher.lastRequest != "req-1" {
		t.Fatalf("expected dispatch of approved action, got %+v", dispatcher)
	}
	if dispatcher.lastParams["insight_id"] != "ins-1" || dispatcher.lastActor != "user:approver" {
		t.Fatalf("expected params/actor from local row, got %+v", dispatcher)
	}
	if repo.lastUpdates["executed_at"] == nil {
		t.Fatalf("expected executed_at recorded, got %#v", repo.lastUpdates)
	}
	if repo.lastUpdates["result_json"] == nil {
		t.Fatalf("expected result_json recorded, got %#v", repo.lastUpdates)
	}
	reportCalls, lastReport, attestCalls, attestation, evidenceCalls := nexus.snapshot()
	if reportCalls != 1 || !lastReport.Success {
		t.Fatalf("expected successful report result, got calls=%d body=%+v", reportCalls, lastReport)
	}
	if attestCalls != 1 || attestation["attester"] != "ponti-backend" || attestation["signature"] == "" {
		t.Fatalf("expected hmac attestation by ponti-backend, got calls=%d body=%#v", attestCalls, attestation)
	}
	if evidenceCalls != 1 || repo.saveEvidenceCalls != 1 {
		t.Fatalf("expected evidence fetched and cached, got fetch=%d save=%d", evidenceCalls, repo.saveEvidenceCalls)
	}
}

func TestApprovedExecutorSkipsAttestationWithoutSecret(t *testing.T) {
	t.Parallel()
	tenantID := uuid.New()
	repo := newStubRepo()
	row := approvedRow(tenantID, nexusclient.ActionTypeStockCountApply, map[string]any{"project_id": 10})
	repo.rows[row.NexusRequestID] = row
	nexus := &stubExecutorNexus{evidencePack: evidencePackForTenant(tenantID)}
	executor := governance.NewApprovedExecutor(repo, nexus, governance.ExecutorConfig{GovernedWritesEnabled: true})
	executor.SetDispatcher(&stubDispatcher{result: map[string]any{"draft_id": "d-1"}})

	if err := executor.ExecuteApproved(context.Background(), row); err != nil {
		t.Fatalf("execute approved: %v", err)
	}
	executor.Wait()

	reportCalls, _, attestCalls, _, _ := nexus.snapshot()
	if reportCalls != 1 || attestCalls != 0 {
		t.Fatalf("expected report without attestation, got report=%d attest=%d", reportCalls, attestCalls)
	}
}

func TestApprovedExecutorNoopWhenWritesDisabled(t *testing.T) {
	t.Parallel()
	repo := newStubRepo()
	row := approvedRow(uuid.New(), nexusclient.ActionTypeInsightResolve, nil)
	nexus := &stubExecutorNexus{}
	dispatcher := &stubDispatcher{}
	executor := governance.NewApprovedExecutor(repo, nexus, governance.ExecutorConfig{GovernedWritesEnabled: false})
	executor.SetDispatcher(dispatcher)

	if err := executor.ExecuteApproved(context.Background(), row); err != nil {
		t.Fatalf("execute approved: %v", err)
	}
	executor.Wait()

	reportCalls, _, _, _, _ := nexus.snapshot()
	if dispatcher.calls != 0 || reportCalls != 0 || repo.updateCalls != 0 {
		t.Fatalf("expected noop with flag off, got dispatcher=%d report=%d updates=%d", dispatcher.calls, reportCalls, repo.updateCalls)
	}
}

func TestApprovedExecutorMarksUnknownActionTypesWithoutCrashing(t *testing.T) {
	t.Parallel()
	repo := newStubRepo()
	row := approvedRow(uuid.New(), "ponti.unknown.action", nil)
	repo.rows[row.NexusRequestID] = row
	nexus := &stubExecutorNexus{}
	dispatcher := &stubDispatcher{}
	executor := governance.NewApprovedExecutor(repo, nexus, governance.ExecutorConfig{GovernedWritesEnabled: true})
	executor.SetDispatcher(dispatcher)

	if err := executor.ExecuteApproved(context.Background(), row); err != nil {
		t.Fatalf("execute approved must not fail: %v", err)
	}
	executor.Wait()

	if dispatcher.calls != 0 {
		t.Fatalf("dispatcher must not run for unknown action types, got %d", dispatcher.calls)
	}
	if msg, _ := repo.lastUpdates["error_message"].(string); msg == "" {
		t.Fatalf("expected error_message recorded, got %#v", repo.lastUpdates)
	}
	if _, ok := repo.lastUpdates["executed_at"]; ok {
		t.Fatalf("unknown action types must not mark executed_at, got %#v", repo.lastUpdates)
	}
	reportCalls, _, _, _, _ := nexus.snapshot()
	if reportCalls != 0 {
		t.Fatalf("unknown action types must not report to nexus, got %d", reportCalls)
	}
}

func TestApprovedExecutorReportsFailureWhenDispatchFails(t *testing.T) {
	t.Parallel()
	tenantID := uuid.New()
	repo := newStubRepo()
	row := approvedRow(tenantID, nexusclient.ActionTypeInsightResolve, map[string]any{"insight_id": "ins-1"})
	repo.rows[row.NexusRequestID] = row
	nexus := &stubExecutorNexus{evidencePack: evidencePackForTenant(tenantID)}
	executor := governance.NewApprovedExecutor(repo, nexus, governance.ExecutorConfig{GovernedWritesEnabled: true})
	executor.SetDispatcher(&stubDispatcher{err: errors.New("insight not found")})

	if err := executor.ExecuteApproved(context.Background(), row); err != nil {
		t.Fatalf("execute approved must not propagate dispatch failures: %v", err)
	}
	executor.Wait()

	if msg, _ := repo.lastUpdates["error_message"].(string); msg != "insight not found" {
		t.Fatalf("expected error_message recorded, got %#v", repo.lastUpdates)
	}
	_, lastReport, _, _, _ := nexus.snapshot()
	if lastReport.Success || lastReport.ErrorMessage != "insight not found" {
		t.Fatalf("expected failure report, got %+v", lastReport)
	}
}

func TestApprovedExecutorSkipsAlreadyExecutedRows(t *testing.T) {
	t.Parallel()
	repo := newStubRepo()
	row := approvedRow(uuid.New(), nexusclient.ActionTypeInsightResolve, nil)
	executedAt := row.CreatedAt
	row.ExecutedAt = &executedAt
	dispatcher := &stubDispatcher{}
	executor := governance.NewApprovedExecutor(repo, &stubExecutorNexus{}, governance.ExecutorConfig{GovernedWritesEnabled: true})
	executor.SetDispatcher(dispatcher)

	if err := executor.ExecuteApproved(context.Background(), row); err != nil {
		t.Fatalf("execute approved: %v", err)
	}
	executor.Wait()
	if dispatcher.calls != 0 {
		t.Fatalf("already executed rows must not re-dispatch, got %d", dispatcher.calls)
	}
}

// TestApprovalResolvedCallbackRunsApprovedExecutor cubre el flujo completo:
// callback approval_resolved(approved) → Service → ApprovedExecutor →
// dispatcher + ReportResult/Attest/GetEvidence contra el Nexus falso.
func TestApprovalResolvedCallbackRunsApprovedExecutor(t *testing.T) {
	t.Parallel()
	tenantID := uuid.New()
	repo := newStubRepo()
	row := approvedRow(tenantID, nexusclient.ActionTypeInsightResolve, map[string]any{"insight_id": "ins-1"})
	row.Status = nexusclient.StatusPendingApproval
	row.DecidedBy = ""
	repo.rows[row.NexusRequestID] = row
	nexus := &stubExecutorNexus{evidencePack: evidencePackForTenant(tenantID)}
	dispatcher := &stubDispatcher{result: map[string]any{"draft_id": "d-1"}}
	executor := governance.NewApprovedExecutor(repo, nexus, governance.ExecutorConfig{
		GovernedWritesEnabled: true,
		AttestationHMACSecret: "attest-secret",
	})
	executor.SetDispatcher(dispatcher)
	svc := governance.NewService(repo, &stubNexus{}, governance.Config{CallbackToken: "token"}, executor)

	err := svc.HandleCallback(context.Background(), governance.CallbackEvent{
		Event:      governance.EventApprovalResolved,
		OrgID:      tenantID.String(),
		RequestID:  row.NexusRequestID,
		ApprovalID: "appr-1",
		Decision:   nexusclient.StatusApproved,
		DecidedBy:  "user:approver",
	})
	if err != nil {
		t.Fatalf("handle callback: %v", err)
	}
	executor.Wait()

	if dispatcher.calls != 1 || dispatcher.lastTenant != tenantID || dispatcher.lastActor != "user:approver" {
		t.Fatalf("expected executor dispatch from callback, got %+v", dispatcher)
	}
	reportCalls, lastReport, attestCalls, _, evidenceCalls := nexus.snapshot()
	if reportCalls != 1 || !lastReport.Success || attestCalls != 1 || evidenceCalls != 1 {
		t.Fatalf("expected full nexus loop, got report=%d attest=%d evidence=%d", reportCalls, attestCalls, evidenceCalls)
	}
	if repo.saveEvidenceCalls != 1 {
		t.Fatalf("expected evidence cached locally, got %d", repo.saveEvidenceCalls)
	}
}
