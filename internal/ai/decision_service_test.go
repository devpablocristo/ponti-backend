package ai

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
	stockdomain "github.com/devpablocristo/ponti-backend/internal/stock/usecases/domain"
	supplydomain "github.com/devpablocristo/ponti-backend/internal/supply/usecases/domain"
)

type stubDecisionStockSource struct {
	stocks []*stockdomain.Stock
	err    error
}

func (s stubDecisionStockSource) GetStocksSummary(context.Context, int64, time.Time) ([]*stockdomain.Stock, error) {
	return s.stocks, s.err
}

func TestDecisionServiceRequiresWorkspace(t *testing.T) {
	t.Parallel()
	db := newDecisionServiceTestDB(t)
	svc := NewDecisionService(NewDecisionRepository(db), nil, nil, nil)

	_, err := svc.Run(context.Background(), uuid.New(), "user-1", decisionRunInput{})
	if err == nil {
		t.Fatal("expected workspace validation error")
	}
}

func TestDecisionServiceGeneratesAndDedupesStockDecision(t *testing.T) {
	t.Parallel()
	db := newDecisionServiceTestDB(t)
	projectID := int64(30)
	customerID := int64(17)
	campaignID := int64(9)
	stockID := int64(44)
	supplyID := int64(70)
	svc := NewDecisionService(
		NewDecisionRepository(db),
		stubDecisionStockSource{stocks: []*stockdomain.Stock{{
			ID:                stockID,
			Supply:            &supplydomain.Supply{ID: supplyID, Name: "Glifosato", UnitName: "Lt"},
			RealStockUnits:    decimal.NewFromInt(-3),
			HasRealStockCount: true,
			Base:              shareddomain.Base{UpdatedAt: time.Now().UTC()},
		}}},
		nil,
		nil,
	)
	input := decisionRunInput{Workspace: workspaceRequest{
		CustomerID: &customerID,
		ProjectID:  &projectID,
		CampaignID: &campaignID,
	}}
	tenantID := uuid.New()

	first, err := svc.Run(context.Background(), tenantID, "owner-1", input)
	if err != nil {
		t.Fatalf("run first decision analysis: %v", err)
	}
	if first.Run.CardsCreated != 1 || first.Run.CardsUpdated != 0 || len(first.Cards) != 1 {
		t.Fatalf("unexpected first run counters: run=%#v cards=%d", first.Run, len(first.Cards))
	}
	card := first.Cards[0]
	if card.Domain != "stock" || card.RouteHint != "stock" || card.Severity != "critical" || card.Bucket != decisionBucketUrgent {
		t.Fatalf("unexpected stock card classification: %#v", card)
	}
	if card.Status != decisionStatusOpen {
		t.Fatalf("unexpected card status: %s", card.Status)
	}
	action := unmarshalDecisionMap(card.ActionJSON)
	if action["capability_id"] != "ponti.stock_count.draft" || action["requires_approval"] != true {
		t.Fatalf("unexpected governed action: %#v", action)
	}

	second, err := svc.Run(context.Background(), tenantID, "owner-1", input)
	if err != nil {
		t.Fatalf("run second decision analysis: %v", err)
	}
	if second.Run.CardsCreated != 0 || second.Run.CardsUpdated != 1 || len(second.Cards) != 1 {
		t.Fatalf("expected dedupe update on second run, got run=%#v cards=%d", second.Run, len(second.Cards))
	}
	if second.Cards[0].ID != card.ID || second.Cards[0].OccurrenceCount != 2 {
		t.Fatalf("expected same card occurrence increment, got first=%s second=%#v", card.ID, second.Cards[0])
	}
}

func TestDecisionServicePrepareActionDoesNotWriteFinalDraft(t *testing.T) {
	t.Parallel()
	db := newDecisionServiceTestDB(t)
	projectID := int64(30)
	customerID := int64(17)
	campaignID := int64(9)
	svc := NewDecisionService(
		NewDecisionRepository(db),
		stubDecisionStockSource{stocks: []*stockdomain.Stock{{
			ID:                44,
			Supply:            &supplydomain.Supply{ID: 70, Name: "Glifosato", UnitName: "Lt"},
			RealStockUnits:    decimal.NewFromInt(-3),
			HasRealStockCount: true,
			Base:              shareddomain.Base{UpdatedAt: time.Now().UTC()},
		}}},
		nil,
		nil,
	)
	tenantID := uuid.New()
	run, err := svc.Run(context.Background(), tenantID, "owner-1", decisionRunInput{Workspace: workspaceRequest{
		CustomerID: &customerID,
		ProjectID:  &projectID,
		CampaignID: &campaignID,
	}})
	if err != nil {
		t.Fatalf("run decision analysis: %v", err)
	}

	card, action, err := svc.PrepareCardAction(context.Background(), tenantID, run.Cards[0].ID.String(), "create_stock_count_draft", "owner-1")
	if err != nil {
		t.Fatalf("prepare card action: %v", err)
	}
	if card.Status != decisionStatusAccepted {
		t.Fatalf("expected accepted card after action preparation, got %s", card.Status)
	}
	if action["requires_approval"] != true || action["nexus_action_type"] != pontiActionTypeStockCountApply {
		t.Fatalf("expected nexus-governed pending action, got %#v", action)
	}
	if _, ok := action["draft_id"]; ok {
		t.Fatalf("decision action must not fabricate final draft ids: %#v", action)
	}
}

