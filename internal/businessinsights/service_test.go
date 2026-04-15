package businessinsights_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/devpablocristo/core/governance/go/reviewclient"
	"github.com/google/uuid"

	"github.com/devpablocristo/ponti-backend/internal/businessinsights"
)

type stubReview struct {
	calls    int
	response reviewclient.SubmitResponse
	err      error
	lastBody reviewclient.SubmitRequestBody
}

func (s *stubReview) SubmitRequest(_ context.Context, _ string, body reviewclient.SubmitRequestBody) (reviewclient.SubmitResponse, error) {
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

func TestNotifyStockNegative_PolicyMatched_NotifiesOnce(t *testing.T) {
	repo := &stubRepo{shouldNotify: true}
	review := &stubReview{response: reviewclient.SubmitResponse{
		RequestID:      "req-1",
		Decision:       "allow",
		DecisionReason: "Policy 'ponti-stock-negative-notify'",
	}}
	svc := businessinsights.NewService(repo, review, businessinsights.Config{})

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
}

func TestNotifyStockNegative_NoPolicyMatch_SkipsUpsert(t *testing.T) {
	repo := &stubRepo{}
	review := &stubReview{response: reviewclient.SubmitResponse{
		RequestID:      "req-2",
		Decision:       "allow",
		DecisionReason: "No policy matched; default for risk low",
	}}
	svc := businessinsights.NewService(repo, review, businessinsights.Config{})

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
	svc := businessinsights.NewService(repo, review, businessinsights.Config{})

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
	review := &stubReview{response: reviewclient.SubmitResponse{
		Decision:       "allow",
		DecisionReason: "Policy 'p'",
	}}
	svc := businessinsights.NewService(repo, review, businessinsights.Config{
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
	svc := businessinsights.NewService(repo, review, businessinsights.Config{})

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
	svc := businessinsights.NewService(repo, nil, businessinsights.Config{})

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
