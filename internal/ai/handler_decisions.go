package ai

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	"github.com/gin-gonic/gin"

	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
)

type decisionStatusPatchRequest struct {
	Status      string `json:"status"`
	SnoozeUntil string `json:"snooze_until,omitempty"`
}

func (h *Handler) SetDecisionService(svc *DecisionService) {
	h.decisions = svc
}

func (h *Handler) CreateDecisionRun(c *gin.Context) {
	svc, ok := h.requireDecisionService(c)
	if !ok {
		return
	}
	orgID, err := sharedhandlers.ParseOrgID(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	actor, err := sharedhandlers.ParseActor(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	var req decisionRunInput
	if err := c.ShouldBindJSON(&req); err != nil {
		sharedhandlers.RespondError(c, domainerr.Validation("invalid request payload"))
		return
	}
	result, err := svc.Run(c.Request.Context(), orgID, actor, req)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"run":   decisionRunResponse(result.Run),
		"cards": decisionCardResponses(result.Cards),
	})
}

func (h *Handler) ListDecisionRuns(c *gin.Context) {
	svc, ok := h.requireDecisionService(c)
	if !ok {
		return
	}
	orgID, err := sharedhandlers.ParseOrgID(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	runs, err := svc.ListRuns(c.Request.Context(), orgID, parseIntQuery(c, "limit", 25))
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	items := make([]map[string]any, 0, len(runs))
	for _, run := range runs {
		items = append(items, decisionRunResponse(run))
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) ListDecisionCards(c *gin.Context) {
	svc, ok := h.requireDecisionService(c)
	if !ok {
		return
	}
	orgID, err := sharedhandlers.ParseOrgID(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	filters := decisionCardFilters{
		RouteHint: strings.TrimSpace(c.Query("route_hint")),
		Domain:    strings.TrimSpace(c.Query("domain")),
		Bucket:    strings.TrimSpace(c.Query("bucket")),
		Status:    strings.TrimSpace(c.Query("status")),
		Limit:     parseIntQuery(c, "limit", 100),
	}
	filters.IncludeResolved = strings.EqualFold(c.Query("include_resolved"), "true")
	cards, err := svc.ListCards(c.Request.Context(), orgID, filters)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": decisionCardResponses(cards)})
}

func (h *Handler) PatchDecisionCard(c *gin.Context) {
	svc, ok := h.requireDecisionService(c)
	if !ok {
		return
	}
	orgID, err := sharedhandlers.ParseOrgID(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	actor, err := sharedhandlers.ParseActor(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	var req decisionStatusPatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sharedhandlers.RespondError(c, domainerr.Validation("invalid request payload"))
		return
	}
	var snoozeUntil *time.Time
	if strings.TrimSpace(req.SnoozeUntil) != "" {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(req.SnoozeUntil))
		if err != nil {
			sharedhandlers.RespondError(c, domainerr.Validation("snooze_until must be RFC3339"))
			return
		}
		snoozeUntil = &parsed
	}
	card, err := svc.PatchCardStatus(c.Request.Context(), orgID, c.Param("id"), actor, req.Status, snoozeUntil)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, decisionCardResponse(card))
}

func (h *Handler) ImportExternalDecisionCard(c *gin.Context) {
	svc, ok := h.requireDecisionService(c)
	if !ok {
		return
	}
	orgID, err := sharedhandlers.ParseOrgID(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	actor, err := sharedhandlers.ParseActor(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	var req externalDecisionInput
	if err := c.ShouldBindJSON(&req); err != nil {
		sharedhandlers.RespondError(c, domainerr.Validation("invalid request payload"))
		return
	}
	card, err := svc.ImportExternalCard(c.Request.Context(), orgID, actor, req)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, decisionCardResponse(card))
}

func (h *Handler) ExecuteDecisionCardAction(c *gin.Context) {
	svc, ok := h.requireDecisionService(c)
	if !ok {
		return
	}
	orgID, err := sharedhandlers.ParseOrgID(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	actor, err := sharedhandlers.ParseActor(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	card, action, err := svc.PrepareCardAction(c.Request.Context(), orgID, c.Param("id"), c.Param("action_id"), actor)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	approvalRequired, _ := action["requires_approval"].(bool)
	status := "ready"
	if approvalRequired {
		status = "pending_approval"
	}
	auditRef := "ponti.ai.decision_cards." + card.ID.String()
	c.JSON(http.StatusAccepted, gin.H{
		"status":               status,
		"execution_status":     status,
		"action_id":            action["id"],
		"capability_id":        action["capability_id"],
		"approval_required":    approvalRequired,
		"nexus_action_type":    action["nexus_action_type"],
		"side_effect_type":     "write",
		"write_performed":      false,
		"draft_id":             nil,
		"nexus_request_id":     nil,
		"audit_ref":            auditRef,
		"pending_execution":    approvalRequired,
		"execution_allowed":    !approvalRequired,
		"execution_blocked_by": blockedBy(approvalRequired),
		"proposal":             action["payload"],
		"missing_inputs":       action["missing_inputs"],
		"card":                 decisionCardResponse(card),
		"pending_confirmation": gin.H{
			"card_id":           card.ID.String(),
			"action_id":         action["id"],
			"capability_id":     action["capability_id"],
			"approval_required": approvalRequired,
		},
		"evidence": gin.H{
			"source_ref":        auditRef,
			"captured_at":       time.Now().UTC().Format(time.RFC3339),
			"tenant_scope":      orgID.String(),
			"actor_id":          actor,
			"decision_card_id":  card.ID.String(),
			"approval_required": approvalRequired,
			"workspace":         unmarshalDecisionMap(card.WorkspaceJSON),
		},
	})
}

func (h *Handler) requireDecisionService(c *gin.Context) (*DecisionService, bool) {
	if h.decisions == nil {
		sharedhandlers.RespondError(c, domainerr.Internal("ai decision service unavailable"))
		return nil, false
	}
	return h.decisions, true
}

func decisionRunResponse(run decisionRunModel) map[string]any {
	out := map[string]any{
		"id":              run.ID.String(),
		"tenant_id":       run.TenantID.String(),
		"workspace":       unmarshalDecisionMap(run.WorkspaceJSON),
		"requested_by":    run.RequestedBy,
		"status":          run.Status,
		"routing_source":  run.RoutingSource,
		"axis_run_id":     run.AxisRunID,
		"axis_task_id":    run.AxisTaskID,
		"degraded_reason": run.DegradedReason,
		"cards_created":   run.CardsCreated,
		"cards_updated":   run.CardsUpdated,
		"cards_total":     run.CardsTotal,
		"started_at":      run.StartedAt.Format(time.RFC3339),
		"created_at":      run.CreatedAt.Format(time.RFC3339),
		"updated_at":      run.UpdatedAt.Format(time.RFC3339),
	}
	if run.CompletedAt != nil {
		out["completed_at"] = run.CompletedAt.Format(time.RFC3339)
	}
	return out
}

func decisionCardResponses(cards []decisionCardModel) []map[string]any {
	out := make([]map[string]any, 0, len(cards))
	for _, card := range cards {
		out = append(out, decisionCardResponse(card))
	}
	return out
}

func decisionCardResponse(card decisionCardModel) map[string]any {
	out := map[string]any{
		"id":               card.ID.String(),
		"tenant_id":        card.TenantID.String(),
		"workspace":        unmarshalDecisionMap(card.WorkspaceJSON),
		"fingerprint":      card.Fingerprint,
		"domain":           card.Domain,
		"route_hint":       card.RouteHint,
		"severity":         card.Severity,
		"bucket":           card.Bucket,
		"status":           card.Status,
		"title":            card.Title,
		"summary":          card.Summary,
		"recommendation":   card.Recommendation,
		"impact_label":     card.ImpactLabel,
		"impact_value":     card.ImpactValue,
		"source":           card.Source,
		"evidence":         unmarshalDecisionMap(card.EvidenceJSON),
		"tools":            unmarshalDecisionArray(card.ToolsJSON),
		"action":           unmarshalDecisionMap(card.ActionJSON),
		"axis_run_id":      card.AxisRunID,
		"axis_task_id":     card.AxisTaskID,
		"occurrence_count": card.OccurrenceCount,
		"first_seen_at":    card.FirstSeenAt.Format(time.RFC3339),
		"last_seen_at":     card.LastSeenAt.Format(time.RFC3339),
		"last_actor":       card.LastActor,
		"created_at":       card.CreatedAt.Format(time.RFC3339),
		"updated_at":       card.UpdatedAt.Format(time.RFC3339),
	}
	if card.DecisionRunID != nil {
		out["decision_run_id"] = card.DecisionRunID.String()
	}
	if card.SnoozeUntil != nil {
		out["snooze_until"] = card.SnoozeUntil.Format(time.RFC3339)
	}
	if card.StatusChangedAt != nil {
		out["status_changed_at"] = card.StatusChangedAt.Format(time.RFC3339)
	}
	return out
}

func parseIntQuery(c *gin.Context, key string, fallback int) int {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

func blockedBy(approvalRequired bool) string {
	if approvalRequired {
		return "nexus_required"
	}
	return ""
}
