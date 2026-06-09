package businessinsights_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/devpablocristo/platform/kernels/governance/go/governanceclient"
	"github.com/google/uuid"

	"github.com/devpablocristo/ponti-backend/internal/businessinsights"
)

type stubReview struct {
	calls    int
	response governanceclient.SubmitResponse
	err      error
	lastBody governanceclient.SubmitRequestBody
}

func (s *stubReview) SubmitRequest(_ context.Context, _ string, body governanceclient.SubmitRequestBody) (governanceclient.SubmitResponse, error) {
	s.calls++
	s.lastBody = body
	return s.response, s.err
}

type stubRepo struct {
	upsertCalls  int
	markCalls    int
	shouldNotify bool
	upsertErr    error
	markErr      error
	lastUpsert   businessinsights.CandidateUpsert
	lastMark     [2]string // tenantID, candidateID
}

type stubResolver struct {
	resolveByEntityCalls int
	lastResolveByEntity  [5]string // tenantID, eventType, entityType, entityID, actor
}

func (s *stubRepo) Upsert(_ context.Context, in businessinsights.CandidateUpsert) (businessinsights.CandidateRecord, bool, error) {
	s.upsertCalls++
	s.lastUpsert = in
	if s.upsertErr != nil {
		return businessinsights.CandidateRecord{}, false, s.upsertErr
	}
	return businessinsights.CandidateRecord{
		ID:       "cand-1",
		TenantID: in.TenantID,
	}, s.shouldNotify, nil
}

func (s *stubRepo) MarkNotified(_ context.Context, tenantID, candidateID string, _ time.Time) error {
	s.markCalls++
	s.lastMark = [2]string{tenantID, candidateID}
	return s.markErr
}

func (s *stubResolver) ResolveByEntity(_ context.Context, tenantID, eventType, entityType, entityID, actor string, _ time.Time) (int64, error) {
	s.resolveByEntityCalls++
	s.lastResolveByEntity = [5]string{tenantID, eventType, entityType, entityID, actor}
	return 1, nil
}

func (s *stubResolver) ResolveByID(_ context.Context, _ string, _ string, _ string, _ time.Time) error {
	return nil
}

func (s *stubResolver) ReopenByID(_ context.Context, _ string, _ string, _ string, _ time.Time) error {
	return nil
}

func TestNotifyStockNegative_PolicyMatched_NotifiesOnce(t *testing.T) {
	repo := &stubRepo{shouldNotify: true}
	review := &stubReview{response: governanceclient.SubmitResponse{
		RequestID:      "req-1",
		Decision:       "allow",
		DecisionReason: "Policy 'ponti-stock-negative-notify'",
	}}
	svc := businessinsights.NewService(repo, nil, nil, review, businessinsights.Config{})

	err := svc.NotifyStockNegative(context.Background(), uuid.New(), "user-1", businessinsights.StockLevel{
		ProductID:   "p-1",
		ProductName: "Fertilizante",
		Quantity:    -3,
	})
	if err != nil {
		t.Fatalf("NotifyStockNegative: %v", err)
	}
	if review.calls != 1 {
		t.Fatalf("review calls = %d, want 1", review.calls)
	}
	if repo.upsertCalls != 1 {
		t.Fatalf("upsert calls = %d, want 1", repo.upsertCalls)
	}
	if repo.markCalls != 1 {
		t.Fatalf("mark notified calls = %d, want 1", repo.markCalls)
	}
	if repo.lastUpsert.EventType != "ponti.stock.negative" {
		t.Fatalf("event_type = %q", repo.lastUpsert.EventType)
	}
	if repo.lastUpsert.Evidence["source_ref"] != "ponti.stock.real" {
		t.Fatalf("source ref mismatch: %#v", repo.lastUpsert.Evidence)
	}
	if repo.lastUpsert.Evidence["suggested_action"] != "review_stock_movements" {
		t.Fatalf("suggested action mismatch: %#v", repo.lastUpsert.Evidence)
	}
}

