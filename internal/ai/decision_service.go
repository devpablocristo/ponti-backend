package ai

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/devpablocristo/ponti-backend/internal/businessinsights"
	reportdomain "github.com/devpablocristo/ponti-backend/internal/report/usecases/domain"
	stockdomain "github.com/devpablocristo/ponti-backend/internal/stock/usecases/domain"
)

type stockDecisionSource interface {
	GetStocksSummary(ctx context.Context, projectID int64, closeDate time.Time) ([]*stockdomain.Stock, error)
}

type reportDecisionSource interface {
	GetSummaryResultsReport(ctx context.Context, filters reportdomain.SummaryResultsFilter) (*reportdomain.SummaryResultsResponse, error)
}

type insightsDecisionSource interface {
	ListByTenantForUser(ctx context.Context, tenantID, userID string, opts businessinsights.ListOptions) ([]businessinsights.CandidateView, error)
}

type DecisionService struct {
	repo     *decisionRepository
	stock    stockDecisionSource
	report   reportDecisionSource
	insights insightsDecisionSource
}

func NewDecisionService(repo *decisionRepository, stock stockDecisionSource, report reportDecisionSource, insights insightsDecisionSource) *DecisionService {
	return &DecisionService{repo: repo, stock: stock, report: report, insights: insights}
}

func (s *DecisionService) Run(ctx context.Context, tenantID uuid.UUID, actor string, in decisionRunInput) (decisionRunResult, error) {
	if s == nil || s.repo == nil {
		return decisionRunResult{}, domainerr.Internal("ai decisions unavailable")
	}
	if tenantID == uuid.Nil {
		return decisionRunResult{}, domainerr.Forbidden("invalid org_id")
	}
	if err := validateDecisionWorkspace(in.Workspace); err != nil {
		return decisionRunResult{}, err
	}

	now := time.Now().UTC()
	workspace := decisionWorkspaceMap(in.Workspace)
	run, err := s.repo.createRun(ctx, decisionRunModel{
		TenantID:      tenantID,
		WorkspaceJSON: marshalDecisionJSON(workspace, "{}"),
		RequestedBy:   actor,
		Status:        decisionRunStatusRunning,
		RoutingSource: "deterministic",
		StartedAt:     now,
	})
	if err != nil {
		return decisionRunResult{}, err
	}

	var drafts []decisionCardDraft
	var generatorErrors []string
	appendDrafts := func(source string, items []decisionCardDraft, err error) {
		if err != nil {
			generatorErrors = append(generatorErrors, source+": "+err.Error())
			return
		}
		drafts = append(drafts, items...)
	}

	stockDrafts, stockErr := s.generateStockCards(ctx, in.Workspace)
	appendDrafts("stock", stockDrafts, stockErr)
	insightDrafts, insightErr := s.generateInsightCards(ctx, tenantID, actor, in.Workspace)
	appendDrafts("insights", insightDrafts, insightErr)
	reportDrafts, reportErr := s.generateReportCards(ctx, in.Workspace)
	appendDrafts("reports", reportDrafts, reportErr)
	lotDrafts, lotErr := s.generateLotCards(ctx, in.Workspace)
	appendDrafts("lots", lotDrafts, lotErr)
	supplyDrafts, supplyErr := s.generateSupplyQualityCards(ctx, in.Workspace)
	appendDrafts("supplies", supplyDrafts, supplyErr)
	operationDrafts, operationErr := s.generateOperationCards(ctx, in.Workspace)
	appendDrafts("workorders", operationDrafts, operationErr)

	cards, created, updated, err := s.repo.upsertCards(ctx, tenantID, run.ID, workspace, actor, drafts)
	if err != nil {
		run.Status = decisionRunStatusFailed
		run.DegradedReason = err.Error()
		_, _ = s.repo.completeRun(ctx, run)
		return decisionRunResult{}, err
	}

	run.CardsCreated = created
	run.CardsUpdated = updated
	run.CardsTotal = len(cards)
	run.Status = decisionRunStatusCompleted
	if len(generatorErrors) > 0 {
		run.Status = decisionRunStatusDegraded
		run.DegradedReason = strings.Join(generatorErrors, " | ")
	}
	run, err = s.repo.completeRun(ctx, run)
	if err != nil {
		return decisionRunResult{}, err
	}
	return decisionRunResult{Run: run, Cards: cards}, nil
}

func (s *DecisionService) ListRuns(ctx context.Context, tenantID uuid.UUID, limit int) ([]decisionRunModel, error) {
	if s == nil || s.repo == nil {
		return nil, domainerr.Internal("ai decisions unavailable")
	}
	return s.repo.listRuns(ctx, tenantID, limit)
}

