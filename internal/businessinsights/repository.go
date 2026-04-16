package businessinsights

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	candidatesdomain "github.com/devpablocristo/core/notifications/go/candidates/usecases/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/devpablocristo/ponti-backend/internal/businessinsights/repository/models"
)

// Aliases al tipo del core para evitar duplicacion.
type CandidateUpsert = candidatesdomain.UpsertInput
type CandidateRecord = candidatesdomain.Candidate

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Upsert(ctx context.Context, in CandidateUpsert) (CandidateRecord, bool, error) {
	now := in.Now.UTC()
	if now.IsZero() {
		now = time.Now().UTC()
	}
	tenantID, err := uuid.Parse(in.TenantID)
	if err != nil {
		return CandidateRecord{}, false, fmt.Errorf("parse tenant_id: %w", err)
	}

	var row models.CandidateModel
	err = r.db.WithContext(ctx).
		Where("tenant_id = ? AND fingerprint = ?", tenantID, in.Fingerprint).
		First(&row).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return CandidateRecord{}, false, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row = models.CandidateModel{
			ID:              uuid.New(),
			TenantID:        tenantID,
			Kind:            in.Kind,
			EventType:       in.EventType,
			EntityType:      in.EntityType,
			EntityID:        in.EntityID,
			Fingerprint:     in.Fingerprint,
			Severity:        in.Severity,
			Status:          candidatesdomain.StatusNew,
			Title:           in.Title,
			Body:            in.Body,
			EvidenceJSON:    marshalEvidence(in.Evidence),
			OccurrenceCount: 1,
			FirstSeenAt:     now,
			LastSeenAt:      now,
			LastActor:       in.Actor,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
			return CandidateRecord{}, false, err
		}
		return toCandidateRecord(row), true, nil
	}

	shouldNotify := row.Status != candidatesdomain.StatusNotified
	row.Kind = in.Kind
	row.EventType = in.EventType
	row.EntityType = in.EntityType
	row.EntityID = in.EntityID
	row.Severity = in.Severity
	row.Title = in.Title
	row.Body = in.Body
	row.EvidenceJSON = marshalEvidence(in.Evidence)
	row.LastSeenAt = now
	row.LastActor = in.Actor
	row.UpdatedAt = now
	row.OccurrenceCount++
	reopened := row.Status == candidatesdomain.StatusResolved
	if reopened {
		row.Status = candidatesdomain.StatusNew
		row.ResolvedAt = nil
		shouldNotify = true
	}
	if err := r.db.WithContext(ctx).Save(&row).Error; err != nil {
		return CandidateRecord{}, false, err
	}
	if reopened {
		if err := r.db.WithContext(ctx).
			Where("insight_id = ?", row.ID).
			Delete(&models.ReadModel{}).Error; err != nil {
			return CandidateRecord{}, false, fmt.Errorf("clear reads on reopen: %w", err)
		}
	}
	return toCandidateRecord(row), shouldNotify, nil
}

