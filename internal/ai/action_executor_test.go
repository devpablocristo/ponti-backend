package ai

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/devpablocristo/ponti-backend/internal/governance"
	projectdomain "github.com/devpablocristo/ponti-backend/internal/project/usecases/domain"
	stockdomain "github.com/devpablocristo/ponti-backend/internal/stock/usecases/domain"
	supplydomain "github.com/devpablocristo/ponti-backend/internal/supply/usecases/domain"
)

type stubStockWriter struct {
	stock           *stockdomain.Stock
	getByIDCalls    int
	getLastCalls    int
	updateCalls     int
	lastUpdatedID   int64
	lastUpdatedUnit decimal.Decimal
	lastUpdatedBy   string
}

func (s *stubStockWriter) GetStockByID(_ context.Context, _ int64) (*stockdomain.Stock, error) {
	s.getByIDCalls++
	return s.stock, nil
}

func (s *stubStockWriter) GetLastStockByProjectID(_ context.Context, _ int64, _ int64) (*stockdomain.Stock, bool, error) {
	s.getLastCalls++
	return s.stock, s.stock != nil, nil
}

func (s *stubStockWriter) UpdateRealStockUnits(_ context.Context, stockID int64, stock *stockdomain.Stock) error {
	s.updateCalls++
	s.lastUpdatedID = stockID
	s.lastUpdatedUnit = stock.RealStockUnits
	if stock.UpdatedBy != nil {
		s.lastUpdatedBy = *stock.UpdatedBy
	}
	return nil
}

func testStock(stockID, projectID, supplyID int64) *stockdomain.Stock {
	return &stockdomain.Stock{
		ID:      stockID,
		Project: &projectdomain.Project{ID: projectID},
		Supply:  &supplydomain.Supply{ID: supplyID, Name: "Glifosato"},
	}
}

func TestApplyInsightResolutionAppliesVerifiedWrite(t *testing.T) {
	t.Parallel()
	db := newActionDraftTestDB(t)
	resolver := &stubInsightResolver{}
	verifier := &stubActionVerifier{}
	executor := NewActionExecutor(NewActionDraftRepository(db), resolver, nil, verifier, ActionExecutorConfig{GovernedWritesEnabled: true})
	tenantID := uuid.New()
	insightID := uuid.NewString()

	result, err := executor.ApplyInsightResolution(context.Background(), tenantID, "owner-1", "req-approved", InsightResolutionInput{
		InsightID:      insightID,
		ResolutionNote: "Resolución aprobada.",
	})
	if err != nil {
		t.Fatalf("apply insight resolution: %v", err)
	}
	if !result.Applied || result.Status != actionDraftStatusApplied {
		t.Fatalf("expected applied result, got %+v", result)
	}
	if resolver.calls != 1 || resolver.lastID != insightID || resolver.lastActor != "owner-1" {
		t.Fatalf("expected resolver call, got %+v", resolver)
	}
	if verifier.lastActionType != pontiActionTypeInsightResolve {
		t.Fatalf("expected verification against per-tool action type, got %q", verifier.lastActionType)
	}
	var draft actionDraftModel
	if err := db.First(&draft, "id = ?", result.DraftID).Error; err != nil {
		t.Fatalf("load draft: %v", err)
	}
	if draft.Status != actionDraftStatusApplied || draft.AppliedAt == nil || draft.AppliedBy != "owner-1" {
		t.Fatalf("expected applied draft row, got %+v", draft)
	}
}

func TestApplyInsightResolutionStaysStagedWhenUnverified(t *testing.T) {
	t.Parallel()
	db := newActionDraftTestDB(t)
	resolver := &stubInsightResolver{}
	verifier := &stubActionVerifier{err: &governance.NotApprovedError{Detail: "nexus request not found"}}
	executor := NewActionExecutor(NewActionDraftRepository(db), resolver, nil, verifier, ActionExecutorConfig{GovernedWritesEnabled: true})

	result, err := executor.ApplyInsightResolution(context.Background(), uuid.New(), "owner-1", "req-unknown", InsightResolutionInput{
		InsightID: uuid.NewString(),
	})
	if err != nil {
		t.Fatalf("apply insight resolution: %v", err)
	}
	if result.Applied || result.Status != actionDraftStatusStaged {
		t.Fatalf("expected staged result, got %+v", result)
	}
	if resolver.calls != 0 {
		t.Fatalf("resolver must not run without verification, got %d calls", resolver.calls)
	}
}