func (s *DecisionService) ListCards(ctx context.Context, tenantID uuid.UUID, filters decisionCardFilters) ([]decisionCardModel, error) {
	if s == nil || s.repo == nil {
		return nil, domainerr.Internal("ai decisions unavailable")
	}
	return s.repo.listCards(ctx, tenantID, filters)
}

func (s *DecisionService) PatchCardStatus(ctx context.Context, tenantID uuid.UUID, cardID string, actor string, status string, snoozeUntil *time.Time) (decisionCardModel, error) {
	if s == nil || s.repo == nil {
		return decisionCardModel{}, domainerr.Internal("ai decisions unavailable")
	}
	return s.repo.patchCardStatus(ctx, tenantID, cardID, actor, strings.TrimSpace(status), snoozeUntil)
}

func (s *DecisionService) PrepareCardAction(ctx context.Context, tenantID uuid.UUID, cardID string, actionID string, actor string) (decisionCardModel, map[string]any, error) {
	if s == nil || s.repo == nil {
		return decisionCardModel{}, nil, domainerr.Internal("ai decisions unavailable")
	}
	card, err := s.repo.getCard(ctx, tenantID, cardID)
	if err != nil {
		return decisionCardModel{}, nil, err
	}
	action := unmarshalDecisionMap(card.ActionJSON)
	expected := strings.TrimSpace(stringFromAny(action["id"]))
	if expected == "" {
		return decisionCardModel{}, nil, domainerr.Validation("decision card has no executable action")
	}
	if strings.TrimSpace(actionID) != expected {
		return decisionCardModel{}, nil, domainerr.Validation("action_id does not match decision card")
	}
	nextStatus := decisionStatusAccepted
	if approval, _ := action["requires_approval"].(bool); !approval {
		nextStatus = decisionStatusResolved
	}
	card, err = s.repo.patchCardStatus(ctx, tenantID, cardID, actor, nextStatus, nil)
	if err != nil {
		return decisionCardModel{}, nil, err
	}
	return card, action, nil
}

func (s *DecisionService) ImportExternalCard(ctx context.Context, tenantID uuid.UUID, actor string, in externalDecisionInput) (decisionCardModel, error) {
	if s == nil || s.repo == nil {
		return decisionCardModel{}, domainerr.Internal("ai decisions unavailable")
	}
	if tenantID == uuid.Nil {
		return decisionCardModel{}, domainerr.Forbidden("invalid org_id")
	}
	if err := validateDecisionWorkspace(in.Workspace); err != nil {
		return decisionCardModel{}, err
	}
	if strings.TrimSpace(in.Fingerprint) == "" {
		return decisionCardModel{}, domainerr.Validation("fingerprint is required")
	}
	if strings.TrimSpace(in.Title) == "" || strings.TrimSpace(in.Summary) == "" {
		return decisionCardModel{}, domainerr.Validation("title and summary are required")
	}

	now := time.Now().UTC()
	workspace := decisionWorkspaceMap(in.Workspace)
	run, err := s.repo.createRun(ctx, decisionRunModel{
		TenantID:      tenantID,
		WorkspaceJSON: marshalDecisionJSON(workspace, "{}"),
		RequestedBy:   actor,
		Status:        decisionRunStatusCompleted,
		RoutingSource: "axis.watcher",
		AxisRunID:     strings.TrimSpace(in.AxisRunID),
		AxisTaskID:    strings.TrimSpace(in.AxisTaskID),
		StartedAt:     now,
	})
	if err != nil {
		return decisionCardModel{}, err
	}

	draft := decisionCardDraft{
		Fingerprint:    strings.TrimSpace(in.Fingerprint),
		Domain:         nonEmptyString(in.Domain, "axis"),
		RouteHint:      nonEmptyString(in.RouteHint, "dashboard"),
		Severity:       nonEmptyString(normalizeDecisionSeverity(in.Severity), "info"),
		Bucket:         nonEmptyString(normalizeDecisionBucket(in.Bucket), decisionBucketFollowUp),
		Title:          strings.TrimSpace(in.Title),
		Summary:        strings.TrimSpace(in.Summary),
		Recommendation: nonEmptyString(in.Recommendation, "Revisar la evidencia de Axis y decidir el próximo paso."),
		ImpactLabel:    strings.TrimSpace(in.ImpactLabel),
		ImpactValue:    in.ImpactValue,
		Source:         nonEmptyString(in.Source, "axis.watcher"),
		Evidence:       in.Evidence,
		Tools:          in.Tools,
		Action:         in.Action,
		AxisRunID:      strings.TrimSpace(in.AxisRunID),
		AxisTaskID:     strings.TrimSpace(in.AxisTaskID),
	}
	if draft.Evidence == nil {
		draft.Evidence = map[string]any{}
	}
	draft.Evidence["workspace"] = workspace
	draft.Evidence["captured_at"] = now.Format(time.RFC3339)
	if len(draft.Tools) == 0 {
		draft.Tools = []any{draft.Source}
	}

	cards, created, updated, err := s.repo.upsertCards(ctx, tenantID, run.ID, workspace, actor, []decisionCardDraft{draft})
	if err != nil {
		run.Status = decisionRunStatusFailed
		run.DegradedReason = err.Error()
		_, _ = s.repo.completeRun(ctx, run)
		return decisionCardModel{}, err
	}
	run.CardsCreated = created
	run.CardsUpdated = updated
	run.CardsTotal = len(cards)
	_, _ = s.repo.completeRun(ctx, run)
	if len(cards) == 0 {
		return decisionCardModel{}, domainerr.Internal("external decision card was not persisted")
	}
	return cards[0], nil
}