// ResolveByEntity marca como resueltos todos los candidatos abiertos del tenant
// que matcheen (event_type, entity_type, entity_id). Devuelve cuantos se resolvieron.
// Se usa para auto-resolver desde el dominio (ej: stock vuelve a positivo).
func (r *Repository) ResolveByEntity(ctx context.Context, tenantID, eventType, entityType, entityID, actor string, now time.Time) (int64, error) {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	tID, err := uuid.Parse(tenantID)
	if err != nil {
		return 0, fmt.Errorf("parse tenant_id: %w", err)
	}
	res := r.db.WithContext(ctx).Model(&models.CandidateModel{}).
		Where("tenant_id = ? AND event_type = ? AND entity_type = ? AND entity_id = ? AND status <> ?",
			tID, eventType, entityType, entityID, candidatesdomain.StatusResolved).
		Updates(map[string]any{
			"status":      candidatesdomain.StatusResolved,
			"resolved_at": now.UTC(),
			"last_actor":  actor,
			"updated_at":  now.UTC(),
		})
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

// ResolveByID marca un candidato puntual como resuelto (resolve manual).
func (r *Repository) ResolveByID(ctx context.Context, tenantID, candidateID, actor string, now time.Time) error {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	tID, err := uuid.Parse(tenantID)
	if err != nil {
		return fmt.Errorf("parse tenant_id: %w", err)
	}
	id, err := uuid.Parse(candidateID)
	if err != nil {
		return fmt.Errorf("parse candidate_id: %w", err)
	}
	return r.db.WithContext(ctx).Model(&models.CandidateModel{}).
		Where("id = ? AND tenant_id = ?", id, tID).
		Updates(map[string]any{
			"status":      candidatesdomain.StatusResolved,
			"resolved_at": now.UTC(),
			"last_actor":  actor,
			"updated_at":  now.UTC(),
		}).Error
}

// ReopenByID reactiva un candidato resuelto y limpia las marcas "leida" para
// que vuelva a aparecer no-leida para todos los usuarios.
func (r *Repository) ReopenByID(ctx context.Context, tenantID, candidateID, actor string, now time.Time) error {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	tID, err := uuid.Parse(tenantID)
	if err != nil {
		return fmt.Errorf("parse tenant_id: %w", err)
	}
	id, err := uuid.Parse(candidateID)
	if err != nil {
		return fmt.Errorf("parse candidate_id: %w", err)
	}
	if err := r.db.WithContext(ctx).Model(&models.CandidateModel{}).
		Where("id = ? AND tenant_id = ?", id, tID).
		Updates(map[string]any{
			"status":      candidatesdomain.StatusNotified,
			"resolved_at": nil,
			"last_actor":  actor,
			"updated_at":  now.UTC(),
		}).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).
		Where("insight_id = ?", id).
		Delete(&models.ReadModel{}).Error
}

// MarkRead inserta (o ignora si existe) la marca "leida" del usuario sobre el candidato.
func (r *Repository) MarkRead(ctx context.Context, tenantID, candidateID, userID string, now time.Time) error {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	tID, err := uuid.Parse(tenantID)
	if err != nil {
		return fmt.Errorf("parse tenant_id: %w", err)
	}
	id, err := uuid.Parse(candidateID)
	if err != nil {
		return fmt.Errorf("parse candidate_id: %w", err)
	}
	// Validar que el candidato pertenece al tenant antes de insertar la lectura.
	var exists int64
	if err := r.db.WithContext(ctx).Model(&models.CandidateModel{}).
		Where("id = ? AND tenant_id = ?", id, tID).
		Count(&exists).Error; err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("candidate not found")
	}
	read := models.ReadModel{InsightID: id, UserID: userID, ReadAt: now.UTC()}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "insight_id"}, {Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"read_at"}),
		}).
		Create(&read).Error
}

// MarkUnread borra la marca "leida" del usuario sobre el candidato.
func (r *Repository) MarkUnread(ctx context.Context, tenantID, candidateID, userID string) error {
	tID, err := uuid.Parse(tenantID)
	if err != nil {
		return fmt.Errorf("parse tenant_id: %w", err)
	}
	id, err := uuid.Parse(candidateID)
	if err != nil {
		return fmt.Errorf("parse candidate_id: %w", err)
	}
	var exists int64
	if err := r.db.WithContext(ctx).Model(&models.CandidateModel{}).
		Where("id = ? AND tenant_id = ?", id, tID).
		Count(&exists).Error; err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("candidate not found")
	}
	return r.db.WithContext(ctx).
		Where("insight_id = ? AND user_id = ?", id, userID).
		Delete(&models.ReadModel{}).Error
}

// ListOptions filtra y configura la lectura de candidatos.
type ListOptions struct {
	IncludeResolved bool
	Limit           int
}

// CandidateView es un CandidateRecord enriquecido con el estado de lectura
// del usuario que consulta.
type CandidateView struct {
	CandidateRecord
	ReadAt *time.Time
}