func TestDecisionServiceImportsExternalAxisDecision(t *testing.T) {
	t.Parallel()
	db := newDecisionServiceTestDB(t)
	projectID := int64(30)
	customerID := int64(17)
	campaignID := int64(9)
	tenantID := uuid.New()
	svc := NewDecisionService(NewDecisionRepository(db), nil, nil, nil)
	input := externalDecisionInput{
		Workspace: workspaceRequest{
			CustomerID: &customerID,
			ProjectID:  &projectID,
			CampaignID: &campaignID,
		},
		Fingerprint:    "axis:watcher:proposal-1",
		Domain:         "stock",
		RouteHint:      "stock",
		Severity:       "warning",
		Bucket:         "important",
		Title:          "Axis detectó stock a revisar",
		Summary:        "Watcher de Axis encontró una propuesta de stock.",
		Recommendation: "Revisar evidencia antes de preparar conteo.",
		Source:         "axis.watcher",
		Evidence:       map[string]any{"proposal_id": "proposal-1"},
		Tools:          []any{map[string]any{"name": "ponti.stock.summary"}},
		AxisRunID:      "run-axis",
		AxisTaskID:     "task-axis",
	}

	first, err := svc.ImportExternalCard(context.Background(), tenantID, "owner-1", input)
	if err != nil {
		t.Fatalf("import external decision: %v", err)
	}
	if first.Domain != "stock" || first.RouteHint != "stock" || first.Source != "axis.watcher" {
		t.Fatalf("unexpected external card: %#v", first)
	}
	if first.AxisRunID != "run-axis" || first.AxisTaskID != "task-axis" {
		t.Fatalf("axis trace ids were not preserved: %#v", first)
	}
	evidence := unmarshalDecisionMap(first.EvidenceJSON)
	if evidence["proposal_id"] != "proposal-1" || evidence["workspace"] == nil {
		t.Fatalf("unexpected external evidence: %#v", evidence)
	}

	second, err := svc.ImportExternalCard(context.Background(), tenantID, "owner-1", input)
	if err != nil {
		t.Fatalf("reimport external decision: %v", err)
	}
	if second.ID != first.ID || second.OccurrenceCount != 2 {
		t.Fatalf("expected fingerprint dedupe, first=%s second=%#v", first.ID, second)
	}
}

func TestDecisionServiceGeneratesLotRiskDecision(t *testing.T) {
	t.Parallel()
	db := newDecisionServiceTestDB(t)
	projectID := int64(30)
	customerID := int64(17)
	campaignID := int64(9)
	fieldID := int64(39)
	if err := db.Exec(`INSERT INTO v4_report.lot_list (
		project_id, field_id, field_name, id, lot_name, current_crop_id, current_crop,
		sowed_area_ha, harvested_area_ha, cost_usd_per_ha, yield_tn_per_ha, operating_result_per_ha_usd
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		projectID, fieldID, "SJDD", 102, "LOTE 54", 8, "Poroto rojo",
		decimal.NewFromInt(77), decimal.Zero, decimal.NewFromInt(234), decimal.Zero, decimal.NewFromInt(-434),
	).Error; err != nil {
		t.Fatalf("insert lot signal: %v", err)
	}
	svc := NewDecisionService(NewDecisionRepository(db), nil, nil, nil)

	run, err := svc.Run(context.Background(), uuid.New(), "owner-1", decisionRunInput{Workspace: workspaceRequest{
		CustomerID: &customerID,
		ProjectID:  &projectID,
		CampaignID: &campaignID,
		FieldID:    &fieldID,
	}})
	if err != nil {
		t.Fatalf("run decision analysis: %v", err)
	}
	if len(run.Cards) != 1 {
		t.Fatalf("expected one lot card, got %d", len(run.Cards))
	}
	card := run.Cards[0]
	if card.Domain != "lots" || card.RouteHint != "lots" || card.Source != "ponti.lots.summary" {
		t.Fatalf("unexpected lot card: %#v", card)
	}
	evidence := unmarshalDecisionMap(card.EvidenceJSON)
	if evidence["source"] != "ponti.lots.summary" {
		t.Fatalf("unexpected lot evidence: %#v", evidence)
	}
}

func newDecisionServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&decisionRunModel{}, &decisionCardModel{}); err != nil {
		t.Fatalf("auto migrate decisions: %v", err)
	}
	if err := db.Exec(`CREATE TABLE supplies (
		id INTEGER PRIMARY KEY,
		project_id INTEGER,
		name TEXT,
		price TEXT,
		is_partial_price BOOLEAN,
		is_pending BOOLEAN,
		deleted_at DATETIME NULL
	);`).Error; err != nil {
		t.Fatalf("create supplies table: %v", err)
	}
	if err := db.Exec(`CREATE TABLE work_order_drafts (
		id INTEGER PRIMARY KEY,
		project_id INTEGER,
		field_id INTEGER,
		status TEXT,
		deleted_at DATETIME NULL
	);`).Error; err != nil {
		t.Fatalf("create work_order_drafts table: %v", err)
	}
	if err := db.Exec(`ATTACH DATABASE ':memory:' AS v4_report;`).Error; err != nil {
		t.Fatalf("attach v4_report schema: %v", err)
	}
	if err := db.Exec(`CREATE TABLE v4_report.lot_list (
		project_id INTEGER,
		field_id INTEGER,
		field_name TEXT,
		id INTEGER,
		lot_name TEXT,
		current_crop_id INTEGER,
		current_crop TEXT,
		sowed_area_ha TEXT,
		harvested_area_ha TEXT,
		cost_usd_per_ha TEXT,
		yield_tn_per_ha TEXT,
		operating_result_per_ha_usd TEXT
	);`).Error; err != nil {
		t.Fatalf("create lot_list view stub: %v", err)
	}
	return db
}