func (s *DecisionService) generateStockCards(ctx context.Context, workspace workspaceRequest) ([]decisionCardDraft, error) {
	if s.stock == nil || workspace.ProjectID == nil {
		return []decisionCardDraft{}, nil
	}
	stocks, err := s.stock.GetStocksSummary(ctx, *workspace.ProjectID, time.Time{})
	if err != nil {
		return nil, err
	}
	out := make([]decisionCardDraft, 0)
	missingCounts := 0
	for _, st := range stocks {
		if st == nil || st.Supply == nil || st.Supply.ID <= 0 {
			continue
		}
		systemUnits := st.GetStockUnits()
		realUnits := st.RealStockUnits
		diff := st.GetStockDifference()
		supplyName := nonEmptyString(st.Supply.Name, "Insumo")
		unit := nonEmptyString(st.GetSupplyUnitName(), "unid.")

		if systemUnits.LessThan(decimal.Zero) || (st.HasRealStockCount && realUnits.LessThan(decimal.Zero)) {
			observed := systemUnits
			impactLabel := "Stock sistema: " + formatDecimal(systemUnits) + " " + unit
			if st.HasRealStockCount && realUnits.LessThan(systemUnits) {
				observed = realUnits
				impactLabel = "Stock campo: " + formatDecimal(realUnits) + " " + unit
			}
			impact := decimalAbsFloat(observed)
			out = append(out, decisionCardDraft{
				Fingerprint:    fmt.Sprintf("stock:negative:%d:%d", *workspace.ProjectID, st.Supply.ID),
				Domain:         "stock",
				RouteHint:      "stock",
				Severity:       "critical",
				Bucket:         decisionBucketUrgent,
				Title:          "Stock negativo en " + supplyName,
				Summary:        fmt.Sprintf("%s quedó con stock negativo. Revisá consumos, remitos y conteo de campo antes de nuevas labores.", supplyName),
				Recommendation: "Investigar movimientos recientes y preparar un conteo de stock gobernado si el dato de campo no está validado.",
				ImpactLabel:    impactLabel,
				ImpactValue:    &impact,
				Source:         "ponti.stock.summary",
				Evidence:       stockEvidence(workspace, st, "negative_stock"),
				Tools:          decisionTools("ponti.stock.summary"),
				Action:         stockCountAction(workspace, st, "Regularizar stock negativo con conteo aprobado"),
			})
			continue
		}

		if st.HasRealStockCount && diff.Abs().GreaterThan(decimal.NewFromFloat(5)) {
			systemFloat := math.Abs(decimalFloat(systemUnits))
			diffFloat := math.Abs(decimalFloat(diff))
			if systemFloat < 1 || diffFloat/systemFloat >= 0.2 || diffFloat >= 20 {
				impact := diffFloat
				out = append(out, decisionCardDraft{
					Fingerprint:    fmt.Sprintf("stock:difference:%d:%d", *workspace.ProjectID, st.Supply.ID),
					Domain:         "stock",
					RouteHint:      "stock",
					Severity:       "warning",
					Bucket:         decisionBucketImportant,
					Title:          "Diferencia relevante de stock en " + supplyName,
					Summary:        fmt.Sprintf("El stock de campo difiere del sistema en %s %s.", formatDecimal(diff), unit),
					Recommendation: "Validar el conteo físico y los consumos imputados; si corresponde, pedir aprobación para un borrador de conteo.",
					ImpactLabel:    "Diferencia: " + formatDecimal(diff) + " " + unit,
					ImpactValue:    &impact,
					Source:         "ponti.stock.summary",
					Evidence:       stockEvidence(workspace, st, "stock_difference"),
					Tools:          decisionTools("ponti.stock.summary"),
					Action:         stockCountAction(workspace, st, "Corregir diferencia de stock con conteo aprobado"),
				})
			}
			continue
		}

		if !st.HasRealStockCount && systemUnits.GreaterThan(decimal.Zero) && missingCounts < 5 {
			missingCounts++
			impact := decimalFloat(systemUnits)
			out = append(out, decisionCardDraft{
				Fingerprint:    fmt.Sprintf("stock:no_real_count:%d:%d", *workspace.ProjectID, st.Supply.ID),
				Domain:         "stock",
				RouteHint:      "stock",
				Severity:       "info",
				Bucket:         decisionBucketFollowUp,
				Title:          "Falta conteo de campo para " + supplyName,
				Summary:        fmt.Sprintf("Ponti tiene stock teórico de %s %s, pero no hay conteo real cargado.", formatDecimal(systemUnits), unit),
				Recommendation: "Programar un conteo de campo antes de usar este insumo como base de decisiones.",
				ImpactLabel:    "Stock teórico: " + formatDecimal(systemUnits) + " " + unit,
				ImpactValue:    &impact,
				Source:         "ponti.stock.summary",
				Evidence:       stockEvidence(workspace, st, "missing_real_count"),
				Tools:          decisionTools("ponti.stock.summary"),
				Action:         stockCountAction(workspace, st, "Cargar conteo de campo aprobado"),
			})
		}
	}
	return out, nil
}

