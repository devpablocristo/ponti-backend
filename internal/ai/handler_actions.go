package ai

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
)

const draftActionStatusPreview = "preview"

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
	workspace := req.Workspace.toMap()
	proposal := map[string]any{
		"insight_id":      req.InsightID,
		"resolution_note": req.ResolutionNote,
		"preview_only":    true,
		"write_performed": false,
	}
	c.JSON(http.StatusOK, draftActionPreviewResponse("ponti.insight.resolve.prepare", orgID, actor, workspace, proposal))
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
	c.JSON(http.StatusOK, draftActionPreviewResponse("ponti.workorder.draft.prepare", orgID, actor, workspace, proposal))
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
	workspace := req.Workspace.withFallbackProject(req.ProjectID).toMap()
	proposal := map[string]any{
		"project_id":      req.ProjectID,
		"supply_id":       req.SupplyID,
		"quantity_delta":  req.QuantityDelta,
		"reason":          req.Reason,
		"preview_only":    true,
		"write_performed": false,
	}
	c.JSON(http.StatusOK, draftActionPreviewResponse("ponti.stock_adjustment.prepare", orgID, actor, workspace, proposal))
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
	draftID := "insight-resolution-" + uuid.NewString()
	workspace := req.Workspace.toMap()
	c.JSON(http.StatusOK, draftActionExecutionResponse(
		"ponti.insight_resolution.draft",
		orgID,
		actor,
		c.GetHeader("X-Nexus-Request-ID"),
		draftID,
		false,
		"draft_staged",
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
	draftID := "stock-count-" + uuid.NewString()
	workspace := req.Workspace.withFallbackProject(req.ProjectID).toMap()
	c.JSON(http.StatusOK, draftActionExecutionResponse(
		"ponti.stock_count.draft",
		orgID,
		actor,
		c.GetHeader("X-Nexus-Request-ID"),
		draftID,
		false,
		"draft_staged",
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

func draftActionPreviewResponse(action string, orgID uuid.UUID, actor string, workspace map[string]any, proposal map[string]any) map[string]any {
	return map[string]any{
		"status":               draftActionStatusPreview,
		"action":               action,
		"approval_required":    true,
		"nexus_action_type":    pontiNexusActionType,
		"side_effect_type":     "write",
		"preview_only":         true,
		"write_performed":      false,
		"proposal":             proposal,
		"pending_execution":    true,
		"execution_allowed":    false,
		"execution_blocked_by": "nexus_required",
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

func draftActionExecutionResponse(action string, orgID uuid.UUID, actor string, nexusRequestID string, draftID any, writePerformed bool, executionStatus string, workspace map[string]any, proposal map[string]any) map[string]any {
	nexusRequestID = strings.TrimSpace(nexusRequestID)
	auditRef := fmt.Sprintf("ponti.ai.actions.%s:%v", action, draftID)
	return map[string]any{
		"status":               "draft",
		"action":               action,
		"approval_required":    true,
		"nexus_action_type":    pontiNexusActionType,
		"side_effect_type":     "write",
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
