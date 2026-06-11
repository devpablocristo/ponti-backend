package ai

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/devpablocristo/ponti-backend/internal/governance"
	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
)

const (
	draftActionStatusPreview = "preview"

	// nexusRequestIDHeader trae el request id de Nexus que respalda la acción.
	nexusRequestIDHeader = "X-Nexus-Request-ID"
	// axisCompanionActor es el actor con el que el middleware de integración
	// identifica a Axis (axis_product_integration_auth).
	axisCompanionActor = "axis-companion"
	// executionBlockedByNexus es la causa de bloqueo expuesta en las 412.
	executionBlockedByNexus = "nexus_required"
)

type workspaceRequest struct {
	CustomerID *int64 `json:"customer_id,omitempty"`
	ProjectID  *int64 `json:"project_id,omitempty"`
	CampaignID *int64 `json:"campaign_id,omitempty"`
	FieldID    *int64 `json:"field_id,omitempty"`
}

type insightResolvePrepareRequest struct {
	InsightID      string           `json:"insight_id"`
	ResolutionNote string           `json:"resolution_note,omitempty"`
	Workspace      workspaceRequest `json:"workspace,omitempty"`
}

type workOrderDraftPrepareRequest struct {
	ProjectID     int64            `json:"project_id"`
	FieldID       *int64           `json:"field_id,omitempty"`
	CampaignID    *int64           `json:"campaign_id,omitempty"`
	WorkType      string           `json:"work_type"`
	ScheduledDate string           `json:"scheduled_date,omitempty"`
	Notes         string           `json:"notes,omitempty"`
	Workspace     workspaceRequest `json:"workspace,omitempty"`
}

type stockAdjustmentPrepareRequest struct {
	ProjectID     int64            `json:"project_id"`
	SupplyID      int64            `json:"supply_id"`
	QuantityDelta float64          `json:"quantity_delta"`
	Reason        string           `json:"reason"`
	Workspace     workspaceRequest `json:"workspace,omitempty"`
}

type insightResolutionDraftRequest struct {
	InsightID      string           `json:"insight_id"`
	ResolutionNote string           `json:"resolution_note,omitempty"`
	Workspace      workspaceRequest `json:"workspace,omitempty"`
}

type stockCountDraftRequest struct {
	ProjectID      int64            `json:"project_id"`
	StockID        *int64           `json:"stock_id,omitempty"`
	SupplyID       int64            `json:"supply_id"`
	RealStockUnits float64          `json:"real_stock_units"`
	Reason         string           `json:"reason"`
	Workspace      workspaceRequest `json:"workspace,omitempty"`
}

func (h *Handler) PrepareInsightResolve(c *gin.Context) {
	orgID, actor, err := requireActionContext(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	var req insightResolvePrepareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sharedhandlers.RespondError(c, domainerr.Validation("invalid request payload"))
		return
	}
	req.InsightID = strings.TrimSpace(req.InsightID)
	req.ResolutionNote = strings.TrimSpace(req.ResolutionNote)
	if req.InsightID == "" {
		sharedhandlers.RespondError(c, domainerr.Validation("insight_id is required"))
		return
	}
	if _, ok := h.requireNexusApproval(c, orgID, actor, pontiActionTypeInsightResolve); !ok {
		return
	}
	workspace := req.Workspace.toMap()
	proposal := map[string]any{
		"insight_id":      req.InsightID,
		"resolution_note": req.ResolutionNote,
		"preview_only":    true,
		"write_performed": false,
	}
	c.JSON(http.StatusOK, draftActionPreviewResponse("ponti.insight.resolve.prepare", pontiActionTypeInsightResolve, orgID, actor, workspace, proposal))
}