func (s *DecisionService) generateInsightCards(ctx context.Context, tenantID uuid.UUID, actor string, workspace workspaceRequest) ([]decisionCardDraft, error) {
	if s.insights == nil {
		return []decisionCardDraft{}, nil
	}
	views, err := s.insights.ListByTenantForUser(ctx, tenantID.String(), actor, businessinsights.ListOptions{Limit: 100})
	if err != nil {
		return nil, err
	}
	out := make([]decisionCardDraft, 0, len(views))
	for _, view := range views {
		if strings.EqualFold(view.Status, "resolved") || !candidateMatchesWorkspace(view, workspace) {
			continue
		}
		severity := normalizeDecisionSeverity(view.Severity)
		bucket := bucketForSeverity(severity)
		action := insightResolutionAction(workspace, view)
		out = append(out, decisionCardDraft{
			Fingerprint:    "insight:" + view.ID,
			Domain:         "insights",
			RouteHint:      routeForInsight(view),
			Severity:       severity,
			Bucket:         bucket,
			Title:          view.Title,
			Summary:        view.Body,
			Recommendation: recommendationForInsight(view),
			ImpactLabel:    impactForInsight(view),
			Source:         "ponti.insights.list",
			Evidence: map[string]any{
				"source":           "ponti.insights.list",
				"workspace":        decisionWorkspaceMap(workspace),
				"captured_at":      time.Now().UTC().Format(time.RFC3339),
				"insight_id":       view.ID,
				"event_type":       view.EventType,
				"entity_type":      view.EntityType,
				"entity_id":        view.EntityID,
				"occurrence_count": view.OccurrenceCount,
				"first_seen_at":    view.FirstSeenAt.Format(time.RFC3339),
				"last_seen_at":     view.LastSeenAt.Format(time.RFC3339),
				"raw":              view.Evidence,
			},
			Tools:  decisionTools("ponti.insights.list", "ponti.insights.explain"),
			Action: action,
		})
	}
	return out, nil
}

func (s *DecisionService) generateReportCards(ctx context.Context, workspace workspaceRequest) ([]decisionCardDraft, error) {
	if s.report == nil {
		return []decisionCardDraft{}, nil
	}
	report, err := s.report.GetSummaryResultsReport(ctx, reportdomain.SummaryResultsFilter{
		ProjectID:  workspace.ProjectID,
		CustomerID: workspace.CustomerID,
		CampaignID: workspace.CampaignID,
		FieldID:    workspace.FieldID,
	})
	if err != nil {
		return nil, err
	}
	if report == nil || report.ProjectID <= 0 {
		return []decisionCardDraft{}, nil
	}
	out := []decisionCardDraft{}
	if report.Totals.TotalOperatingResultUsd.LessThan(decimal.Zero) {
		impact := decimalAbsFloat(report.Totals.TotalOperatingResultUsd)
		out = append(out, decisionCardDraft{
			Fingerprint:    fmt.Sprintf("reports:operating_result:%d", report.ProjectID),
			Domain:         "reports",
			RouteHint:      "reports",
			Severity:       "warning",
			Bucket:         decisionBucketImportant,
			Title:          "Resultado operativo negativo",
			Summary:        "El resultado operativo total del proyecto está negativo.",
			Recommendation: "Revisar cultivos con margen bajo, costos directos, renta y estructura antes de cerrar decisiones de campaña.",
			ImpactLabel:    "Resultado: USD " + formatDecimal(report.Totals.TotalOperatingResultUsd),
			ImpactValue:    &impact,
			Source:         "ponti.reports.summary_results.summary",
			Evidence:       reportEvidence(workspace, report, nil, "operating_result_negative"),
			Tools:          decisionTools("ponti.reports.summary_results.summary"),
			Action:         explainAction("explain_report_result", "Explicar resultado", "ponti.reports.summary_results.summary", workspace),
		})
	}
	negativeCrops := 0
	for _, crop := range report.Crops {
		if !crop.OperatingResultUsd.LessThan(decimal.Zero) || negativeCrops >= 3 {
			continue
		}
		negativeCrops++
		impact := decimalAbsFloat(crop.OperatingResultUsd)
		out = append(out, decisionCardDraft{
			Fingerprint:    fmt.Sprintf("reports:crop_margin:%d:%d", report.ProjectID, crop.CropID),
			Domain:         "reports",
			RouteHint:      "reports",
			Severity:       "warning",
			Bucket:         decisionBucketImportant,
			Title:          "Margen bajo en " + nonEmptyString(crop.CropName, "cultivo"),
			Summary:        fmt.Sprintf("%s muestra resultado operativo negativo en el resumen de campaña.", nonEmptyString(crop.CropName, "El cultivo")),
			Recommendation: "Analizar costos por hectárea, ingresos netos y superficie antes de repetir el paquete operativo.",
			ImpactLabel:    "Resultado cultivo: USD " + formatDecimal(crop.OperatingResultUsd),
			ImpactValue:    &impact,
			Source:         "ponti.reports.summary_results.summary",
			Evidence:       reportEvidence(workspace, report, &crop, "crop_margin_negative"),
			Tools:          decisionTools("ponti.reports.summary_results.summary"),
			Action:         explainAction("explain_crop_margin", "Investigar cultivo", "ponti.reports.summary_results.summary", workspace),
		})
	}
	return out, nil
}

