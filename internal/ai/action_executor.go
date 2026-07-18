package ai

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	ctxkeys "github.com/devpablocristo/platform/security/go/contextkeys"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	stockdomain "github.com/devpablocristo/ponti-backend/internal/stock/usecases/domain"
)

const (
	actionDraftTypeInsightResolution = "insight_resolution"
	actionDraftTypeStockCount        = "stock_count"

	actionDraftStatusStaged  = "staged"
	actionDraftStatusApplied = "applied"

	// actionExecutorSystemActor identifica los writes aplicados por el agente
	// cuando no hay un actor humano trazable (ej: callback de Nexus sin
	// decided_by).
	actionExecutorSystemActor = "ai:governed-executor"
)

// InsightResolverPort es el subset de businessinsights.Service que aplica una
// resolución de insight.
type InsightResolverPort interface {
	ResolveManual(ctx context.Context, tenantID uuid.UUID, candidateID, actor string) error
}

// StockWriterPort es el subset de stock.UseCases que aplica un conteo de stock
// real (misma semántica que PUT /stocks/real-stock/:id).
type StockWriterPort interface {
	GetStockByID(ctx context.Context, stockID int64) (*stockdomain.Stock, error)
	GetLastStockByProjectID(ctx context.Context, projectID int64, supplyID int64) (*stockdomain.Stock, bool, error)
	UpdateRealStockUnits(ctx context.Context, stockID int64, stock *stockdomain.Stock) error
}

// ActionVerifierPort abstrae el governance.Verifier: confirma contra Nexus que
// un request id está aprobado para el action type esperado.
type ActionVerifierPort interface {
	VerifyApproved(ctx context.Context, tenantID uuid.UUID, nexusRequestID, expectedActionType string) error
}

// ActionExecutorConfig agrupa los flags del executor (subset de config.Nexus).
type ActionExecutorConfig struct {
	GovernedWritesEnabled bool
}

// ActionDraftResult describe el estado del draft luego de un Apply*.
type ActionDraftResult struct {
	DraftID uuid.UUID
	Status  string
	Applied bool
}

// ActionExecutor materializa las acciones gobernadas del agente: siempre
// persiste el draft (staged) y solo ejecuta el write real cuando
// AI_GOVERNED_WRITES_ENABLED está activo y el nexus request id verifica como
// aprobado (fail closed: cualquier otra combinación preserva la semántica
// preview y deja el draft staged).
type ActionExecutor struct {
	drafts   *actionDraftRepository
	insights InsightResolverPort
	stocks   StockWriterPort
	verifier ActionVerifierPort
	cfg      ActionExecutorConfig
}

func NewActionExecutor(drafts *actionDraftRepository, insights InsightResolverPort, stocks StockWriterPort, verifier ActionVerifierPort, cfg ActionExecutorConfig) *ActionExecutor {
	return &ActionExecutor{drafts: drafts, insights: insights, stocks: stocks, verifier: verifier, cfg: cfg}
}

// InsightResolutionInput es el payload de un draft de resolución de insight.
type InsightResolutionInput struct {
	InsightID      string
	ResolutionNote string
	Workspace      map[string]any
}

// StockCountInput es el payload de un draft de conteo de stock real.
type StockCountInput struct {
	ProjectID      int64
	StockID        *int64
	SupplyID       int64
	RealStockUnits float64
	Reason         string
	Workspace      map[string]any
}

// ApplyInsightResolution stagea el draft y, si el write está habilitado y
// verificado, resuelve el insight vía businessinsights y marca el draft
// applied. Idempotente por (tenant, draft_type, nexus_request_id).
func (e *ActionExecutor) ApplyInsightResolution(ctx context.Context, tenantID uuid.UUID, actor, nexusRequestID string, in InsightResolutionInput) (ActionDraftResult, error) {
	if e == nil || e.drafts == nil {
		return ActionDraftResult{}, domainerr.Internal("ai action executor unavailable")
	}
	in.InsightID = strings.TrimSpace(in.InsightID)
	in.ResolutionNote = strings.TrimSpace(in.ResolutionNote)
	if in.InsightID == "" {
		return ActionDraftResult{}, domainerr.Validation("insight_id is required")
	}
	payload := map[string]any{
		"insight_id":      in.InsightID,
		"resolution_note": in.ResolutionNote,
	}
	if len(in.Workspace) > 0 {
		payload["workspace"] = in.Workspace
	}
	draft, err := e.stageDraft(ctx, tenantID, actionDraftTypeInsightResolution, nexusRequestID, actor, payload)
	if err != nil {
		return ActionDraftResult{}, err
	}
	if draft.Status == actionDraftStatusApplied {
		// Replay del mismo nexus request: el write ya se aplicó una vez.
		return actionDraftResult(draft), nil
	}
	if !e.canExecute(ctx, tenantID, nexusRequestID, pontiActionTypeInsightResolve) {
		return actionDraftResult(draft), nil
	}
	if e.insights == nil {
		return ActionDraftResult{}, domainerr.Unavailable("insight resolver not configured")
	}
	if err := e.insights.ResolveManual(ctx, tenantID, in.InsightID, executorActor(actor)); err != nil {
		return ActionDraftResult{}, err
	}
	return e.markApplied(ctx, draft, actor)
}