func (h *Handler) PrepareWorkOrderDraft(c *gin.Context) {
	orgID, actor, err := requireActionContext(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	var req workOrderDraftPrepareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sharedhandlers.RespondError(c, domainerr.Validation("invalid request payload"))
		return
	}
	req.WorkType = strings.TrimSpace(req.WorkType)
	req.ScheduledDate = strings.TrimSpace(req.ScheduledDate)
	req.Notes = strings.TrimSpace(req.Notes)
	if req.ProjectID <= 0 {
		sharedhandlers.RespondError(c, domainerr.Validation("project_id is required"))
		return
	}
	if req.WorkType == "" {
		sharedhandlers.RespondError(c, domainerr.Validation("work_type is required"))
		return
	}
	if req.ScheduledDate != "" {
		if _, err := time.Parse("2006-01-02", req.ScheduledDate); err != nil {
			sharedhandlers.RespondError(c, domainerr.Validation("scheduled_date must use YYYY-MM-DD"))
			return
		}
	}
	if _, ok := h.requireNexusApproval(c, orgID, actor, pontiActionTypeWorkOrderDraftCreate); !ok {
		return
	}
	workspace := req.Workspace.withFallbackProject(req.ProjectID).withFallbackField(req.FieldID).withFallbackCampaign(req.CampaignID).toMap()
	proposal := map[string]any{
		"project_id":      req.ProjectID,
		"field_id":        valueOrNil(req.FieldID),
		"campaign_id":     valueOrNil(req.CampaignID),
		"work_type":       req.WorkType,
		"scheduled_date":  req.ScheduledDate,
		"notes":           req.Notes,
		"preview_only":    true,
		"write_performed": false,
	}
	c.JSON(http.StatusOK, draftActionPreviewResponse("ponti.workorder.draft.prepare", pontiActionTypeWorkOrderDraftCreate, orgID, actor, workspace, proposal))
}

func (h *Handler) PrepareStockAdjustment(c *gin.Context) {
	orgID, actor, err := requireActionContext(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	var req stockAdjustmentPrepareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sharedhandlers.RespondError(c, domainerr.Validation("invalid request payload"))
		return
	}
	req.Reason = strings.TrimSpace(req.Reason)
	if req.ProjectID <= 0 {
		sharedhandlers.RespondError(c, domainerr.Validation("project_id is required"))
		return
	}
	if req.SupplyID <= 0 {
		sharedhandlers.RespondError(c, domainerr.Validation("supply_id is required"))
		return
	}
	if req.QuantityDelta == 0 {
		sharedhandlers.RespondError(c, domainerr.Validation("quantity_delta must be different from zero"))
		return
	}
	if req.Reason == "" {
		sharedhandlers.RespondError(c, domainerr.Validation("reason is required"))
		return
	}
	if _, ok := h.requireNexusApproval(c, orgID, actor, pontiActionTypeStockAdjust); !ok {
		return
	}
	workspace := req.Workspace.withFallbackProject(req.ProjectID).toMap()
	proposal := map[string]any{
		"project_id":      req.ProjectID,
		"supply_id":       req.SupplyID,
		"quantity_delta":  req.QuantityDelta,
		"reason":          req.Reason,
		"preview_only":    true,
		"write_performed": false,
	}
	c.JSON(http.StatusOK, draftActionPreviewResponse("ponti.stock_adjustment.prepare", pontiActionTypeStockAdjust, orgID, actor, workspace, proposal))
}

func (h *Handler) DraftInsightResolution(c *gin.Context) {
	orgID, actor, err := requireActionContext(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	var req insightResolutionDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sharedhandlers.RespondError(c, domainerr.Validation("invalid request payload"))
		return
	}
	req.InsightID = strings.TrimSpace(req.InsightID)
	req.ResolutionNote = strings.TrimSpace(req.ResolutionNote)
	if req.InsightID == "" {
		sharedhandlers.RespondError(c, domainerr.Validation("insight_id is required"))
		return
	}
	nexusRequestID, ok := h.requireNexusApproval(c, orgID, actor, pontiActionTypeInsightResolve)
	if !ok {
		return
	}
	workspace := req.Workspace.toMap()
	draftID := any("insight-resolution-" + uuid.NewString())
	writePerformed := false
	if h.actions != nil {
		result, aerr := h.actions.ApplyInsightResolution(c.Request.Context(), orgID, actor, nexusRequestID, InsightResolutionInput{
			InsightID:      req.InsightID,
			ResolutionNote: req.ResolutionNote,
			Workspace:      workspace,
		})
		if aerr != nil {
			sharedhandlers.RespondError(c, aerr)
			return
		}
		draftID = "insight-resolution-" + result.DraftID.String()
		writePerformed = result.Applied
	}
	c.JSON(http.StatusOK, draftActionExecutionResponse(
		"ponti.insight_resolution.draft",
		pontiActionTypeInsightResolve,
		orgID,
		actor,
		nexusRequestID,
		draftID,
		writePerformed,
		draftExecutionStatus(writePerformed),
		workspace,
		map[string]any{
			"insight_id":      req.InsightID,
			"resolution_note": req.ResolutionNote,
		},
	))
}

