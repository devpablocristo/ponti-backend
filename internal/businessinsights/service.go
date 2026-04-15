// Package businessinsights genera candidatos de notificacion a partir de
// eventos de dominio (stock bajo, resultado negativo, etc.). Consulta a
// Nexus Review para decidir si el evento amerita notificar y, si si,
// dedupica via core/notifications/go/candidates.
package businessinsights

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	corecandidates "github.com/devpablocristo/core/notifications/go/candidates"
	"github.com/devpablocristo/core/governance/go/reviewclient"
	"github.com/google/uuid"
)

// ReviewClient es el subset de reviewclient.Client que usa el service.
type ReviewClient interface {
	SubmitRequest(ctx context.Context, idempotencyKey string, body reviewclient.SubmitRequestBody) (reviewclient.SubmitResponse, error)
}

// Config centraliza thresholds y ventanas de dedup del service.
type Config struct {
	LowStockDedupWindow time.Duration
}

// CandidateRepository es la persistencia de candidatos (implementada en repository.go).
type CandidateRepository interface {
	Upsert(ctx context.Context, in CandidateUpsert) (CandidateRecord, bool, error)
	MarkNotified(ctx context.Context, tenantID, candidateID string, notifiedAt time.Time) error
}

// Service aplica reglas de dominio a eventos y crea candidatos de notificacion.
type Service struct {
	candidates *corecandidates.Usecases
	review     ReviewClient
	config     Config
}

// NewService construye un Service. Si review es nil, el Service degrada
// gracioso (no emite notificaciones).
func NewService(repo CandidateRepository, review ReviewClient, cfg Config) *Service {
	if cfg.LowStockDedupWindow <= 0 {
		cfg.LowStockDedupWindow = 6 * time.Hour
	}
	return &Service{
		candidates: corecandidates.NewWriteUsecases(repo, repo),
		review:     review,
		config:     cfg,
	}
}

// StockLevel describe el estado de stock que motiva la notificacion.
type StockLevel struct {
	ProductID   string
	ProductName string
	Quantity    float64
	MinQuantity float64
}

// NotifyStockLow evalua via Nexus si corresponde notificar stock bajo y, si
// la policy matchea, upserta el candidato (dedup por fingerprint bucketed).
// Retorna nil y no hace nada si no hay review client configurado.
func (s *Service) NotifyStockLow(ctx context.Context, tenantID uuid.UUID, actor string, level StockLevel) error {
	if s == nil || s.review == nil {
		return nil
	}
	if !isLowStock(level) {
		return nil
	}

	now := time.Now().UTC()
	decision, err := s.review.SubmitRequest(ctx, bucketedID("ponti.stock.low", level.ProductID, s.config.LowStockDedupWindow, now), reviewclient.SubmitRequestBody{
		RequesterType: "service",
		RequesterID:   "ponti-backend",
		ActionType:    "ponti.stock.low",
		TargetSystem:  "ponti",
		Params: map[string]any{
			"product_id":   level.ProductID,
			"product_name": level.ProductName,
			"quantity":     level.Quantity,
			"min_quantity": level.MinQuantity,
		},
	})
	if err != nil {
		return fmt.Errorf("review submit: %w", err)
	}
	if !policyMatched(decision) {
		return nil
	}

	body := fmt.Sprintf(
		"%s quedo con stock critico: %s disponibles sobre un minimo de %s. Revisa reposicion y riesgo de quiebre.",
		nonEmpty(level.ProductName, "El insumo"),
		formatNumber(level.Quantity),
		formatNumber(level.MinQuantity),
	)

	record, shouldNotify, err := s.candidates.Record(ctx, CandidateUpsert{
		TenantID:    tenantID.String(),
		Kind:        "insight",
		EventType:   "ponti.stock.low",
		EntityType:  "supply",
		EntityID:    level.ProductID,
		Fingerprint: bucketedID("ponti.stock.low", level.ProductID, s.config.LowStockDedupWindow, now),
		Severity:    "warning",
		Title:       "Stock critico",
		Body:        body,
		Evidence: map[string]any{
			"product_id":         level.ProductID,
			"product_name":       level.ProductName,
			"quantity":           level.Quantity,
			"min_quantity":       level.MinQuantity,
			"review_request_id":  decision.RequestID,
			"review_policy_hit":  decision.DecisionReason,
		},
		Actor: actor,
		Now:   now,
	})
	if err != nil {
		return fmt.Errorf("upsert candidate: %w", err)
	}
	if !shouldNotify {
		return nil
	}
	return s.candidates.MarkNotified(ctx, record.TenantID, record.ID)
}

// policyMatched distingue un allow por policy matcheada vs allow default
// (no policy matched). Nexus serializa la razon como "Policy '<name>'".
func policyMatched(d reviewclient.SubmitResponse) bool {
	return d.Decision == "allow" && strings.HasPrefix(d.DecisionReason, "Policy '")
}

func isLowStock(level StockLevel) bool {
	return level.MinQuantity > 0 && level.Quantity < level.MinQuantity
}

// bucketedID genera un fingerprint que agrupa eventos del mismo entity en
// ventanas fijas (ej: 6h) para evitar spam de notificaciones.
func bucketedID(eventType, entityID string, window time.Duration, now time.Time) string {
	seconds := int64(window / time.Second)
	if seconds <= 0 {
		seconds = int64((6 * time.Hour) / time.Second)
	}
	bucket := int64(math.Floor(float64(now.UTC().Unix()) / float64(seconds)))
	return fmt.Sprintf("%s:%s:%d", eventType, strings.TrimSpace(entityID), bucket)
}

func formatNumber(v float64) string {
	if v == float64(int64(v)) {
		return fmt.Sprintf("%d", int64(v))
	}
	return fmt.Sprintf("%.2f", v)
}

func nonEmpty(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return strings.TrimSpace(v)
}