func TestNotifyStockNegative_NoPolicyMatch_SkipsUpsert(t *testing.T) {
	repo := &stubRepo{}
	review := &stubReview{response: governanceclient.SubmitResponse{
		RequestID:      "req-2",
		Decision:       "allow",
		DecisionReason: "No policy matched; default for risk low",
	}}
	svc := businessinsights.NewService(repo, nil, nil, review, businessinsights.Config{})

	err := svc.NotifyStockNegative(context.Background(), uuid.New(), "user-1", businessinsights.StockLevel{
		ProductID: "p-1", Quantity: -1,
	})
	if err != nil {
		t.Fatalf("NotifyStockNegative: %v", err)
	}
	if review.calls != 1 {
		t.Fatalf("review calls = %d, want 1", review.calls)
	}
	if repo.upsertCalls != 0 {
		t.Fatalf("upsert calls = %d, want 0 (no policy match)", repo.upsertCalls)
	}
}

func TestNotifyStockNegative_PositiveStock_SkipsReview(t *testing.T) {
	repo := &stubRepo{}
	review := &stubReview{}
	svc := businessinsights.NewService(repo, nil, nil, review, businessinsights.Config{})

	err := svc.NotifyStockNegative(context.Background(), uuid.New(), "user-1", businessinsights.StockLevel{
		ProductID: "p-1", Quantity: 5,
	})
	if err != nil {
		t.Fatalf("NotifyStockNegative: %v", err)
	}
	if review.calls != 0 {
		t.Fatalf("review calls = %d, want 0 (quantity >= 0)", review.calls)
	}
	if repo.upsertCalls != 0 {
		t.Fatalf("upsert calls = %d, want 0", repo.upsertCalls)
	}
}

func TestNotifyStockNegative_DedupBucketConsistent(t *testing.T) {
	repo := &stubRepo{shouldNotify: false} // segunda invocacion: no re-notifica
	review := &stubReview{response: governanceclient.SubmitResponse{
		Decision:       "allow",
		DecisionReason: "Policy 'p'",
	}}
	svc := businessinsights.NewService(repo, nil, nil, review, businessinsights.Config{
		NegativeStockDedupWindow: 24 * time.Hour,
	})
	level := businessinsights.StockLevel{ProductID: "p-x", Quantity: -2}

	if err := svc.NotifyStockNegative(context.Background(), uuid.New(), "u", level); err != nil {
		t.Fatal(err)
	}
	fpFirst := repo.lastUpsert.Fingerprint

	if err := svc.NotifyStockNegative(context.Background(), uuid.New(), "u", level); err != nil {
		t.Fatal(err)
	}
	fpSecond := repo.lastUpsert.Fingerprint

	if fpFirst != fpSecond {
		t.Fatalf("fingerprint difiere entre invocaciones dentro de la misma ventana: %q vs %q", fpFirst, fpSecond)
	}
	if repo.markCalls != 0 {
		t.Fatalf("mark calls = %d, want 0 (no debe re-notificar)", repo.markCalls)
	}
}

func TestNotifyStockNegative_ReviewError_PropagatesFailure(t *testing.T) {
	repo := &stubRepo{}
	review := &stubReview{err: errors.New("network down")}
	svc := businessinsights.NewService(repo, nil, nil, review, businessinsights.Config{})

	err := svc.NotifyStockNegative(context.Background(), uuid.New(), "u", businessinsights.StockLevel{
		ProductID: "p-1", Quantity: -1,
	})
	if err == nil {
		t.Fatal("expected error when review client fails")
	}
	if repo.upsertCalls != 0 {
		t.Fatalf("upsert should not be called when review fails")
	}
}

func TestNotifyStockNegative_NilReview_NoOp(t *testing.T) {
	repo := &stubRepo{}
	svc := businessinsights.NewService(repo, nil, nil, nil, businessinsights.Config{})

	err := svc.NotifyStockNegative(context.Background(), uuid.New(), "u", businessinsights.StockLevel{
		ProductID: "p-1", Quantity: -1,
	})
	if err != nil {
		t.Fatalf("NotifyStockNegative should no-op silently when review is nil, got: %v", err)
	}
	if repo.upsertCalls != 0 {
		t.Fatalf("upsert should not be called when review is nil")
	}
}

