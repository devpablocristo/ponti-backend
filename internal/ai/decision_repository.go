package ai

import (
	"context"
	"errors"
	"time"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	shareddb "github.com/devpablocristo/ponti-backend/internal/shared/db"
)

type decisionRepository struct {
	db *gorm.DB
}

func NewDecisionRepository(db *gorm.DB) *decisionRepository {
	return &decisionRepository{db: db}
}

func (r *decisionRepository) createRun(ctx context.Context, run decisionRunModel) (decisionRunModel, error) {
	if r == nil || r.db == nil {
		return decisionRunModel{}, domainerr.Internal("ai decision repository unavailable")
	}
	if run.ID == uuid.Nil {
		run.ID = uuid.New()
	}
	now := time.Now().UTC()
	if run.StartedAt.IsZero() {
		run.StartedAt = now
	}
	if run.CreatedAt.IsZero() {
		run.CreatedAt = now
	}
	if run.UpdatedAt.IsZero() {
		run.UpdatedAt = now
	}
	if run.Status == "" {
		run.Status = decisionRunStatusRunning
	}
	if run.RoutingSource == "" {
		run.RoutingSource = "deterministic"
	}
	if len(run.WorkspaceJSON) == 0 {
		run.WorkspaceJSON = []byte("{}")
	}
	if err := r.db.WithContext(ctx).Create(&run).Error; err != nil {
		return decisionRunModel{}, err
	}
	return run, nil
}

func (r *decisionRepository) completeRun(ctx context.Context, run decisionRunModel) (decisionRunModel, error) {
	now := time.Now().UTC()
	run.CompletedAt = &now
	run.UpdatedAt = now
	if err := r.db.WithContext(ctx).Model(&decisionRunModel{}).
		Where("id = ? AND tenant_id = ?", run.ID, run.TenantID).
		Updates(map[string]any{
			"status":          run.Status,
			"routing_source":  run.RoutingSource,
			"axis_run_id":     run.AxisRunID,
			"axis_task_id":    run.AxisTaskID,
			"degraded_reason": run.DegradedReason,
			"cards_created":   run.CardsCreated,
			"cards_updated":   run.CardsUpdated,
			"cards_total":     run.CardsTotal,
			"completed_at":    run.CompletedAt,
			"updated_at":      run.UpdatedAt,
		}).Error; err != nil {
		return decisionRunModel{}, err
	}
	return run, nil
}