func TestApplyInsightResolutionStaysStagedWhenWritesDisabled(t *testing.T) {
	t.Parallel()
	db := newActionDraftTestDB(t)
	resolver := &stubInsightResolver{}
	verifier := &stubActionVerifier{}
	executor := NewActionExecutor(NewActionDraftRepository(db), resolver, nil, verifier, ActionExecutorConfig{GovernedWritesEnabled: false})

	result, err := executor.ApplyInsightResolution(context.Background(), uuid.New(), "owner-1", "req-approved", InsightResolutionInput{
		InsightID: uuid.NewString(),
	})
	if err != nil {
		t.Fatalf("apply insight resolution: %v", err)
	}
	if result.Applied {
		t.Fatalf("expected staged result with writes disabled, got %+v", result)
	}
	if resolver.calls != 0 || verifier.calls != 0 {
		t.Fatalf("no write nor verification expected with flag off, got resolver=%d verifier=%d", resolver.calls, verifier.calls)
	}
}

func TestApplyInsightResolutionIsIdempotentPerNexusRequest(t *testing.T) {
	t.Parallel()
	db := newActionDraftTestDB(t)
	resolver := &stubInsightResolver{}
	verifier := &stubActionVerifier{}
	executor := NewActionExecutor(NewActionDraftRepository(db), resolver, nil, verifier, ActionExecutorConfig{GovernedWritesEnabled: true})
	tenantID := uuid.New()
	in := InsightResolutionInput{InsightID: uuid.NewString()}

	first, err := executor.ApplyInsightResolution(context.Background(), tenantID, "owner-1", "req-approved", in)
	if err != nil {
		t.Fatalf("first apply: %v", err)
	}
	second, err := executor.ApplyInsightResolution(context.Background(), tenantID, "owner-1", "req-approved", in)
	if err != nil {
		t.Fatalf("second apply: %v", err)
	}
	if first.DraftID != second.DraftID || !second.Applied {
		t.Fatalf("expected same applied draft on replay, got %+v vs %+v", first, second)
	}
	if resolver.calls != 1 {
		t.Fatalf("write must run once per nexus request, got %d calls", resolver.calls)
	}
}

func TestApplyStockCountAppliesVerifiedWriteWithSystemActor(t *testing.T) {
	t.Parallel()
	db := newActionDraftTestDB(t)
	stocks := &stubStockWriter{stock: testStock(44, 10, 70)}
	verifier := &stubActionVerifier{}
	executor := NewActionExecutor(NewActionDraftRepository(db), nil, stocks, verifier, ActionExecutorConfig{GovernedWritesEnabled: true})
	stockID := int64(44)

	result, err := executor.ApplyStockCount(context.Background(), uuid.New(), "", "req-approved", StockCountInput{
		ProjectID:      10,
		StockID:        &stockID,
		SupplyID:       70,
		RealStockUnits: 12.5,
		Reason:         "Conteo aprobado.",
	})
	if err != nil {
		t.Fatalf("apply stock count: %v", err)
	}
	if !result.Applied {
		t.Fatalf("expected applied result, got %+v", result)
	}
	if stocks.getByIDCalls != 1 || stocks.updateCalls != 1 || stocks.lastUpdatedID != 44 {
		t.Fatalf("expected stock write through GetStockByID, got %+v", stocks)
	}
	if !stocks.lastUpdatedUnit.Equal(decimal.NewFromFloat(12.5)) {
		t.Fatalf("expected real stock 12.5, got %s", stocks.lastUpdatedUnit)
	}
	if stocks.lastUpdatedBy != actionExecutorSystemActor {
		t.Fatalf("expected system actor on write, got %q", stocks.lastUpdatedBy)
	}
	if verifier.lastActionType != pontiActionTypeStockCountApply {
		t.Fatalf("expected stock count action type, got %q", verifier.lastActionType)
	}
}

