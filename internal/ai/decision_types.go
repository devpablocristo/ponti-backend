package ai

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	decisionStatusOpen      = "open"
	decisionStatusAccepted  = "accepted"
	decisionStatusDrafted   = "drafted"
	decisionStatusDismissed = "dismissed"
	decisionStatusSnoozed   = "snoozed"
	decisionStatusResolved  = "resolved"

	decisionRunStatusCompleted = "completed"
	decisionRunStatusDegraded  = "degraded"
	decisionRunStatusFailed    = "failed"
	decisionRunStatusRunning   = "running"

	decisionBucketUrgent      = "urgent"
	decisionBucketImportant   = "important"
	decisionBucketOpportunity = "opportunity"
	decisionBucketFollowUp    = "follow_up"
)

var validDecisionStatuses = map[string]struct{}{
	decisionStatusOpen:      {},
	decisionStatusAccepted:  {},
	decisionStatusDrafted:   {},
	decisionStatusDismissed: {},
	decisionStatusSnoozed:   {},
	decisionStatusResolved:  {},
}

type decisionRunModel struct {
	ID             uuid.UUID  `gorm:"column:id;type:uuid;primaryKey"`
	TenantID       uuid.UUID  `gorm:"column:tenant_id;type:uuid;not null"`
	WorkspaceJSON  []byte     `gorm:"column:workspace_json;type:jsonb;not null"`
	RequestedBy    string     `gorm:"column:requested_by;not null"`
	Status         string     `gorm:"column:status;not null"`
	RoutingSource  string     `gorm:"column:routing_source;not null"`
	AxisRunID      string     `gorm:"column:axis_run_id;not null"`
	AxisTaskID     string     `gorm:"column:axis_task_id;not null"`
	DegradedReason string     `gorm:"column:degraded_reason;not null"`
	CardsCreated   int        `gorm:"column:cards_created;not null"`
	CardsUpdated   int        `gorm:"column:cards_updated;not null"`
	CardsTotal     int        `gorm:"column:cards_total;not null"`
	StartedAt      time.Time  `gorm:"column:started_at;not null"`
	CompletedAt    *time.Time `gorm:"column:completed_at"`
	CreatedAt      time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;not null"`
}

func (decisionRunModel) TableName() string {
	return "ai_decision_runs"
}

type decisionCardModel struct {
	ID              uuid.UUID  `gorm:"column:id;type:uuid;primaryKey"`
	TenantID        uuid.UUID  `gorm:"column:tenant_id;type:uuid;not null"`
	DecisionRunID   *uuid.UUID `gorm:"column:decision_run_id;type:uuid"`
	WorkspaceJSON   []byte     `gorm:"column:workspace_json;type:jsonb;not null"`
	Fingerprint     string     `gorm:"column:fingerprint;not null"`
	Domain          string     `gorm:"column:domain;not null"`
	RouteHint       string     `gorm:"column:route_hint;not null"`
	Severity        string     `gorm:"column:severity;not null"`
	Bucket          string     `gorm:"column:bucket;not null"`
	Status          string     `gorm:"column:status;not null"`
	Title           string     `gorm:"column:title;not null"`
	Summary         string     `gorm:"column:summary;not null"`
	Recommendation  string     `gorm:"column:recommendation;not null"`
	ImpactLabel     string     `gorm:"column:impact_label;not null"`
	ImpactValue     *float64   `gorm:"column:impact_value"`
	Source          string     `gorm:"column:source;not null"`
	EvidenceJSON    []byte     `gorm:"column:evidence_json;type:jsonb;not null"`
	ToolsJSON       []byte     `gorm:"column:tools_json;type:jsonb;not null"`
	ActionJSON      []byte     `gorm:"column:action_json;type:jsonb;not null"`
	AxisRunID       string     `gorm:"column:axis_run_id;not null"`
	AxisTaskID      string     `gorm:"column:axis_task_id;not null"`
	OccurrenceCount int        `gorm:"column:occurrence_count;not null"`
	FirstSeenAt     time.Time  `gorm:"column:first_seen_at;not null"`
	LastSeenAt      time.Time  `gorm:"column:last_seen_at;not null"`
	SnoozeUntil     *time.Time `gorm:"column:snooze_until"`
	StatusChangedAt *time.Time `gorm:"column:status_changed_at"`
	LastActor       string     `gorm:"column:last_actor;not null"`
	CreatedAt       time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;not null"`
}

func (decisionCardModel) TableName() string {
	return "ai_decision_cards"
}

type decisionCardDraft struct {
	Fingerprint    string
	Domain         string
	RouteHint      string
	Severity       string
	Bucket         string
	Title          string
	Summary        string
	Recommendation string
	ImpactLabel    string
	ImpactValue    *float64
	Source         string
	Evidence       map[string]any
	Tools          []any
	Action         map[string]any
	AxisRunID      string
	AxisTaskID     string
}

type decisionRunInput struct {
	Workspace workspaceRequest `json:"workspace"`
	RouteHint string           `json:"route_hint,omitempty"`
}

type externalDecisionInput struct {
	Workspace      workspaceRequest `json:"workspace"`
	Fingerprint    string           `json:"fingerprint"`
	Domain         string           `json:"domain"`
	RouteHint      string           `json:"route_hint"`
	Severity       string           `json:"severity"`
	Bucket         string           `json:"bucket"`
	Title          string           `json:"title"`
	Summary        string           `json:"summary"`
	Recommendation string           `json:"recommendation"`
	ImpactLabel    string           `json:"impact_label,omitempty"`
	ImpactValue    *float64         `json:"impact_value,omitempty"`
	Source         string           `json:"source,omitempty"`
	Evidence       map[string]any   `json:"evidence,omitempty"`
	Tools          []any            `json:"tools,omitempty"`
	Action         map[string]any   `json:"action,omitempty"`
	AxisRunID      string           `json:"axis_run_id,omitempty"`
	AxisTaskID     string           `json:"axis_task_id,omitempty"`
}

type decisionCardFilters struct {
	RouteHint       string
	Domain          string
	Bucket          string
	Status          string
	IncludeResolved bool
	Limit           int
}

type decisionRunResult struct {
	Run   decisionRunModel
	Cards []decisionCardModel
}

func marshalDecisionJSON(v any, fallback string) []byte {
	raw, err := json.Marshal(v)
	if err != nil || len(raw) == 0 {
		return []byte(fallback)
	}
	return raw
}

func unmarshalDecisionMap(raw []byte) map[string]any {
	out := map[string]any{}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &out)
	}
	return out
}

func unmarshalDecisionArray(raw []byte) []any {
	out := []any{}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &out)
	}
	return out
}
