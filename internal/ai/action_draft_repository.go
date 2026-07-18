package ai

import (
	"context"
	"errors"
	"time"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// actionDraftModel es una fila de ai_action_drafts (migración 000251): el
// borrador staged por el agente que solo pasa a applied tras la verificación
// de Nexus.
type actionDraftModel struct {
	ID             uuid.UUID  `gorm:"column:id;type:uuid;primaryKey"`
	TenantID       uuid.UUID  `gorm:"column:tenant_id;type:uuid;not null"`
	DraftType      string     `gorm:"column:draft_type;not null"`
	Status         string     `gorm:"column:status;not null"`
	PayloadJSON    []byte     `gorm:"column:payload_json;type:jsonb;not null"`
	NexusRequestID string     `gorm:"column:nexus_request_id"`
	CreatedBy      string     `gorm:"column:created_by"`
	AppliedAt      *time.Time `gorm:"column:applied_at"`
	AppliedBy      string     `gorm:"column:applied_by"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at"`
}

func (actionDraftModel) TableName() string {
	return "ai_action_drafts"
}

type actionDraftRepository struct {
	db *gorm.DB
}

func NewActionDraftRepository(db *gorm.DB) *actionDraftRepository {
	return &actionDraftRepository{db: db}
}

func (r *actionDraftRepository) create(ctx context.Context, draft actionDraftModel) (actionDraftModel, error) {
	if r == nil || r.db == nil {
		return actionDraftModel{}, domainerr.Internal("ai action draft repository unavailable")
	}
	if draft.ID == uuid.Nil {
		draft.ID = uuid.New()
	}
	if draft.Status == "" {
		draft.Status = actionDraftStatusStaged
	}
	if len(draft.PayloadJSON) == 0 {
		draft.PayloadJSON = []byte("{}")
	}
	now := time.Now().UTC()
	if draft.CreatedAt.IsZero() {
		draft.CreatedAt = now
	}
	if draft.UpdatedAt.IsZero() {
		draft.UpdatedAt = now
	}
	if err := r.db.WithContext(ctx).Create(&draft).Error; err != nil {
		return actionDraftModel{}, err
	}
	return draft, nil
}

func (r *actionDraftRepository) getByNexusRequestID(ctx context.Context, tenantID uuid.UUID, draftType, nexusRequestID string) (actionDraftModel, error) {
	if r == nil || r.db == nil {
		return actionDraftModel{}, domainerr.Internal("ai action draft repository unavailable")
	}
	var draft actionDraftModel
	err := r.db.WithContext(ctx).
		First(&draft, "tenant_id = ? AND draft_type = ? AND nexus_request_id = ?", tenantID, draftType, nexusRequestID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return actionDraftModel{}, domainerr.NotFound("action draft not found")
	}
	if err != nil {
		return actionDraftModel{}, err
	}
	return draft, nil
}

func (r *actionDraftRepository) markApplied(ctx context.Context, tenantID, draftID uuid.UUID, appliedBy string, appliedAt time.Time) error {
	if r == nil || r.db == nil {
		return domainerr.Internal("ai action draft repository unavailable")
	}
	return r.db.WithContext(ctx).Model(&actionDraftModel{}).
		Where("id = ? AND tenant_id = ?", draftID, tenantID).
		Updates(map[string]any{
			"status":     actionDraftStatusApplied,
			"applied_at": appliedAt,
			"applied_by": appliedBy,
			"updated_at": appliedAt,
		}).Error
}
