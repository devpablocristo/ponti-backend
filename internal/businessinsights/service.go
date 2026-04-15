// Package businessinsights genera candidatos de notificacion a partir de
// eventos de dominio (stock negativo, resultado negativo, etc.). Consulta a
// Nexus Review para decidir si el evento amerita notificar y, si si,
// dedupica via core/notifications/go/candidates.
package businessinsights

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/devpablocristo/core/governance/go/reviewclient"
	corecandidates "github.com/devpablocristo/core/notifications/go/candidates"
	"github.com/google/uuid"
)

// ReviewClient es el subset de reviewclient.Client que usa el service.
type ReviewClient interface {
	SubmitRequest(ctx context.Context, idempotencyKey string, body reviewclient.SubmitRequestBody) (reviewclient.SubmitResponse, error)
}

// Config centraliza thresholds y ventanas de dedup del service.
type Config struct {
	NegativeStockDedupWindow time.Duration
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
	if cfg.NegativeStockDedupWindow <= 0 {
		cfg.NegativeStockDedupWindow = 6 * time.Hour
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
}

// NotifyStockNegative evalua via Nexus si corresponde notificar stock
// negativo y, si la policy matchea, upserta el candidato (dedup por
// fingerprint bucketed). Retorna nil y no hace nada si el stock no es
// negativo o si no hay review client configurado.
func (s *Service) NotifyStockNegative(ctx context.Context, tenantID uuid.UUID, actor string, level StockLevel) error {
	if s == nil || s.review == nil {
		return nil
	}
	if level.Quantity >= 0 {
		return nil
	}

	now := time.Now().UTC()
	fingerprint := bucketedID("ponti.stock.negative", level.ProductID, s.config.NegativeStockDedupWindow, now)

	decision, err := s.review.SubmitRequest(ctx, fingerprint, reviewclient.SubmitRequestBody{
		RequesterType: "service",
		RequesterID:   "ponti-backend",
		ActionType:    "ponti.stock.negative",
		TargetSystem:  "ponti",
		Params: map[string]any{
			"product_id":   level.ProductID,
			"product_name": level.ProductName,
			"quantity":     level.Quantity,
		},
	})
	if err != nil {
		return fmt.Errorf("review submit: %w", err)
	}
	if !policyMatched(decision) {
		return nil
	}

	body := fmt.Sprintf(
		"%s quedo con stock negativo: %s unidades. Hay un desvio entre movimientos y consumos que requiere revision.",
		nonEmpty(level.ProductName, "El insumo"),
		formatNumber(level.Quantity),
	)

	record, shouldNotify, err := s.candidates.Record(ctx, CandidateUpsert{
		TenantID:    tenantID.String(),
		Kind:        "insight",
		EventType:   "ponti.stock.negative",
		EntityType:  "supply",
		EntityID:    level.ProductID,
		Fingerprint: fingerprint,
		Severity:    "warning",
		Title:       "Stock negativo",
		Body:        body,
		Evidence: map[string]any{
			"product_id":        level.ProductID,
			"product_name":      level.ProductName,
			"quantity":          level.Quantity,
			"review_request_id": decision.RequestID,
			"review_policy_hit": decision.DecisionReason,
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