// ListByTenantForUser devuelve candidatos del tenant con `read_at` per-usuario.
// Si IncludeResolved=false (default), filtra los resueltos.
func (r *Repository) ListByTenantForUser(ctx context.Context, tenantID, userID string, opts ListOptions) ([]CandidateView, error) {
	limit := opts.Limit
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	tID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, fmt.Errorf("parse tenant_id: %w", err)
	}
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tID)
	if !opts.IncludeResolved {
		q = q.Where("status <> ?", candidatesdomain.StatusResolved)
	}
	var rows []models.CandidateModel
	if err := q.Order("last_seen_at DESC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []CandidateView{}, nil
	}

	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	var reads []models.ReadModel
	if userID != "" {
		if err := r.db.WithContext(ctx).
			Where("user_id = ? AND insight_id IN ?", userID, ids).
			Find(&reads).Error; err != nil {
			return nil, err
		}
	}
	readByID := make(map[uuid.UUID]time.Time, len(reads))
	for _, rd := range reads {
		readByID[rd.InsightID] = rd.ReadAt
	}

	out := make([]CandidateView, 0, len(rows))
	for _, row := range rows {
		view := CandidateView{CandidateRecord: toCandidateRecord(row)}
		if t, ok := readByID[row.ID]; ok {
			tt := t
			view.ReadAt = &tt
		}
		out = append(out, view)
	}
	return out, nil
}

func (r *Repository) MarkNotified(ctx context.Context, tenantID, candidateID string, notifiedAt time.Time) error {
	if notifiedAt.IsZero() {
		notifiedAt = time.Now().UTC()
	}
	tID, err := uuid.Parse(tenantID)
	if err != nil {
		return fmt.Errorf("parse tenant_id: %w", err)
	}
	id, err := uuid.Parse(candidateID)
	if err != nil {
		return fmt.Errorf("parse candidate_id: %w", err)
	}
	updates := map[string]any{
		"status":           candidatesdomain.StatusNotified,
		"last_notified_at": notifiedAt.UTC(),
		"updated_at":       notifiedAt.UTC(),
	}
	var row models.CandidateModel
	if err := r.db.WithContext(ctx).First(&row, "id = ? AND tenant_id = ?", id, tID).Error; err != nil {
		return err
	}
	if row.FirstNotifiedAt == nil {
		updates["first_notified_at"] = notifiedAt.UTC()
	}
	return r.db.WithContext(ctx).Model(&models.CandidateModel{}).
		Where("id = ? AND tenant_id = ?", id, tID).
		Updates(updates).Error
}

func marshalEvidence(in map[string]any) []byte {
	if len(in) == 0 {
		return []byte("{}")
	}
	raw, err := json.Marshal(in)
	if err != nil || len(raw) == 0 {
		return []byte("{}")
	}
	return raw
}

func toCandidateRecord(row models.CandidateModel) CandidateRecord {
	evidence := map[string]any{}
	if len(row.EvidenceJSON) > 0 {
		_ = json.Unmarshal(row.EvidenceJSON, &evidence)
	}
	return CandidateRecord{
		ID:              row.ID.String(),
		TenantID:        row.TenantID.String(),
		Kind:            row.Kind,
		EventType:       row.EventType,
		EntityType:      row.EntityType,
		EntityID:        row.EntityID,
		Fingerprint:     row.Fingerprint,
		Severity:        row.Severity,
		Status:          row.Status,
		Title:           row.Title,
		Body:            row.Body,
		Evidence:        evidence,
		OccurrenceCount: row.OccurrenceCount,
		FirstSeenAt:     row.FirstSeenAt,
		LastSeenAt:      row.LastSeenAt,
		FirstNotifiedAt: row.FirstNotifiedAt,
		LastNotifiedAt:  row.LastNotifiedAt,
		ResolvedAt:      row.ResolvedAt,
		LastActor:       row.LastActor,
	}
}
