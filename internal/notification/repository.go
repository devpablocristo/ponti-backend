package notification

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type inboxNotification struct {
	ID              int64      `gorm:"column:id"`
	OrgID           uuid.UUID  `gorm:"column:org_id"`
	ProjectID       *int64     `gorm:"column:project_id"`
	RecipientActor  *string    `gorm:"column:recipient_actor"`
	Kind            string     `gorm:"column:kind"`
	Source          string     `gorm:"column:source"`
	SourceRef       *string    `gorm:"column:source_ref"`
	NotificationKey string     `gorm:"column:notification_key"`
	Title           string     `gorm:"column:title"`
	Body            string     `gorm:"column:body"`
	Severity        int        `gorm:"column:severity"`
	Status          string     `gorm:"column:status"`
	RouteHint       string     `gorm:"column:route_hint"`
	Payload         []byte     `gorm:"column:payload"`
	CreatedBy       string     `gorm:"column:created_by"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at"`
	ReadAt          *time.Time `gorm:"column:read_at"`
	DismissedAt     *time.Time `gorm:"column:dismissed_at"`
}

func (inboxNotification) TableName() string { return "ponti_notifications" }

type notificationCandidate struct {
	ID           int64      `gorm:"column:id"`
	OrgID        uuid.UUID  `gorm:"column:org_id"`
	ProjectID    *int64     `gorm:"column:project_id"`
	CandidateKey string     `gorm:"column:candidate_key"`
	Kind         string     `gorm:"column:kind"`
	Source       string     `gorm:"column:source"`
	SourceRef    *string    `gorm:"column:source_ref"`
	Title        string     `gorm:"column:title"`
	Body         string     `gorm:"column:body"`
	Severity     int        `gorm:"column:severity"`
	Status       string     `gorm:"column:status"`
	Payload      []byte     `gorm:"column:payload"`
	CreatedBy    string     `gorm:"column:created_by"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"`
	ResolvedAt   *time.Time `gorm:"column:resolved_at"`
}

func (notificationCandidate) TableName() string { return "ponti_notification_candidates" }

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

type ListFilters struct {
	OrgID     uuid.UUID
	Actor     string
	ProjectID *int64
	Status    string
	Kind      string
	Limit     int
	Offset    int
}

type Summary struct {
	Total          int64 `json:"total"`
	NewCount       int64 `json:"new_count"`
	ReadCount      int64 `json:"read_count"`
	DismissedCount int64 `json:"dismissed_count"`
	HighSeverity   int64 `json:"high_severity_count"`
}