func TestNotifyDataIntegrityCritical_RecordsReadOnlyCandidate(t *testing.T) {
	repo := &stubRepo{shouldNotify: true}
	svc := businessinsights.NewService(repo, nil, nil, nil, businessinsights.Config{})
	tenantID := uuid.New()

	err := svc.NotifyDataIntegrityCritical(context.Background(), tenantID, "user-1", businessinsights.DataIntegrityCritical{
		ProjectID:    "4",
		FailedChecks: 3,
		TotalChecks:  17,
		Controls: []businessinsights.DataIntegrityControlIssue{
			{ControlNumber: 1, DataToVerify: "Costos directos", DifferenceA: "100.00"},
		},
	})
	if err != nil {
		t.Fatalf("NotifyDataIntegrityCritical: %v", err)
	}
	if repo.upsertCalls != 1 {
		t.Fatalf("upsert calls = %d, want 1", repo.upsertCalls)
	}
	if repo.markCalls != 1 {
		t.Fatalf("mark notified calls = %d, want 1", repo.markCalls)
	}
	if repo.lastUpsert.EventType != "ponti.data_integrity.critical" {
		t.Fatalf("event_type = %q", repo.lastUpsert.EventType)
	}
	if repo.lastUpsert.Kind != "integrity" || repo.lastUpsert.EntityType != "project" || repo.lastUpsert.EntityID != "4" {
		t.Fatalf("unexpected candidate shape: %#v", repo.lastUpsert)
	}
	if repo.lastUpsert.Severity != "critical" {
		t.Fatalf("severity = %q", repo.lastUpsert.Severity)
	}
	if repo.lastUpsert.Evidence["failed_checks"] != 3 {
		t.Fatalf("evidence missing failed checks: %#v", repo.lastUpsert.Evidence)
	}
	if repo.lastUpsert.Evidence["source_ref"] != "ponti.data_integrity.costs_check" {
		t.Fatalf("source ref mismatch: %#v", repo.lastUpsert.Evidence)
	}
	workspace, ok := repo.lastUpsert.Evidence["workspace"].(map[string]any)
	if !ok || workspace["project_id"] != "4" {
		t.Fatalf("workspace mismatch: %#v", repo.lastUpsert.Evidence)
	}
}

func TestNotifyDataIntegrityCritical_ZeroFailuresResolvesExistingCandidate(t *testing.T) {
	repo := &stubRepo{}
	resolver := &stubResolver{}
	svc := businessinsights.NewService(repo, resolver, nil, nil, businessinsights.Config{})
	tenantID := uuid.New()

	err := svc.NotifyDataIntegrityCritical(context.Background(), tenantID, "user-1", businessinsights.DataIntegrityCritical{
		ProjectID:    "4",
		FailedChecks: 0,
		TotalChecks:  17,
	})
	if err != nil {
		t.Fatalf("NotifyDataIntegrityCritical: %v", err)
	}
	if repo.upsertCalls != 0 {
		t.Fatalf("upsert calls = %d, want 0", repo.upsertCalls)
	}
	if resolver.resolveByEntityCalls != 1 {
		t.Fatalf("resolve calls = %d, want 1", resolver.resolveByEntityCalls)
	}
	if resolver.lastResolveByEntity[1] != "ponti.data_integrity.critical" || resolver.lastResolveByEntity[3] != "4" {
		t.Fatalf("unexpected resolve target: %#v", resolver.lastResolveByEntity)
	}
}

func TestNotifyOperatingResultNegative_RecordsMarginCandidate(t *testing.T) {
	repo := &stubRepo{shouldNotify: true}
	svc := businessinsights.NewService(repo, nil, nil, nil, businessinsights.Config{})
	tenantID := uuid.New()

	err := svc.NotifyOperatingResultNegative(context.Background(), tenantID, "user-1", businessinsights.OperatingResultNegative{
		ProjectID:               "4",
		CustomerID:              "1",
		CampaignID:              "2",
		TotalOperatingResultUSD: "-1234.50",
		ProjectReturnPct:        "-8.2",
		TotalInvestedProjectUSD: "15000",
		NegativeCrops: []businessinsights.OperatingResultNegativeCrop{
			{CropID: "9", CropName: "Soja", OperatingResultUSD: "-1234.50"},
		},
	})
	if err != nil {
		t.Fatalf("NotifyOperatingResultNegative: %v", err)
	}
	if repo.upsertCalls != 1 {
		t.Fatalf("upsert calls = %d, want 1", repo.upsertCalls)
	}
	if repo.markCalls != 1 {
		t.Fatalf("mark notified calls = %d, want 1", repo.markCalls)
	}
	if repo.lastUpsert.EventType != "ponti.report.operating_result.negative" {
		t.Fatalf("event_type = %q", repo.lastUpsert.EventType)
	}
	if repo.lastUpsert.Kind != "margin" || repo.lastUpsert.EntityType != "project" || repo.lastUpsert.EntityID != "4" {
		t.Fatalf("unexpected candidate shape: %#v", repo.lastUpsert)
	}
	if repo.lastUpsert.Evidence["source_ref"] != "ponti.reports.summary_results" {
		t.Fatalf("source ref mismatch: %#v", repo.lastUpsert.Evidence)
	}
	workspace, ok := repo.lastUpsert.Evidence["workspace"].(map[string]any)
	if !ok || workspace["project_id"] != "4" || workspace["customer_id"] != "1" || workspace["campaign_id"] != "2" {
		t.Fatalf("workspace mismatch: %#v", repo.lastUpsert.Evidence)
	}
}