func (r *decisionRepository) upsertCards(ctx context.Context, tenantID uuid.UUID, runID uuid.UUID, workspace map[string]any, actor string, drafts []decisionCardDraft) ([]decisionCardModel, int, int, error) {
	now := time.Now().UTC()
	out := make([]decisionCardModel, 0, len(drafts))
	created := 0
	updated := 0
	workspaceJSON := marshalDecisionJSON(workspace, "{}")

	for _, draft := range drafts {
		if draft.Fingerprint == "" {
			continue
		}
		var existing decisionCardModel
		err := r.db.WithContext(ctx).
			Where("tenant_id = ? AND fingerprint = ?", tenantID, draft.Fingerprint).
			First(&existing).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, created, updated, err
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			card := decisionCardModel{
				ID:              uuid.New(),
				TenantID:        tenantID,
				DecisionRunID:   &runID,
				WorkspaceJSON:   workspaceJSON,
				Fingerprint:     draft.Fingerprint,
				Domain:          draft.Domain,
				RouteHint:       draft.RouteHint,
				Severity:        draft.Severity,
				Bucket:          draft.Bucket,
				Status:          decisionStatusOpen,
				Title:           draft.Title,
				Summary:         draft.Summary,
				Recommendation:  draft.Recommendation,
				ImpactLabel:     draft.ImpactLabel,
				ImpactValue:     draft.ImpactValue,
				Source:          draft.Source,
				EvidenceJSON:    marshalDecisionJSON(draft.Evidence, "{}"),
				ToolsJSON:       marshalDecisionJSON(draft.Tools, "[]"),
				ActionJSON:      marshalDecisionJSON(draft.Action, "{}"),
				AxisRunID:       draft.AxisRunID,
				AxisTaskID:      draft.AxisTaskID,
				OccurrenceCount: 1,
				FirstSeenAt:     now,
				LastSeenAt:      now,
				LastActor:       actor,
				CreatedAt:       now,
				UpdatedAt:       now,
			}
			if err := r.db.WithContext(ctx).Create(&card).Error; err != nil {
				return nil, created, updated, err
			}
			created++
			out = append(out, card)
			continue
		}

		status := existing.Status
		if status == "" || status == decisionStatusResolved {
			status = decisionStatusOpen
		}
		runIDCopy := runID
		updates := map[string]any{
			"decision_run_id":  &runIDCopy,
			"workspace_json":   workspaceJSON,
			"domain":           draft.Domain,
			"route_hint":       draft.RouteHint,
			"severity":         draft.Severity,
			"bucket":           draft.Bucket,
			"status":           status,
			"title":            draft.Title,
			"summary":          draft.Summary,
			"recommendation":   draft.Recommendation,
			"impact_label":     draft.ImpactLabel,
			"impact_value":     draft.ImpactValue,
			"source":           draft.Source,
			"evidence_json":    marshalDecisionJSON(draft.Evidence, "{}"),
			"tools_json":       marshalDecisionJSON(draft.Tools, "[]"),
			"action_json":      marshalDecisionJSON(draft.Action, "{}"),
			"axis_run_id":      draft.AxisRunID,
			"axis_task_id":     draft.AxisTaskID,
			"occurrence_count": gorm.Expr("occurrence_count + 1"),
			"last_seen_at":     now,
			"last_actor":       actor,
			"updated_at":       now,
		}
		if err := r.db.WithContext(ctx).Model(&decisionCardModel{}).
			Where("id = ? AND tenant_id = ?", existing.ID, tenantID).
			Updates(updates).Error; err != nil {
			return nil, created, updated, err
		}
		if err := r.db.WithContext(ctx).First(&existing, "id = ? AND tenant_id = ?", existing.ID, tenantID).Error; err != nil {
			return nil, created, updated, err
		}
		updated++
		out = append(out, existing)
	}
	return out, created, updated, nil
}

