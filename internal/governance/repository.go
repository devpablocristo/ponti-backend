package governance

import (
	"context"
	"errors"
	"time"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	"github.com/google/uuid"
	"gorm.io/gorm"

	nexusclient "github.com/devpablocristo/ponti-backend/internal/nexus"
)

// RequestRecord es el espejo local de una request gobernada por Nexus.
type RequestRecord struct {
	ID                uuid.UUID  `gorm:"column:id;type:uuid;primaryKey"`
	TenantID          uuid.UUID  `gorm:"column:tenant_id;type:uuid;not null"`
	NexusRequestID    string     `gorm:"column:nexus_request_id;not null"`
	ActionType        string     `gorm:"column:action_type;not null"`
	Origin            string     `gorm:"column:origin;not null"`
	RequesterID       string     `gorm:"column:requester_id"`
	Status            string     `gorm:"column:status;not null"`
	Decision          string     `gorm:"column:decision"`
	RiskLevel         string     `gorm:"column:risk_level"`
	BindingHash       string     `gorm:"column:binding_hash"`
	ActionBindingJSON []byte     `gorm:"column:action_binding_json;type:jsonb"`
	ParamsJSON        []byte     `gorm:"column:params_json;type:jsonb"`
	PayloadJSON       []byte     `gorm:"column:payload_json;type:jsonb"`
	EntityType        string     `gorm:"column:entity_type"`
	EntityID          string     `gorm:"column:entity_id"`
	ApprovalID        string     `gorm:"column:approval_id"`
	DecidedBy         string     `gorm:"column:decided_by"`
	ExecutedAt        *time.Time `gorm:"column:executed_at"`
	ResultJSON        []byte     `gorm:"column:result_json;type:jsonb"`
	ErrorMessage      string     `gorm:"column:error_message"`
	AxisRunID         string     `gorm:"column:axis_run_id"`
	AxisTaskID        string     `gorm:"column:axis_task_id"`
	PontiRequestID    string     `gorm:"column:ponti_request_id"`
	CreatedAt         time.Time  `gorm:"column:created_at"`
	UpdatedAt         time.Time  `gorm:"column:updated_at"`
}

func (RequestRecord) TableName() string {
	return "ai_governance_requests"
}

// EvidenceRecord es el cache local de un evidence pack firmado de Nexus.
type EvidenceRecord struct {
	ID             uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	TenantID       uuid.UUID `gorm:"column:tenant_id;type:uuid;not null"`
	NexusRequestID string    `gorm:"column:nexus_request_id;not null"`
	PackJSON       []byte    `gorm:"column:pack_json;type:jsonb;not null"`
	SignatureKeyID string    `gorm:"column:signature_key_id"`
	Signature      string    `gorm:"column:signature"`
	RetrievedAt    time.Time `gorm:"column:retrieved_at"`
}

func (EvidenceRecord) TableName() string {
	return "ai_evidence_packs"
}

// Repository persiste el espejo local de governance en Postgres.
type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetByNexusRequestID(ctx context.Context, tenantID uuid.UUID, nexusRequestID string) (RequestRecord, error) {
	if r == nil || r.db == nil {
		return RequestRecord{}, domainerr.Internal("governance repository unavailable")
	}
	var row RequestRecord
	err := r.db.WithContext(ctx).
		First(&row, "tenant_id = ? AND nexus_request_id = ?", tenantID, nexusRequestID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return RequestRecord{}, domainerr.NotFound("governance request not found")
	}
	if err != nil {
		return RequestRecord{}, err
	}
	return row, nil
}

func (r *Repository) Create(ctx context.Context, row RequestRecord) error {
	if r == nil || r.db == nil {
		return domainerr.Internal("governance repository unavailable")
	}
	if row.ID == uuid.Nil {
		row.ID = uuid.New()
	}
	now := time.Now().UTC()
	if row.CreatedAt.IsZero() {
		row.CreatedAt = now
	}
	if row.UpdatedAt.IsZero() {
		row.UpdatedAt = now
	}
	return r.db.WithContext(ctx).Create(&row).Error
}

func (r *Repository) UpdateByNexusRequestID(ctx context.Context, tenantID uuid.UUID, nexusRequestID string, updates map[string]any) error {
	if r == nil || r.db == nil {
		return domainerr.Internal("governance repository unavailable")
	}
	if updates == nil {
		updates = map[string]any{}
	}
	updates["updated_at"] = time.Now().UTC()
	return r.db.WithContext(ctx).Model(&RequestRecord{}).
		Where("tenant_id = ? AND nexus_request_id = ?", tenantID, nexusRequestID).
		Updates(updates).Error
}

// ListHistory devuelve las requests locales ya resueltas (status != pending_approval).
func (r *Repository) ListHistory(ctx context.Context, tenantID uuid.UUID, limit int) ([]RequestRecord, error) {
	if r == nil || r.db == nil {
		return nil, domainerr.Internal("governance repository unavailable")
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var rows []RequestRecord
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND status <> ?", tenantID, nexusclient.StatusPendingApproval).
		Order("updated_at DESC").
		Limit(limit).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repository) GetEvidence(ctx context.Context, tenantID uuid.UUID, nexusRequestID string) (EvidenceRecord, error) {
	if r == nil || r.db == nil {
		return EvidenceRecord{}, domainerr.Internal("governance repository unavailable")
	}
	var row EvidenceRecord
	err := r.db.WithContext(ctx).
		First(&row, "tenant_id = ? AND nexus_request_id = ?", tenantID, nexusRequestID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return EvidenceRecord{}, domainerr.NotFound("evidence pack not found")
	}
	if err != nil {
		return EvidenceRecord{}, err
	}
	return row, nil
}

func (r *Repository) SaveEvidence(ctx context.Context, row EvidenceRecord) error {
	if r == nil || r.db == nil {
		return domainerr.Internal("governance repository unavailable")
	}
	existing, err := r.GetEvidence(ctx, row.TenantID, row.NexusRequestID)
	if err != nil && !domainerr.IsNotFound(err) {
		return err
	}
	now := time.Now().UTC()
	if domainerr.IsNotFound(err) {
		if row.ID == uuid.Nil {
			row.ID = uuid.New()
		}
		if row.RetrievedAt.IsZero() {
			row.RetrievedAt = now
		}
		return r.db.WithContext(ctx).Create(&row).Error
	}
	return r.db.WithContext(ctx).Model(&EvidenceRecord{}).
		Where("id = ? AND tenant_id = ?", existing.ID, row.TenantID).
		Updates(map[string]any{
			"pack_json":        row.PackJSON,
			"signature_key_id": row.SignatureKeyID,
			"signature":        row.Signature,
			"retrieved_at":     now,
		}).Error
}