func TestApplyStockCountResolvesStockByProjectWhenIDMissing(t *testing.T) {
	t.Parallel()
	db := newActionDraftTestDB(t)
	stocks := &stubStockWriter{stock: testStock(91, 10, 70)}
	executor := NewActionExecutor(NewActionDraftRepository(db), nil, stocks, &stubActionVerifier{}, ActionExecutorConfig{GovernedWritesEnabled: true})

	result, err := executor.ApplyStockCount(context.Background(), uuid.New(), "owner-1", "req-approved", StockCountInput{
		ProjectID:      10,
		SupplyID:       70,
		RealStockUnits: 3,
		Reason:         "Conteo aprobado.",
	})
	if err != nil {
		t.Fatalf("apply stock count: %v", err)
	}
	if !result.Applied || stocks.getLastCalls != 1 || stocks.lastUpdatedID != 91 {
		t.Fatalf("expected resolution via GetLastStockByProjectID, got %+v stocks=%+v", result, stocks)
	}
}

func TestApplyStockCountRejectsStockFromAnotherProject(t *testing.T) {
	t.Parallel()
	db := newActionDraftTestDB(t)
	stocks := &stubStockWriter{stock: testStock(44, 99, 70)}
	executor := NewActionExecutor(NewActionDraftRepository(db), nil, stocks, &stubActionVerifier{}, ActionExecutorConfig{GovernedWritesEnabled: true})
	stockID := int64(44)

	_, err := executor.ApplyStockCount(context.Background(), uuid.New(), "owner-1", "req-approved", StockCountInput{
		ProjectID:      10,
		StockID:        &stockID,
		SupplyID:       70,
		RealStockUnits: 3,
		Reason:         "Conteo aprobado.",
	})
	if err == nil {
		t.Fatal("expected project mismatch error")
	}
	if stocks.updateCalls != 0 {
		t.Fatalf("write must not run for foreign stock, got %d calls", stocks.updateCalls)
	}
}

func TestDispatchApprovedRoutesActionTypes(t *testing.T) {
	t.Parallel()
	db := newActionDraftTestDB(t)
	resolver := &stubInsightResolver{}
	stocks := &stubStockWriter{stock: testStock(44, 10, 70)}
	executor := NewActionExecutor(NewActionDraftRepository(db), resolver, stocks, &stubActionVerifier{}, ActionExecutorConfig{GovernedWritesEnabled: true})
	tenantID := uuid.New()
	insightID := uuid.NewString()

	result, err := executor.DispatchApproved(context.Background(), tenantID, pontiActionTypeInsightResolve, "req-1", map[string]any{
		"insight_id":      insightID,
		"resolution_note": "ok",
	}, "user:approver")
	if err != nil {
		t.Fatalf("dispatch insight resolve: %v", err)
	}
	if result["insight_id"] != insightID || result["draft_type"] != actionDraftTypeInsightResolution {
		t.Fatalf("unexpected dispatch result %#v", result)
	}
	if resolver.calls != 1 || resolver.lastActor != "user:approver" {
		t.Fatalf("expected resolver call on dispatch, got %+v", resolver)
	}

	// Params como llegan de JSON (float64) deben castear bien.
	result, err = executor.DispatchApproved(context.Background(), tenantID, pontiActionTypeStockCountApply, "req-2", map[string]any{
		"project_id":       float64(10),
		"stock_id":         float64(44),
		"supply_id":        float64(70),
		"real_stock_units": 8.25,
		"reason":           "Conteo aprobado.",
	}, "user:approver")
	if err != nil {
		t.Fatalf("dispatch stock count: %v", err)
	}
	if result["draft_type"] != actionDraftTypeStockCount || stocks.updateCalls != 1 {
		t.Fatalf("expected stock count applied via dispatch, got %#v stocks=%+v", result, stocks)
	}

	if _, err := executor.DispatchApproved(context.Background(), tenantID, "ponti.unknown.action", "req-3", nil, ""); err == nil {
		t.Fatal("expected error for unsupported action type")
	}
}