func (r *decisionRepository) listRuns(ctx context.Context, tenantID uuid.UUID, limit int) ([]decisionRunModel, error) {
	if limit <= 0 || limit > 100 {
		limit = 25
	}
	var rows []decisionRunModel
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Limit(limit).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *decisionRepository) listCards(ctx context.Context, tenantID uuid.UUID, filters decisionCardFilters) ([]decisionCardModel, error) {
	limit := filters.Limit
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if filters.Status != "" {
		q = q.Where("status = ?", filters.Status)
	} else if !filters.IncludeResolved {
		q = q.Where("status NOT IN ?", []string{decisionStatusResolved, decisionStatusDismissed})
	}
	if filters.RouteHint != "" {
		q = q.Where("route_hint = ?", filters.RouteHint)
	}
	if filters.Domain != "" {
		q = q.Where("domain = ?", filters.Domain)
	}
	if filters.Bucket != "" {
		q = q.Where("bucket = ?", filters.Bucket)
	}
	var rows []decisionCardModel
	if err := q.Order("last_seen_at DESC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *decisionRepository) getCard(ctx context.Context, tenantID uuid.UUID, cardID string) (decisionCardModel, error) {
	id, err := uuid.Parse(cardID)
	if err != nil {
		return decisionCardModel{}, domainerr.Validation("invalid decision card id")
	}
	var row decisionCardModel
	err = r.db.WithContext(ctx).First(&row, "id = ? AND tenant_id = ?", id, tenantID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return decisionCardModel{}, domainerr.NotFound("decision card not found")
	}
	if err != nil {
		return decisionCardModel{}, err
	}
	return row, nil
}

func (r *decisionRepository) patchCardStatus(ctx context.Context, tenantID uuid.UUID, cardID string, actor string, status string, snoozeUntil *time.Time) (decisionCardModel, error) {
	if _, ok := validDecisionStatuses[status]; !ok {
		return decisionCardModel{}, domainerr.Validation("invalid decision status")
	}
	card, err := r.getCard(ctx, tenantID, cardID)
	if err != nil {
		return decisionCardModel{}, err
	}
	now := time.Now().UTC()
	updates := map[string]any{
		"status":            status,
		"status_changed_at": now,
		"last_actor":        actor,
		"updated_at":        now,
	}
	if status == decisionStatusSnoozed {
		updates["snooze_until"] = snoozeUntil
	} else {
		updates["snooze_until"] = nil
	}
	if err := r.db.WithContext(ctx).Model(&decisionCardModel{}).
		Where("id = ? AND tenant_id = ?", card.ID, tenantID).
		Updates(updates).Error; err != nil {
		return decisionCardModel{}, err
	}
	return r.getCard(ctx, tenantID, cardID)
}

type decisionSupplySignal struct {
	ID             int64
	Name           string
	Price          decimal.Decimal
	IsPartialPrice bool
	IsPending      bool
}

func (r *decisionRepository) listTentativeSupplies(ctx context.Context, workspace workspaceRequest, limit int) ([]decisionSupplySignal, error) {
	if workspace.ProjectID == nil || *workspace.ProjectID <= 0 {
		return []decisionSupplySignal{}, nil
	}
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	var rows []decisionSupplySignal
	err := r.db.WithContext(ctx).
		Table("supplies").
		Select("id, name, price, is_partial_price, is_pending").
		Where("project_id = ? AND deleted_at IS NULL AND (is_partial_price = true OR is_pending = true)", *workspace.ProjectID).
		Order("is_pending DESC, is_partial_price DESC, lower(name) ASC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

type decisionDraftWorkOrderSignal struct {
	Count int64
}

func (r *decisionRepository) countDraftWorkOrders(ctx context.Context, workspace workspaceRequest) (decisionDraftWorkOrderSignal, error) {
	if workspace.ProjectID == nil || *workspace.ProjectID <= 0 {
		return decisionDraftWorkOrderSignal{}, nil
	}
	q := r.db.WithContext(ctx).
		Table("work_order_drafts").
		Where("project_id = ? AND status = ? AND deleted_at IS NULL", *workspace.ProjectID, "draft")
	if workspace.FieldID != nil && *workspace.FieldID > 0 {
		q = q.Where("field_id = ?", *workspace.FieldID)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return decisionDraftWorkOrderSignal{}, err
	}
	return decisionDraftWorkOrderSignal{Count: count}, nil
}

type decisionLotRiskSignal struct {
	ID                   int64
	LotName              string
	FieldID              int64
	FieldName            string
	CurrentCropID        int64
	CurrentCrop          string
	SowedArea            decimal.Decimal
	HarvestedArea        decimal.Decimal
	CostUSDPerHa         decimal.Decimal
	YieldTnPerHa         decimal.Decimal
	OperatingResultPerHa decimal.Decimal
}

func (r *decisionRepository) listLotRiskSignals(ctx context.Context, workspace workspaceRequest, limit int) ([]decisionLotRiskSignal, error) {
	if workspace.ProjectID == nil || *workspace.ProjectID <= 0 {
		return []decisionLotRiskSignal{}, nil
	}
	if limit <= 0 || limit > 20 {
		limit = 5
	}
	where := "project_id = ?"
	args := []any{*workspace.ProjectID}
	if workspace.FieldID != nil && *workspace.FieldID > 0 {
		where += " AND field_id = ?"
		args = append(args, *workspace.FieldID)
	}
	args = append(args, limit)

	var rows []decisionLotRiskSignal
	query := `
		SELECT
			id,
			lot_name,
			field_id,
			field_name,
			current_crop_id,
			current_crop,
			sowed_area_ha AS sowed_area,
			harvested_area_ha AS harvested_area,
			cost_usd_per_ha,
			yield_tn_per_ha,
			operating_result_per_ha_usd AS operating_result_per_ha
		FROM ` + shareddb.ReportView("lot_list") + `
		WHERE ` + where + `
		  AND (
			operating_result_per_ha_usd < 0
			OR (COALESCE(sowed_area_ha, 0) > 0 AND COALESCE(harvested_area_ha, 0) = 0)
		  )
		ORDER BY operating_result_per_ha_usd ASC, id DESC
		LIMIT ?
	`
	if err := r.db.WithContext(ctx).Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}