func (s *DecisionService) generateSupplyQualityCards(ctx context.Context, workspace workspaceRequest) ([]decisionCardDraft, error) {
	if s.repo == nil {
		return []decisionCardDraft{}, nil
	}
	items, err := s.repo.listTentativeSupplies(ctx, workspace, 10)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 || workspace.ProjectID == nil {
		return []decisionCardDraft{}, nil
	}
	samples := make([]map[string]any, 0, len(items))
	for _, item := range items {
		samples = append(samples, map[string]any{
			"supply_id":        item.ID,
			"name":             item.Name,
			"price":            item.Price.StringFixed(2),
			"is_partial_price": item.IsPartialPrice,
			"is_pending":       item.IsPending,
		})
	}
	impact := float64(len(items))
	return []decisionCardDraft{{
		Fingerprint:    fmt.Sprintf("supplies:tentative_prices:%d", *workspace.ProjectID),
		Domain:         "supplies",
		RouteHint:      "supplies",
		Severity:       "warning",
		Bucket:         decisionBucketImportant,
		Title:          "Insumos con precios tentativos o pendientes",
		Summary:        fmt.Sprintf("Hay %d insumos que pueden distorsionar costos y márgenes.", len(items)),
		Recommendation: "Validar precios pendientes antes de tomar decisiones de margen, stock o nuevas labores.",
		ImpactLabel:    fmt.Sprintf("%d insumos a validar", len(items)),
		ImpactValue:    &impact,
		Source:         "ponti.supplies.summary",
		Evidence: map[string]any{
			"source":      "ponti.supplies.summary",
			"workspace":   decisionWorkspaceMap(workspace),
			"filters":     map[string]any{"mode": "pending"},
			"captured_at": time.Now().UTC().Format(time.RFC3339),
			"items":       samples,
			"summary":     map[string]any{"tentative_or_pending_count": len(items)},
		},
		Tools:  decisionTools("ponti.supplies.summary"),
		Action: explainAction("review_supplies", "Revisar insumos", "ponti.supplies.summary", workspace),
	}}, nil
}

