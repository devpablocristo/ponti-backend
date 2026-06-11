package businessinsights_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/devpablocristo/platform/kernels/governance/go/governanceclient"
	"github.com/google/uuid"

	bparamsdomain "github.com/devpablocristo/ponti-backend/internal/business-parameters/usecases/domain"
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

type stubParams struct {
	calls int
	value string
	err   error
}

func (s *stubParams) GetParameter(_ context.Context, key string) (*bparamsdomain.BusinessParameter, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	return &bparamsdomain.BusinessParameter{Key: key, Value: s.value, Type: "decimal"}, nil
}

func TestNotifyStockLow_BelowThreshold_RecordsCandidate(t *testing.T) {
	repo := &stubRepo{shouldNotify: true}
	svc := businessinsights.NewService(repo, nil, nil, nil, businessinsights.Config{
		LowStockEnabled:   true,
		LowStockThreshold: 10,
	})
	tenantID := uuid.New()

	err := svc.NotifyStockLow(context.Background(), tenantID, "user-1", businessinsights.StockLowLevel{
		SupplyID:   "p-1",
		StockID:    "s-1",
		SupplyName: "Fertilizante",
		Level:      3,
	})
	if err != nil {
		t.Fatalf("NotifyStockLow: %v", err)
	}
	if repo.upsertCalls != 1 {
		t.Fatalf("upsert calls = %d, want 1", repo.upsertCalls)
	}
	if repo.markCalls != 1 {
		t.Fatalf("mark notified calls = %d, want 1", repo.markCalls)
	}
	if repo.lastUpsert.EventType != "ponti.stock.low" {
		t.Fatalf("event_type = %q", repo.lastUpsert.EventType)
	}
	if repo.lastUpsert.Kind != "insight" || repo.lastUpsert.EntityType != "supply" || repo.lastUpsert.EntityID != "p-1" {
		t.Fatalf("unexpected candidate shape: %#v", repo.lastUpsert)
	}
	if repo.lastUpsert.Severity != "info" {
		t.Fatalf("severity = %q, want info (menor que warning)", repo.lastUpsert.Severity)
	}
	if repo.lastUpsert.Evidence["source_ref"] != "ponti.stock.low" {
		t.Fatalf("source ref mismatch: %#v", repo.lastUpsert.Evidence)
	}
	if repo.lastUpsert.Evidence["level"] != 3.0 || repo.lastUpsert.Evidence["threshold"] != 10.0 {
		t.Fatalf("level/threshold mismatch: %#v", repo.lastUpsert.Evidence)
	}
	if repo.lastUpsert.Evidence["supply_id"] != "p-1" || repo.lastUpsert.Evidence["stock_id"] != "s-1" {
		t.Fatalf("supply/stock id mismatch: %#v", repo.lastUpsert.Evidence)
	}
}

func TestNotifyStockLow_LevelAtThreshold_ResolvesCandidate(t *testing.T) {
	repo := &stubRepo{}
	resolver := &stubResolver{}
	svc := businessinsights.NewService(repo, resolver, nil, nil, businessinsights.Config{
		LowStockEnabled:   true,
		LowStockThreshold: 10,
	})
	tenantID := uuid.New()

	err := svc.NotifyStockLow(context.Background(), tenantID, "user-1", businessinsights.StockLowLevel{
		SupplyID: "p-1",
		Level:    12,
	})
	if err != nil {
		t.Fatalf("NotifyStockLow: %v", err)
	}
	if repo.upsertCalls != 0 {
		t.Fatalf("upsert calls = %d, want 0", repo.upsertCalls)
	}
	if resolver.resolveByEntityCalls != 1 {
		t.Fatalf("resolve calls = %d, want 1", resolver.resolveByEntityCalls)
	}
	if resolver.lastResolveByEntity[1] != "ponti.stock.low" || resolver.lastResolveByEntity[3] != "p-1" {
		t.Fatalf("unexpected resolve target: %#v", resolver.lastResolveByEntity)
	}
}

func TestNotifyStockLow_FlagOff_NoOp(t *testing.T) {
	repo := &stubRepo{}
	resolver := &stubResolver{}
	params := &stubParams{value: "10"}
	svc := businessinsights.NewService(repo, resolver, nil, nil, businessinsights.Config{
		LowStockThreshold: 10,
	})
	svc.SetBusinessParameters(params)
	tenantID := uuid.New()

	err := svc.NotifyStockLow(context.Background(), tenantID, "user-1", businessinsights.StockLowLevel{
		SupplyID: "p-1",
		Level:    1,
	})
	if err != nil {
		t.Fatalf("NotifyStockLow: %v", err)
	}
	if repo.upsertCalls != 0 || resolver.resolveByEntityCalls != 0 || params.calls != 0 {
		t.Fatalf("expected full no-op with flag off: %#v %#v %#v", repo, resolver, params)
	}
}

func TestNotifyStockLow_ThresholdZero_NoOp(t *testing.T) {
	repo := &stubRepo{}
	resolver := &stubResolver{}
	svc := businessinsights.NewService(repo, resolver, nil, nil, businessinsights.Config{
		LowStockEnabled: true,
	})
	tenantID := uuid.New()

	err := svc.NotifyStockLow(context.Background(), tenantID, "user-1", businessinsights.StockLowLevel{
		SupplyID: "p-1",
		Level:    1,
	})
	if err != nil {
		t.Fatalf("NotifyStockLow: %v", err)
	}
	if repo.upsertCalls != 0 || resolver.resolveByEntityCalls != 0 {
		t.Fatalf("expected no-op without threshold: %#v %#v", repo, resolver)
	}
}

func TestNotifyStockLow_TenantParameterOverridesFallback(t *testing.T) {
	repo := &stubRepo{shouldNotify: true}
	params := &stubParams{value: "20"}
	svc := businessinsights.NewService(repo, nil, nil, nil, businessinsights.Config{
		LowStockEnabled:   true,
		LowStockThreshold: 5, // con el fallback no dispararía (10 >= 5)
	})
	svc.SetBusinessParameters(params)
	tenantID := uuid.New()

	err := svc.NotifyStockLow(context.Background(), tenantID, "user-1", businessinsights.StockLowLevel{
		SupplyID: "p-1",
		Level:    10,
	})
	if err != nil {
		t.Fatalf("NotifyStockLow: %v", err)
	}
	if params.calls != 1 {
		t.Fatalf("params calls = %d, want 1", params.calls)
	}
	if repo.upsertCalls != 1 {
		t.Fatalf("upsert calls = %d, want 1 (umbral per-tenant 20 > nivel 10)", repo.upsertCalls)
	}
	if repo.lastUpsert.Evidence["threshold"] != 20.0 {
		t.Fatalf("threshold mismatch: %#v", repo.lastUpsert.Evidence)
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