// ApplyStockCount stagea el draft y, si el write está habilitado y verificado,
// aplica el conteo real con la misma semántica que el handler HTTP de stock
// (punto único de mutación UpdateRealStockUnits) usando un actor de sistema.
func (e *ActionExecutor) ApplyStockCount(ctx context.Context, tenantID uuid.UUID, actor, nexusRequestID string, in StockCountInput) (ActionDraftResult, error) {
	if e == nil || e.drafts == nil {
		return ActionDraftResult{}, domainerr.Internal("ai action executor unavailable")
	}
	in.Reason = strings.TrimSpace(in.Reason)
	if in.ProjectID <= 0 {
		return ActionDraftResult{}, domainerr.Validation("project_id is required")
	}
	if in.SupplyID <= 0 {
		return ActionDraftResult{}, domainerr.Validation("supply_id is required")
	}
	if in.Reason == "" {
		return ActionDraftResult{}, domainerr.Validation("reason is required")
	}
	payload := map[string]any{
		"project_id":       in.ProjectID,
		"supply_id":        in.SupplyID,
		"real_stock_units": in.RealStockUnits,
		"reason":           in.Reason,
	}
	if in.StockID != nil && *in.StockID > 0 {
		payload["stock_id"] = *in.StockID
	}
	if len(in.Workspace) > 0 {
		payload["workspace"] = in.Workspace
	}
	draft, err := e.stageDraft(ctx, tenantID, actionDraftTypeStockCount, nexusRequestID, actor, payload)
	if err != nil {
		return ActionDraftResult{}, err
	}
	if draft.Status == actionDraftStatusApplied {
		return actionDraftResult(draft), nil
	}
	if !e.canExecute(ctx, tenantID, nexusRequestID, pontiActionTypeStockCountApply) {
		return actionDraftResult(draft), nil
	}
	if e.stocks == nil {
		return ActionDraftResult{}, domainerr.Unavailable("stock writer not configured")
	}
	if err := e.applyStockWrite(ctx, tenantID, actor, in); err != nil {
		return ActionDraftResult{}, err
	}
	return e.markApplied(ctx, draft, actor)
}

func (e *ActionExecutor) applyStockWrite(ctx context.Context, tenantID uuid.UUID, actor string, in StockCountInput) error {
	// El punto único de mutación de stock evalúa notificaciones reactivas a
	// partir del tenant/actor del contexto: inyectarlos para que el write del
	// agente dispare los mismos insights que el write humano.
	ctx = context.WithValue(ctx, ctxkeys.OrgID, tenantID)
	ctx = context.WithValue(ctx, ctxkeys.Actor, executorActor(actor))

	var stock *stockdomain.Stock
	var err error
	if in.StockID != nil && *in.StockID > 0 {
		stock, err = e.stocks.GetStockByID(ctx, *in.StockID)
		if err != nil {
			return err
		}
	} else {
		var found bool
		stock, found, err = e.stocks.GetLastStockByProjectID(ctx, in.ProjectID, in.SupplyID)
		if err != nil {
			return err
		}
		if !found || stock == nil {
			return domainerr.NotFound("stock not found for project and supply")
		}
	}
	if stock == nil || stock.Project == nil || stock.Project.ID != in.ProjectID {
		// Misma defensa que el handler HTTP: el stock debe ser del proyecto.
		return domainerr.NotFound("stock not found")
	}
	systemActor := executorActor(actor)
	stock.RealStockUnits = decimal.NewFromFloat(in.RealStockUnits)
	stock.HasRealStockCount = true
	stock.UpdatedBy = &systemActor
	stock.UpdatedAt = time.Time{}
	return e.stocks.UpdateRealStockUnits(ctx, stock.ID, stock)
}