func (s *DecisionService) generateLotCards(ctx context.Context, workspace workspaceRequest) ([]decisionCardDraft, error) {
	if s.repo == nil || workspace.ProjectID == nil {
		return []decisionCardDraft{}, nil
	}
	items, err := s.repo.listLotRiskSignals(ctx, workspace, 5)
	if err != nil {
		return nil, err
	}
	out := make([]decisionCardDraft, 0, len(items))
	for _, item := range items {
		signal := "lot_operating_result_negative"
		severity := "warning"
		bucket := decisionBucketImportant
		title := "Resultado bajo en " + nonEmptyString(item.LotName, "lote")
		summary := fmt.Sprintf("%s muestra resultado operativo por hectárea negativo.", nonEmptyString(item.LotName, "El lote"))
		recommendation := "Revisar costos por hectárea, cultivo actual y superficie antes de repetir el paquete operativo en este lote."
		impact := decimalAbsFloat(item.OperatingResultPerHa)
		impactLabel := "Resultado/ha: USD " + formatDecimal(item.OperatingResultPerHa)
		fingerprintKind := "negative_result"
		if !item.OperatingResultPerHa.LessThan(decimal.Zero) && item.SowedArea.GreaterThan(decimal.Zero) && item.HarvestedArea.IsZero() {
			signal = "lot_without_harvest"
			severity = "info"
			bucket = decisionBucketFollowUp
			title = "Lote sembrado sin cosecha cargada: " + nonEmptyString(item.LotName, "lote")
			summary = fmt.Sprintf("%s tiene superficie sembrada pero no registra cosecha en Ponti.", nonEmptyString(item.LotName, "El lote"))
			recommendation = "Confirmar si falta cargar cosecha, si el lote sigue en curso o si debe quedar marcado como seguimiento."
			impact = decimalFloat(item.SowedArea)
			impactLabel = "Superficie sembrada: " + formatDecimal(item.SowedArea) + " ha"
			fingerprintKind = "missing_harvest"
		}
		out = append(out, decisionCardDraft{
			Fingerprint:    fmt.Sprintf("lots:%s:%d:%d", fingerprintKind, *workspace.ProjectID, item.ID),
			Domain:         "lots",
			RouteHint:      "lots",
			Severity:       severity,
			Bucket:         bucket,
			Title:          title,
			Summary:        summary,
			Recommendation: recommendation,
			ImpactLabel:    impactLabel,
			ImpactValue:    &impact,
			Source:         "ponti.lots.summary",
			Evidence:       lotEvidence(workspace, item, signal),
			Tools:          decisionTools("ponti.lots.summary"),
			Action:         explainAction("review_lot", "Investigar lote", "ponti.lots.summary", workspace),
		})
	}
	return out, nil
}

func (s *DecisionService) generateOperationCards(ctx context.Context, workspace workspaceRequest) ([]decisionCardDraft, error) {
	if s.repo == nil || workspace.ProjectID == nil {
		return []decisionCardDraft{}, nil
	}
	signal, err := s.repo.countDraftWorkOrders(ctx, workspace)
	if err != nil {
		return nil, err
	}
	if signal.Count == 0 {
		return []decisionCardDraft{}, nil
	}
	impact := float64(signal.Count)
	return []decisionCardDraft{{
		Fingerprint:    fmt.Sprintf("workorders:drafts_pending:%d:%s", *workspace.ProjectID, optionalIntFingerprint(workspace.FieldID)),
		Domain:         "workorders",
		RouteHint:      "labors",
		Severity:       "info",
		Bucket:         decisionBucketFollowUp,
		Title:          "Borradores de órdenes pendientes",
		Summary:        fmt.Sprintf("Hay %d borradores digitales de orden de trabajo sin publicar.", signal.Count),
		Recommendation: "Revisar si esos borradores siguen vigentes antes de crear nuevas órdenes o recalcular costos.",
		ImpactLabel:    fmt.Sprintf("%d borradores", signal.Count),
		ImpactValue:    &impact,
		Source:         "ponti.workorders.list",
		Evidence: map[string]any{
			"source":      "ponti.workorders.list",
			"workspace":   decisionWorkspaceMap(workspace),
			"filters":     map[string]any{"status": "draft"},
			"captured_at": time.Now().UTC().Format(time.RFC3339),
			"summary":     map[string]any{"draft_count": signal.Count},
		},
		Tools:  decisionTools("ponti.workorders.list"),
		Action: explainAction("review_workorder_drafts", "Revisar borradores", "ponti.workorders.list", workspace),
	}}, nil
}

func validateDecisionWorkspace(w workspaceRequest) error {
	if w.CustomerID == nil || *w.CustomerID <= 0 || w.ProjectID == nil || *w.ProjectID <= 0 || w.CampaignID == nil || *w.CampaignID <= 0 {
		return domainerr.Validation("workspace.customer_id, workspace.project_id and workspace.campaign_id are required")
	}
	if w.FieldID != nil && *w.FieldID <= 0 {
		return domainerr.Validation("workspace.field_id must be greater than zero")
	}
	return nil
}

func decisionWorkspaceMap(w workspaceRequest) map[string]any {
	return w.toMap()
}

func stockEvidence(workspace workspaceRequest, st *stockdomain.Stock, signal string) map[string]any {
	evidence := map[string]any{
		"source":      "ponti.stock.summary",
		"workspace":   decisionWorkspaceMap(workspace),
		"captured_at": time.Now().UTC().Format(time.RFC3339),
		"signal":      signal,
	}
	if st == nil || st.Supply == nil {
		return evidence
	}
	evidence["items"] = []any{map[string]any{
		"stock_id":             st.ID,
		"supply_id":            st.Supply.ID,
		"supply_name":          st.Supply.Name,
		"unit":                 st.GetSupplyUnitName(),
		"system_stock_units":   st.GetStockUnits().StringFixed(2),
		"real_stock_units":     st.RealStockUnits.StringFixed(2),
		"stock_difference":     st.GetStockDifference().StringFixed(2),
		"has_real_stock_count": st.HasRealStockCount,
		"updated_at":           st.UpdatedAt.Format(time.RFC3339),
	}}
	evidence["summary"] = map[string]any{
		"system_stock_units":   st.GetStockUnits().StringFixed(2),
		"real_stock_units":     st.RealStockUnits.StringFixed(2),
		"stock_difference":     st.GetStockDifference().StringFixed(2),
		"has_real_stock_count": st.HasRealStockCount,
	}
	return evidence
}