func TestMaybeResolveOperatingResultNegative_ResolvesExistingCandidate(t *testing.T) {
	resolver := &stubResolver{}
	svc := businessinsights.NewService(&stubRepo{}, resolver, nil, nil, businessinsights.Config{})
	tenantID := uuid.New()

	err := svc.MaybeResolveOperatingResultNegative(context.Background(), tenantID, "4")
	if err != nil {
		t.Fatalf("MaybeResolveOperatingResultNegative: %v", err)
	}
	if resolver.resolveByEntityCalls != 1 {
		t.Fatalf("resolve calls = %d, want 1", resolver.resolveByEntityCalls)
	}
	if resolver.lastResolveByEntity[1] != "ponti.report.operating_result.negative" || resolver.lastResolveByEntity[3] != "4" {
		t.Fatalf("unexpected resolve target: %#v", resolver.lastResolveByEntity)
	}
}

func TestNotifyTentativePrices_RecordsIntegrityCandidate(t *testing.T) {
	repo := &stubRepo{shouldNotify: true}
	svc := businessinsights.NewService(repo, nil, nil, nil, businessinsights.Config{})
	tenantID := uuid.New()

	err := svc.NotifyTentativePrices(context.Background(), tenantID, "user-1", businessinsights.TentativePricesIssue{
		ProjectID:  "4",
		CustomerID: "1",
		Count:      2,
		SampleItems: []businessinsights.TentativePriceItem{
			{SupplyID: "10", Name: "Insumo", Price: "123.45"},
		},
	})
	if err != nil {
		t.Fatalf("NotifyTentativePrices: %v", err)
	}
	if repo.upsertCalls != 1 {
		t.Fatalf("upsert calls = %d, want 1", repo.upsertCalls)
	}
	if repo.markCalls != 1 {
		t.Fatalf("mark notified calls = %d, want 1", repo.markCalls)
	}
	if repo.lastUpsert.EventType != "ponti.data_integrity.tentative_prices" {
		t.Fatalf("event_type = %q", repo.lastUpsert.EventType)
	}
	if repo.lastUpsert.Kind != "integrity" || repo.lastUpsert.EntityID != "4" {
		t.Fatalf("unexpected candidate shape: %#v", repo.lastUpsert)
	}
	if repo.lastUpsert.Evidence["source_ref"] != "ponti.data_integrity.tentative_prices" {
		t.Fatalf("source ref mismatch: %#v", repo.lastUpsert.Evidence)
	}
	workspace, ok := repo.lastUpsert.Evidence["workspace"].(map[string]any)
	if !ok || workspace["project_id"] != "4" || workspace["customer_id"] != "1" {
		t.Fatalf("workspace mismatch: %#v", repo.lastUpsert.Evidence)
	}
}

func TestNotifyTentativePrices_ZeroCountResolvesExistingCandidate(t *testing.T) {
	resolver := &stubResolver{}
	svc := businessinsights.NewService(&stubRepo{}, resolver, nil, nil, businessinsights.Config{})
	tenantID := uuid.New()

	err := svc.NotifyTentativePrices(context.Background(), tenantID, "user-1", businessinsights.TentativePricesIssue{
		ProjectID: "4",
		Count:     0,
	})
	if err != nil {
		t.Fatalf("NotifyTentativePrices: %v", err)
	}
	if resolver.resolveByEntityCalls != 1 {
		t.Fatalf("resolve calls = %d, want 1", resolver.resolveByEntityCalls)
	}
	if resolver.lastResolveByEntity[1] != "ponti.data_integrity.tentative_prices" || resolver.lastResolveByEntity[3] != "4" {
		t.Fatalf("unexpected resolve target: %#v", resolver.lastResolveByEntity)
	}
}