type NotificationRecord struct {
	ID              int64          `json:"id"`
	OrgID           uuid.UUID      `json:"org_id"`
	ProjectID       *int64         `json:"project_id,omitempty"`
	RecipientActor  *string        `json:"recipient_actor,omitempty"`
	Kind            string         `json:"kind"`
	Source          string         `json:"source"`
	SourceRef       *string        `json:"source_ref,omitempty"`
	NotificationKey string         `json:"notification_key"`
	Title           string         `json:"title"`
	Body            string         `json:"body"`
	Severity        int            `json:"severity"`
	Status          string         `json:"status"`
	RouteHint       string         `json:"route_hint"`
	Payload         map[string]any `json:"payload"`
	CreatedBy       string         `json:"created_by"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	ReadAt          *time.Time     `json:"read_at,omitempty"`
	DismissedAt     *time.Time     `json:"dismissed_at,omitempty"`
}

func (r *Repository) List(ctx context.Context, f ListFilters) ([]NotificationRecord, error) {
	limit := f.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}
	q := r.baseInboxQuery(ctx, f.OrgID, f.Actor, f.ProjectID)
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.Kind != "" {
		q = q.Where("kind = ?", f.Kind)
	}
	var rows []inboxNotification
	err := q.Order("CASE WHEN status = 'new' THEN 0 ELSE 1 END").
		Order("created_at DESC").
		Limit(limit).
		Offset(maxInt(f.Offset, 0)).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return mapNotificationRows(rows), nil
}

func (r *Repository) Get(ctx context.Context, orgID uuid.UUID, actor string, id int64) (*NotificationRecord, error) {
	var row inboxNotification
	err := r.baseInboxQuery(ctx, orgID, actor, nil).Where("id = ?", id).Take(&row).Error
	if err != nil {
		return nil, err
	}
	record := mapNotificationRow(row)
	return &record, nil
}

func (r *Repository) GetSummary(ctx context.Context, orgID uuid.UUID, actor string, projectID *int64) (Summary, error) {
	type row struct {
		Total          int64
		NewCount       int64
		ReadCount      int64
		DismissedCount int64
		HighSeverity   int64
	}
	var out row
	err := r.baseInboxQuery(ctx, orgID, actor, projectID).
		Select(`
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE status = 'new') AS new_count,
			COUNT(*) FILTER (WHERE status = 'read') AS read_count,
			COUNT(*) FILTER (WHERE status = 'dismissed') AS dismissed_count,
			COUNT(*) FILTER (WHERE severity >= 80 AND status = 'new') AS high_severity
		`).
		Scan(&out).Error
	return Summary{
		Total:          out.Total,
		NewCount:       out.NewCount,
		ReadCount:      out.ReadCount,
		DismissedCount: out.DismissedCount,
		HighSeverity:   out.HighSeverity,
	}, err
}

func (r *Repository) MarkStatus(ctx context.Context, orgID uuid.UUID, actor string, id int64, status string) error {
	updates := map[string]any{
		"status":     status,
		"updated_at": time.Now().UTC(),
	}
	switch status {
	case "read":
		updates["read_at"] = time.Now().UTC()
	case "dismissed":
		updates["dismissed_at"] = time.Now().UTC()
	}
	return r.baseInboxQuery(ctx, orgID, actor, nil).
		Where("id = ?", id).
		Updates(updates).Error
}

type ProjectedNotificationInput struct {
	OrgID           uuid.UUID
	ProjectID       int64
	Actor           string
	Kind            string
	Source          string
	SourceRef       string
	NotificationKey string
	Title           string
	Body            string
	Severity        int
	RouteHint       string
	CreatedBy       string
	Payload         map[string]any
}

func (r *Repository) UpsertProjectedNotifications(ctx context.Context, items []ProjectedNotificationInput) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			payload, err := json.Marshal(item.Payload)
			if err != nil {
				return err
			}
			sourceRef := item.SourceRef
			projectID := item.ProjectID
			candidate := notificationCandidate{
				OrgID:        item.OrgID,
				ProjectID:    &projectID,
				CandidateKey: item.NotificationKey,
				Kind:         item.Kind,
				Source:       item.Source,
				SourceRef:    &sourceRef,
				Title:        item.Title,
				Body:         item.Body,
				Severity:     item.Severity,
				Status:       "new",
				Payload:      payload,
				CreatedBy:    item.CreatedBy,
			}
			if err := tx.
				Where("org_id = ? AND candidate_key = ?", item.OrgID, item.NotificationKey).
				Assign(map[string]any{
					"project_id":  projectID,
					"kind":        item.Kind,
					"source":      item.Source,
					"source_ref":  item.SourceRef,
					"title":       item.Title,
					"body":        item.Body,
					"severity":    item.Severity,
					"status":      "new",
					"payload":     payload,
					"created_by":  item.CreatedBy,
					"updated_at":  time.Now().UTC(),
					"resolved_at": nil,
				}).
				FirstOrCreate(&candidate).Error; err != nil {
				return err
			}

			notification := inboxNotification{
				OrgID:           item.OrgID,
				ProjectID:       &projectID,
				Kind:            item.Kind,
				Source:          item.Source,
				SourceRef:       &sourceRef,
				NotificationKey: item.NotificationKey,
				Title:           item.Title,
				Body:            item.Body,
				Severity:        item.Severity,
				Status:          "new",
				RouteHint:       defaultString(item.RouteHint, "insight_chat"),
				Payload:         payload,
				CreatedBy:       item.CreatedBy,
			}
			if item.Actor != "" {
				notification.RecipientActor = &item.Actor
			}
			if err := tx.
				Where("org_id = ? AND notification_key = ?", item.OrgID, item.NotificationKey).
				Assign(map[string]any{
					"project_id":      projectID,
					"recipient_actor": nullableString(item.Actor),
					"kind":            item.Kind,
					"source":          item.Source,
					"source_ref":      item.SourceRef,
					"title":           item.Title,
					"body":            item.Body,
					"severity":        item.Severity,
					"route_hint":      defaultString(item.RouteHint, "insight_chat"),
					"payload":         payload,
					"created_by":      item.CreatedBy,
					"updated_at":      time.Now().UTC(),
				}).
				FirstOrCreate(&notification).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) baseInboxQuery(ctx context.Context, orgID uuid.UUID, actor string, projectID *int64) *gorm.DB {
	q := r.db.WithContext(ctx).Model(&inboxNotification{}).Where("org_id = ?", orgID)
	if projectID != nil {
		q = q.Where("project_id = ?", *projectID)
	}
	if actor != "" {
		q = q.Where("(recipient_actor IS NULL OR recipient_actor = ?)", actor)
	}
	return q
}

func nullableString(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

func defaultString(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

func maxInt(v, min int) int {
	if v < min {
		return min
	}
	return v
}

func mapNotificationRows(rows []inboxNotification) []NotificationRecord {
	items := make([]NotificationRecord, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapNotificationRow(row))
	}
	return items
}

func mapNotificationRow(row inboxNotification) NotificationRecord {
	return NotificationRecord{
		ID:              row.ID,
		OrgID:           row.OrgID,
		ProjectID:       row.ProjectID,
		RecipientActor:  row.RecipientActor,
		Kind:            row.Kind,
		Source:          row.Source,
		SourceRef:       row.SourceRef,
		NotificationKey: row.NotificationKey,
		Title:           row.Title,
		Body:            row.Body,
		Severity:        row.Severity,
		Status:          row.Status,
		RouteHint:       row.RouteHint,
		Payload:         decodePayload(row.Payload),
		CreatedBy:       row.CreatedBy,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
		ReadAt:          row.ReadAt,
		DismissedAt:     row.DismissedAt,
	}
}

func decodePayload(raw []byte) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return map[string]any{"_raw": string(raw)}
	}
	return payload
}

func (r *Repository) ResolveStaleApprovals(ctx context.Context, orgID uuid.UUID, actor string, pendingKeys map[string]struct{}) error {
	now := time.Now().UTC()
	q := r.baseInboxQuery(ctx, orgID, actor, nil).
		Where("kind = ? AND source = ?", "approval", "nexus").
		Where("status <> ?", "dismissed")
	if len(pendingKeys) > 0 {
		keys := make([]string, 0, len(pendingKeys))
		for key := range pendingKeys {
			keys = append(keys, key)
		}
		q = q.Where("notification_key NOT IN ?", keys)
	}
	return q.Updates(map[string]any{
		"status":     "read",
		"read_at":    now,
		"updated_at": now,
	}).Error
}