func reportEvidence(workspace workspaceRequest, report *reportdomain.SummaryResultsResponse, crop *reportdomain.SummaryResults, signal string) map[string]any {
	evidence := map[string]any{
		"source":      "ponti.reports.summary_results.summary",
		"workspace":   decisionWorkspaceMap(workspace),
		"captured_at": time.Now().UTC().Format(time.RFC3339),
		"signal":      signal,
	}
	if report == nil {
		return evidence
	}
	evidence["summary"] = map[string]any{
		"project_id":                 report.ProjectID,
		"total_operating_result_usd": report.Totals.TotalOperatingResultUsd.StringFixed(2),
		"project_return_pct":         report.Totals.ProjectReturnPct.StringFixed(2),
		"total_invested_project_usd": report.Totals.TotalInvestedProjectUsd.StringFixed(2),
		"total_direct_costs_usd":     report.Totals.TotalDirectCostsUsd.StringFixed(2),
		"total_surface_ha":           report.Totals.TotalSurfaceHa.StringFixed(2),
	}
	if crop != nil {
		evidence["items"] = []any{map[string]any{
			"crop_id":              crop.CropID,
			"crop_name":            crop.CropName,
			"surface_ha":           crop.SurfaceHa.StringFixed(2),
			"operating_result_usd": crop.OperatingResultUsd.StringFixed(2),
			"crop_return_pct":      crop.CropReturnPct.StringFixed(2),
			"total_invested_usd":   crop.TotalInvestedUsd.StringFixed(2),
			"direct_costs_usd":     crop.DirectCostsUsd.StringFixed(2),
		}}
	}
	return evidence
}

func lotEvidence(workspace workspaceRequest, item decisionLotRiskSignal, signal string) map[string]any {
	return map[string]any{
		"source":      "ponti.lots.summary",
		"workspace":   decisionWorkspaceMap(workspace),
		"filters":     map[string]any{"field_id": workspace.FieldID},
		"captured_at": time.Now().UTC().Format(time.RFC3339),
		"signal":      signal,
		"items": []any{map[string]any{
			"lot_id":                  item.ID,
			"lot_name":                item.LotName,
			"field_id":                item.FieldID,
			"field_name":              item.FieldName,
			"current_crop_id":         item.CurrentCropID,
			"current_crop":            item.CurrentCrop,
			"sowed_area_ha":           item.SowedArea.StringFixed(2),
			"harvested_area_ha":       item.HarvestedArea.StringFixed(2),
			"cost_usd_per_ha":         item.CostUSDPerHa.StringFixed(2),
			"yield_tn_per_ha":         item.YieldTnPerHa.StringFixed(2),
			"operating_result_per_ha": item.OperatingResultPerHa.StringFixed(2),
		}},
		"summary": map[string]any{
			"lot_id":                  item.ID,
			"operating_result_per_ha": item.OperatingResultPerHa.StringFixed(2),
			"sowed_area_ha":           item.SowedArea.StringFixed(2),
			"harvested_area_ha":       item.HarvestedArea.StringFixed(2),
		},
		"warnings": []any{},
	}
}

func stockCountAction(workspace workspaceRequest, st *stockdomain.Stock, reason string) map[string]any {
	payload := map[string]any{
		"project_id": *workspace.ProjectID,
		"supply_id":  st.Supply.ID,
		"reason":     reason,
		"workspace":  decisionWorkspaceMap(workspace),
	}
	missingInputs := []string{}
	if st.ID > 0 {
		payload["stock_id"] = st.ID
	}
	if st.HasRealStockCount {
		payload["real_stock_units"] = decimalFloat(st.RealStockUnits)
	} else {
		missingInputs = append(missingInputs, "real_stock_units")
	}
	return map[string]any{
		"id":                "create_stock_count_draft",
		"label":             "Pedir aprobación para conteo",
		"capability_id":     "ponti.stock_count.draft",
		"requires_approval": true,
		"nexus_action_type": pontiNexusActionType,
		"payload":           payload,
		"missing_inputs":    missingInputs,
	}
}

func insightResolutionAction(workspace workspaceRequest, view businessinsights.CandidateView) map[string]any {
	return map[string]any{
		"id":                "create_insight_resolution_draft",
		"label":             "Preparar resolución",
		"capability_id":     "ponti.insight_resolution.draft",
		"requires_approval": true,
		"nexus_action_type": pontiNexusActionType,
		"payload": map[string]any{
			"insight_id":      view.ID,
			"resolution_note": "Preparar resolución reversible con evidencia del insight.",
			"workspace":       decisionWorkspaceMap(workspace),
		},
	}
}