func (h *Handler) DraftStockCount(c *gin.Context) {
	orgID, actor, err := requireActionContext(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	var req stockCountDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sharedhandlers.RespondError(c, domainerr.Validation("invalid request payload"))
		return
	}
	req.Reason = strings.TrimSpace(req.Reason)
	if req.ProjectID <= 0 {
		sharedhandlers.RespondError(c, domainerr.Validation("project_id is required"))
		return
	}
	if req.SupplyID <= 0 {
		sharedhandlers.RespondError(c, domainerr.Validation("supply_id is required"))
		return
	}
	if req.Reason == "" {
		sharedhandlers.RespondError(c, domainerr.Validation("reason is required"))
		return
	}
	nexusRequestID, ok := h.requireNexusApproval(c, orgID, actor, pontiActionTypeStockCountApply)
	if !ok {
		return
	}
	workspace := req.Workspace.withFallbackProject(req.ProjectID).toMap()
	draftID := any("stock-count-" + uuid.NewString())
	writePerformed := false
	if h.actions != nil {
		result, aerr := h.actions.ApplyStockCount(c.Request.Context(), orgID, actor, nexusRequestID, StockCountInput{
			ProjectID:      req.ProjectID,
			StockID:        req.StockID,
			SupplyID:       req.SupplyID,
			RealStockUnits: req.RealStockUnits,
			Reason:         req.Reason,
			Workspace:      workspace,
		})
		if aerr != nil {
			sharedhandlers.RespondError(c, aerr)
			return
		}
		draftID = "stock-count-" + result.DraftID.String()
		writePerformed = result.Applied
	}
	c.JSON(http.StatusOK, draftActionExecutionResponse(
		"ponti.stock_count.draft",
		pontiActionTypeStockCountApply,
		orgID,
		actor,
		nexusRequestID,
		draftID,
		writePerformed,
		draftExecutionStatus(writePerformed),
		workspace,
		map[string]any{
			"project_id":       req.ProjectID,
			"stock_id":         valueOrNil(req.StockID),
			"supply_id":        req.SupplyID,
			"real_stock_units": req.RealStockUnits,
			"reason":           req.Reason,
		},
	))
}

func requireActionContext(c *gin.Context) (uuid.UUID, string, error) {
	orgID, err := sharedhandlers.ParseOrgID(c)
	if err != nil {
		return uuid.Nil, "", err
	}
	actor, err := sharedhandlers.ParseActor(c)
	if err != nil {
		return uuid.Nil, "", err
	}
	return orgID, actor, nil
}

// requireNexusApproval aplica el gate de governance sobre una acción gobernada.
// Con GOVERNANCE_VERIFY_NEXUS=true y caller Axis (actor axis-companion) exige
// X-Nexus-Request-ID verificado como aprobado en Nexus: faltante o inválido →
// 412 {execution_blocked_by: nexus_required}; Nexus inalcanzable → 502 (fail
// closed). Con el flag apagado mantiene el comportamiento previo y solo
// loggea la ausencia del header. Devuelve el request id y si puede continuar.
func (h *Handler) requireNexusApproval(c *gin.Context, orgID uuid.UUID, actor, expectedActionType string) (string, bool) {
	nexusRequestID := strings.TrimSpace(c.GetHeader(nexusRequestIDHeader))
	if !h.governedCfg.VerifyNexus || actor != axisCompanionActor {
		if nexusRequestID == "" {
			log.Printf("[ai-actions] WARN: %s sin %s (actor=%s, GOVERNANCE_VERIFY_NEXUS=%t)", expectedActionType, nexusRequestIDHeader, actor, h.governedCfg.VerifyNexus)
		}
		return nexusRequestID, true
	}
	if nexusRequestID == "" {
		respondNexusBlocked(c, nexusRequestIDHeader+" header is required")
		return "", false
	}
	if h.verifier == nil {
		// Fail closed: enforcement activo sin verifier configurado.
		sharedhandlers.RespondError(c, domainerr.Unavailable("nexus verifier not configured"))
		return "", false
	}
	if err := h.verifier.VerifyApproved(c.Request.Context(), orgID, nexusRequestID, expectedActionType); err != nil {
		if governance.IsNotApproved(err) {
			respondNexusBlocked(c, err.Error())
			return "", false
		}
		sharedhandlers.RespondError(c, err)
		return "", false
	}
	return nexusRequestID, true
}