// DispatchApproved implementa governance.ActionDispatcher: traduce el
// action_type aprobado por Nexus al Apply* correspondiente usando los params
// registrados en la request. Action types desconocidos devuelven error (el
// executor de governance los registra sin romper el callback).
func (e *ActionExecutor) DispatchApproved(ctx context.Context, tenantID uuid.UUID, actionType, nexusRequestID string, params map[string]any, actor string) (map[string]any, error) {
	switch actionType {
	case pontiActionTypeInsightResolve:
		in := InsightResolutionInput{
			InsightID:      stringFromAny(params["insight_id"]),
			ResolutionNote: stringFromAny(params["resolution_note"]),
			Workspace:      mapFromAny(params["workspace"]),
		}
		res, err := e.ApplyInsightResolution(ctx, tenantID, actor, nexusRequestID, in)
		if err != nil {
			return nil, err
		}
		if !res.Applied {
			return nil, fmt.Errorf("insight resolution draft %s was not applied (write disabled or unverified)", res.DraftID)
		}
		return map[string]any{
			"draft_id":   res.DraftID.String(),
			"draft_type": actionDraftTypeInsightResolution,
			"status":     res.Status,
			"insight_id": in.InsightID,
		}, nil
	case pontiActionTypeStockCountApply:
		in := StockCountInput{
			ProjectID:      int64FromAny(params["project_id"]),
			SupplyID:       int64FromAny(params["supply_id"]),
			RealStockUnits: floatFromAny(params["real_stock_units"]),
			Reason:         stringFromAny(params["reason"]),
			Workspace:      mapFromAny(params["workspace"]),
		}
		if stockID := int64FromAny(params["stock_id"]); stockID > 0 {
			in.StockID = &stockID
		}
		res, err := e.ApplyStockCount(ctx, tenantID, actor, nexusRequestID, in)
		if err != nil {
			return nil, err
		}
		if !res.Applied {
			return nil, fmt.Errorf("stock count draft %s was not applied (write disabled or unverified)", res.DraftID)
		}
		return map[string]any{
			"draft_id":         res.DraftID.String(),
			"draft_type":       actionDraftTypeStockCount,
			"status":           res.Status,
			"project_id":       in.ProjectID,
			"supply_id":        in.SupplyID,
			"real_stock_units": in.RealStockUnits,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported action type %q", actionType)
	}
}

// stageDraft persiste el draft staged; si ya existe uno para el mismo nexus
// request lo reutiliza (idempotencia ante retries del agente o del callback).
func (e *ActionExecutor) stageDraft(ctx context.Context, tenantID uuid.UUID, draftType, nexusRequestID, actor string, payload map[string]any) (actionDraftModel, error) {
	nexusRequestID = strings.TrimSpace(nexusRequestID)
	if nexusRequestID != "" {
		existing, err := e.drafts.getByNexusRequestID(ctx, tenantID, draftType, nexusRequestID)
		if err == nil {
			return existing, nil
		}
		if !domainerr.IsNotFound(err) {
			return actionDraftModel{}, err
		}
	}
	return e.drafts.create(ctx, actionDraftModel{
		TenantID:       tenantID,
		DraftType:      draftType,
		Status:         actionDraftStatusStaged,
		PayloadJSON:    marshalDecisionJSON(payload, "{}"),
		NexusRequestID: nexusRequestID,
		CreatedBy:      executorActor(actor),
	})
}

// canExecute decide si corresponde el write real: exige el flag
// AI_GOVERNED_WRITES_ENABLED y un nexus request id verificado como aprobado
// para el action type esperado. Cualquier otra combinación → preview (staged).
func (e *ActionExecutor) canExecute(ctx context.Context, tenantID uuid.UUID, nexusRequestID, expectedActionType string) bool {
	if !e.cfg.GovernedWritesEnabled {
		return false
	}
	nexusRequestID = strings.TrimSpace(nexusRequestID)
	if nexusRequestID == "" || e.verifier == nil {
		return false
	}
	if err := e.verifier.VerifyApproved(ctx, tenantID, nexusRequestID, expectedActionType); err != nil {
		log.Printf("[ai-actions] write bloqueado: verificación Nexus de %s falló: %v", nexusRequestID, err)
		return false
	}
	return true
}

func (e *ActionExecutor) markApplied(ctx context.Context, draft actionDraftModel, actor string) (ActionDraftResult, error) {
	now := time.Now().UTC()
	if err := e.drafts.markApplied(ctx, draft.TenantID, draft.ID, executorActor(actor), now); err != nil {
		return ActionDraftResult{}, err
	}
	draft.Status = actionDraftStatusApplied
	draft.AppliedAt = &now
	return actionDraftResult(draft), nil
}

func actionDraftResult(draft actionDraftModel) ActionDraftResult {
	return ActionDraftResult{
		DraftID: draft.ID,
		Status:  draft.Status,
		Applied: draft.Status == actionDraftStatusApplied,
	}
}

func executorActor(actor string) string {
	if strings.TrimSpace(actor) == "" {
		return actionExecutorSystemActor
	}
	return strings.TrimSpace(actor)
}

func mapFromAny(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return nil
}

func int64FromAny(v any) int64 {
	switch t := v.(type) {
	case int64:
		return t
	case int:
		return int64(t)
	case float64:
		return int64(t)
	default:
		return 0
	}
}

func floatFromAny(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case int64:
		return float64(t)
	case int:
		return float64(t)
	default:
		return 0
	}
}