func explainAction(id, label, capability string, workspace workspaceRequest) map[string]any {
	return map[string]any{
		"id":                id,
		"label":             label,
		"capability_id":     capability,
		"requires_approval": false,
		"payload":           map[string]any{"workspace": decisionWorkspaceMap(workspace)},
	}
}

func decisionTools(names ...string) []any {
	out := make([]any, 0, len(names))
	for _, name := range names {
		if strings.TrimSpace(name) == "" {
			continue
		}
		out = append(out, map[string]any{
			"name":   name,
			"status": "success",
			"source": "ponti-core",
		})
	}
	return out
}

func normalizeDecisionSeverity(in string) string {
	switch strings.ToLower(strings.TrimSpace(in)) {
	case "critical", "critico", "crítico", "error":
		return "critical"
	case "warning", "warn", "medium":
		return "warning"
	case "success", "opportunity":
		return "opportunity"
	default:
		return "info"
	}
}

func bucketForSeverity(severity string) string {
	switch normalizeDecisionSeverity(severity) {
	case "critical":
		return decisionBucketUrgent
	case "warning":
		return decisionBucketImportant
	case "opportunity":
		return decisionBucketOpportunity
	default:
		return decisionBucketFollowUp
	}
}

func normalizeDecisionBucket(in string) string {
	switch strings.ToLower(strings.TrimSpace(in)) {
	case decisionBucketUrgent:
		return decisionBucketUrgent
	case decisionBucketImportant:
		return decisionBucketImportant
	case decisionBucketOpportunity:
		return decisionBucketOpportunity
	case decisionBucketFollowUp, "followup", "follow-up", "seguimiento":
		return decisionBucketFollowUp
	default:
		return ""
	}
}

func routeForInsight(view businessinsights.CandidateView) string {
	event := strings.ToLower(view.EventType)
	switch {
	case strings.Contains(event, "stock"):
		return "stock"
	case strings.Contains(event, "report"), strings.Contains(event, "operating_result"):
		return "reports"
	case strings.Contains(event, "price"), strings.Contains(event, "supply"):
		return "supplies"
	default:
		return "dashboard"
	}
}

func recommendationForInsight(view businessinsights.CandidateView) string {
	event := strings.ToLower(view.EventType)
	switch {
	case strings.Contains(event, "stock"):
		return "Investigar la causa y preparar un borrador reversible si hace falta regularizar el dato."
	case strings.Contains(event, "operating_result"):
		return "Analizar el reporte económico y priorizar cultivos/costos que explican el desvío."
	case strings.Contains(event, "tentative"):
		return "Validar datos pendientes antes de usar informes o costos como base de decisión."
	default:
		return "Explicar el insight en el asistente y preparar una resolución reversible si corresponde."
	}
}

func impactForInsight(view businessinsights.CandidateView) string {
	if view.OccurrenceCount > 1 {
		return fmt.Sprintf("%d recurrencias", view.OccurrenceCount)
	}
	return nonEmptyString(view.Severity, "insight")
}

func candidateMatchesWorkspace(view businessinsights.CandidateView, workspace workspaceRequest) bool {
	if workspace.ProjectID == nil || *workspace.ProjectID <= 0 {
		return true
	}
	want := strconv.FormatInt(*workspace.ProjectID, 10)
	if strings.EqualFold(view.EntityType, "project") && strings.TrimSpace(view.EntityID) == want {
		return true
	}
	if project, ok := stringFromEvidence(view.Evidence, "project_id"); ok && project == want {
		return true
	}
	if rawWorkspace, ok := view.Evidence["workspace"].(map[string]any); ok {
		if project, ok := stringFromEvidence(rawWorkspace, "project_id"); ok && project == want {
			return true
		}
	}
	return false
}

func stringFromEvidence(e map[string]any, key string) (string, bool) {
	if e == nil {
		return "", false
	}
	out := strings.TrimSpace(stringFromAny(e[key]))
	return out, out != ""
}

func stringFromAny(v any) string {
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case fmt.Stringer:
		return strings.TrimSpace(t.String())
	case int:
		return strconv.Itoa(t)
	case int64:
		return strconv.FormatInt(t, 10)
	case float64:
		if math.Trunc(t) == t {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	default:
		return ""
	}
}

func decimalAbsFloat(d decimal.Decimal) float64 {
	return math.Abs(decimalFloat(d))
}

func formatDecimal(d decimal.Decimal) string {
	return d.StringFixed(2)
}

func decimalFloat(d decimal.Decimal) float64 {
	value, _ := d.Float64()
	return value
}

func nonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func optionalIntFingerprint(v *int64) string {
	if v == nil {
		return "all"
	}
	return strconv.FormatInt(*v, 10)
}