func respondNexusBlocked(c *gin.Context, detail string) {
	c.JSON(http.StatusPreconditionFailed, gin.H{
		"execution_blocked_by": executionBlockedByNexus,
		"detail":               detail,
	})
}

func draftExecutionStatus(writePerformed bool) string {
	if writePerformed {
		return "applied"
	}
	return "draft_staged"
}

func draftActionPreviewResponse(action, actionType string, orgID uuid.UUID, actor string, workspace map[string]any, proposal map[string]any) map[string]any {
	return map[string]any{
		"status":               draftActionStatusPreview,
		"action":               action,
		"approval_required":    true,
		"nexus_action_type":    actionType,
		"side_effect_type":     "write",
		"preview_only":         true,
		"write_performed":      false,
		"proposal":             proposal,
		"pending_execution":    true,
		"execution_allowed":    false,
		"execution_blocked_by": executionBlockedByNexus,
		"evidence": map[string]any{
			"source_ref":        "ponti.ai.actions." + action,
			"captured_at":       time.Now().UTC().Format(time.RFC3339),
			"tenant_scope":      orgID.String(),
			"actor_id":          actor,
			"workspace":         workspace,
			"approval_required": true,
		},
	}
}

func draftActionExecutionResponse(action, actionType string, orgID uuid.UUID, actor string, nexusRequestID string, draftID any, writePerformed bool, executionStatus string, workspace map[string]any, proposal map[string]any) map[string]any {
	nexusRequestID = strings.TrimSpace(nexusRequestID)
	auditRef := fmt.Sprintf("ponti.ai.actions.%s:%v", action, draftID)
	return map[string]any{
		"status":               "draft",
		"action":               action,
		"approval_required":    true,
		"nexus_action_type":    actionType,
		"side_effect_type":     "write",
		"preview_only":         !writePerformed,
		"write_performed":      writePerformed,
		"draft_id":             draftID,
		"execution_status":     executionStatus,
		"nexus_request_id":     nexusRequestID,
		"audit_ref":            auditRef,
		"proposal":             proposal,
		"execution_allowed":    true,
		"execution_blocked_by": "",
		"evidence": map[string]any{
			"source_ref":        "ponti.ai.actions." + action,
			"captured_at":       time.Now().UTC().Format(time.RFC3339),
			"tenant_scope":      orgID.String(),
			"actor_id":          actor,
			"workspace":         workspace,
			"approval_required": true,
			"nexus_request_id":  nexusRequestID,
			"audit_ref":         auditRef,
		},
	}
}

func (w workspaceRequest) withFallbackProject(projectID int64) workspaceRequest {
	if w.ProjectID == nil && projectID > 0 {
		w.ProjectID = &projectID
	}
	return w
}

func (w workspaceRequest) withFallbackCampaign(campaignID *int64) workspaceRequest {
	if w.CampaignID == nil && campaignID != nil && *campaignID > 0 {
		w.CampaignID = campaignID
	}
	return w
}

func (w workspaceRequest) withFallbackField(fieldID *int64) workspaceRequest {
	if w.FieldID == nil && fieldID != nil && *fieldID > 0 {
		w.FieldID = fieldID
	}
	return w
}

func (w workspaceRequest) toMap() map[string]any {
	out := map[string]any{}
	if w.CustomerID != nil && *w.CustomerID > 0 {
		out["customer_id"] = *w.CustomerID
	}
	if w.ProjectID != nil && *w.ProjectID > 0 {
		out["project_id"] = *w.ProjectID
	}
	if w.CampaignID != nil && *w.CampaignID > 0 {
		out["campaign_id"] = *w.CampaignID
	}
	if w.FieldID != nil && *w.FieldID > 0 {
		out["field_id"] = *w.FieldID
	}
	return out
}

func valueOrNil(value *int64) any {
	if value == nil {
		return nil
	}
	return *value
}
