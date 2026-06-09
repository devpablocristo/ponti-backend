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

	"github.com/devpablocristo/platform/kernels/governance/go/governanceclient"
	corecandidates "github.com/devpablocristo/platform/notifications/go/candidates"
	"github.com/google/uuid"
)

// ReviewClient es el subset de governanceclient.Client que usa el service.
type ReviewClient interface {
	SubmitRequest(ctx context.Context, idempotencyKey string, body governanceclient.SubmitRequestBody) (governanceclient.SubmitResponse, error)
}

// Config centraliza thresholds y ventanas de dedup del service.
type Config struct {
	NegativeStockDedupWindow time.Duration
	DataIntegrityDedupWindow time.Duration
}

// CandidateRepository es la persistencia de candidatos (implementada en repository.go).
type CandidateRepository interface {
	Upsert(ctx context.Context, in CandidateUpsert) (CandidateRecord, bool, error)
	MarkNotified(ctx context.Context, tenantID, candidateID string, notifiedAt time.Time) error
}

// ResolverRepository ofrece operaciones de cierre/reapertura por entidad o id.
type ResolverRepository interface {
	ResolveByEntity(ctx context.Context, tenantID, eventType, entityType, entityID, actor string, now time.Time) (int64, error)
	ResolveByID(ctx context.Context, tenantID, candidateID, actor string, now time.Time) error
	ReopenByID(ctx context.Context, tenantID, candidateID, actor string, now time.Time) error
}

// ReadRepository es la persistencia per-usuario de "leida".
type ReadRepository interface {
	MarkRead(ctx context.Context, tenantID, candidateID, userID string, now time.Time) error
	MarkUnread(ctx context.Context, tenantID, candidateID, userID string) error
}

// Service aplica reglas de dominio a eventos y crea/cierra candidatos de notificacion.
type Service struct {
	candidates *corecandidates.Usecases
	resolver   ResolverRepository
	reads      ReadRepository
	review     ReviewClient
	config     Config
}

// NewService construye un Service. Si review es nil, el Service degrada
// gracioso (no emite notificaciones).
func NewService(repo CandidateRepository, resolver ResolverRepository, reads ReadRepository, review ReviewClient, cfg Config) *Service {
	if cfg.NegativeStockDedupWindow <= 0 {
		cfg.NegativeStockDedupWindow = 6 * time.Hour
	}
	if cfg.DataIntegrityDedupWindow <= 0 {
		cfg.DataIntegrityDedupWindow = 6 * time.Hour
	}
	return &Service{
		candidates: corecandidates.NewWriteUsecases(repo, repo),
		resolver:   resolver,
		reads:      reads,
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

// DataIntegrityControlIssue resume un control de integridad fallido con la
// evidencia minima necesaria para explicar el desvio.
type DataIntegrityControlIssue struct {
	ControlNumber int
	DataToVerify  string
	Description   string
	DifferenceA   string
	DifferenceB   string
	Tolerance     string
	SystemSource  string
	RecalcASource string
	RecalcBSource string
}

// DataIntegrityCritical describe un resultado de integridad con controles en
// error. Se registra como insight read-only para que Axis lo explique con
// evidencia, sin ejecutar acciones.
type DataIntegrityCritical struct {
	ProjectID    string
	FailedChecks int
	TotalChecks  int
	Controls     []DataIntegrityControlIssue
}

// OperatingResultNegativeCrop resume un cultivo con resultado operativo
// negativo dentro del reporte de resumen.
type OperatingResultNegativeCrop struct {
	CropID             string
	CropName           string
	OperatingResultUSD string
	SurfaceHa          string
	ReturnPct          string
}

// OperatingResultNegative describe un reporte de resumen con resultado
// operativo total negativo. Axis lo consume como insight read-only.
type OperatingResultNegative struct {
	ProjectID               string
	CustomerID              string
	CampaignID              string
	TotalOperatingResultUSD string
	ProjectReturnPct        string
	TotalInvestedProjectUSD string
	NegativeCrops           []OperatingResultNegativeCrop
}

// TentativePriceItem resume un insumo con precio tentativo.
type TentativePriceItem struct {
	SupplyID     string
	Name         string
	CategoryName string
	Price        string
}

// TentativePricesIssue describe insumos con precio parcial/tentativo dentro
// de un workspace. Impacta directamente costos, reportes y decisiones.
type TentativePricesIssue struct {
	ProjectID   string
	CustomerID  string
	CampaignID  string
	FieldID     string
	Count       int64
	SampleItems []TentativePriceItem
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

	decision, err := s.review.SubmitRequest(ctx, fingerprint, governanceclient.SubmitRequestBody{
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
			"suggested_action":  "review_stock_movements",
			"source_ref":        "ponti.stock.real",
			"workspace":         map[string]any{},
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

// MaybeResolveStockNegative se llama cuando el stock real de un producto
// vuelve a ser >= 0; cierra automaticamente cualquier notificacion abierta de
// "stock negativo" para ese producto en el tenant.
func (s *Service) MaybeResolveStockNegative(ctx context.Context, tenantID uuid.UUID, productID string) error {
	if s == nil || s.resolver == nil {
		return nil
	}
	if strings.TrimSpace(productID) == "" {
		return nil
	}
	_, err := s.resolver.ResolveByEntity(
		ctx,
		tenantID.String(),
		"ponti.stock.negative",
		"supply",
		strings.TrimSpace(productID),
		"system",
		time.Now().UTC(),
	)
	return err
}

// NotifyDataIntegrityCritical registra un candidato cuando los controles de
// integridad detectan diferencias fuera de tolerancia para un proyecto.
func (s *Service) NotifyDataIntegrityCritical(ctx context.Context, tenantID uuid.UUID, actor string, issue DataIntegrityCritical) error {
	if s == nil || s.candidates == nil {
		return nil
	}
	if issue.FailedChecks <= 0 {
		return s.MaybeResolveDataIntegrityCritical(ctx, tenantID, issue.ProjectID)
	}

	now := time.Now().UTC()
	projectID := nonEmpty(issue.ProjectID, "unknown")
	fingerprint := bucketedID("ponti.data_integrity.critical", projectID, s.config.DataIntegrityDedupWindow, now)
	body := fmt.Sprintf(
		"Se detectaron %d controles de integridad con diferencias fuera de tolerancia sobre %d controles evaluados. Revisar la consistencia entre dashboard, lotes, informes, stock y ordenes.",
		issue.FailedChecks,
		issue.TotalChecks,
	)

	record, shouldNotify, err := s.candidates.Record(ctx, CandidateUpsert{
		TenantID:    tenantID.String(),
		Kind:        "integrity",
		EventType:   "ponti.data_integrity.critical",
		EntityType:  "project",
		EntityID:    projectID,
		Fingerprint: fingerprint,
		Severity:    dataIntegritySeverity(issue.FailedChecks),
		Title:       "Integridad de datos crítica",
		Body:        body,
		Evidence: map[string]any{
			"project_id":       projectID,
			"failed_checks":    issue.FailedChecks,
			"total_checks":     issue.TotalChecks,
			"failed_controls":  issue.Controls,
			"suggested_action": "review_data_integrity",
			"source_ref":       "ponti.data_integrity.costs_check",
			"workspace": map[string]any{
				"project_id": projectID,
			},
		},
		Actor: actor,
		Now:   now,
	})
	if err != nil {
		return fmt.Errorf("upsert data integrity candidate: %w", err)
	}
	if !shouldNotify {
		return nil
	}
	return s.candidates.MarkNotified(ctx, record.TenantID, record.ID)
}

// MaybeResolveDataIntegrityCritical cierra automaticamente el insight de
// integridad cuando el mismo proyecto vuelve a pasar todos los controles.
func (s *Service) MaybeResolveDataIntegrityCritical(ctx context.Context, tenantID uuid.UUID, projectID string) error {
	if s == nil || s.resolver == nil {
		return nil
	}
	if strings.TrimSpace(projectID) == "" {
		return nil
	}
	_, err := s.resolver.ResolveByEntity(
		ctx,
		tenantID.String(),
		"ponti.data_integrity.critical",
		"project",
		strings.TrimSpace(projectID),
		"system",
		time.Now().UTC(),
	)
	return err
}

// NotifyOperatingResultNegative registra un insight cuando el resumen de
// resultados expone resultado operativo total negativo.
func (s *Service) NotifyOperatingResultNegative(ctx context.Context, tenantID uuid.UUID, actor string, issue OperatingResultNegative) error {
	if s == nil || s.candidates == nil {
		return nil
	}
	projectID := nonEmpty(issue.ProjectID, "")
	if projectID == "" {
		return nil
	}

	now := time.Now().UTC()
	fingerprint := bucketedID("ponti.report.operating_result.negative", projectID, s.config.DataIntegrityDedupWindow, now)
	body := fmt.Sprintf(
		"El proyecto tiene resultado operativo negativo: USD %s. Revisar ingresos, costos directos, renta, estructura y cultivos con desvio.",
		issue.TotalOperatingResultUSD,
	)
	record, shouldNotify, err := s.candidates.Record(ctx, CandidateUpsert{
		TenantID:    tenantID.String(),
		Kind:        "margin",
		EventType:   "ponti.report.operating_result.negative",
		EntityType:  "project",
		EntityID:    projectID,
		Fingerprint: fingerprint,
		Severity:    "warning",
		Title:       "Resultado operativo negativo",
		Body:        body,
		Evidence: map[string]any{
			"project_id":                 projectID,
			"customer_id":                issue.CustomerID,
			"campaign_id":                issue.CampaignID,
			"total_operating_result_usd": issue.TotalOperatingResultUSD,
			"project_return_pct":         issue.ProjectReturnPct,
			"total_invested_project_usd": issue.TotalInvestedProjectUSD,
			"negative_crops":             issue.NegativeCrops,
			"suggested_action":           "review_summary_results",
			"source_ref":                 "ponti.reports.summary_results",
			"workspace": map[string]any{
				"customer_id":  issue.CustomerID,
				"project_id":   projectID,
				"campaign_id":  issue.CampaignID,
			},
		},
		Actor: actor,
		Now:   now,
	})
	if err != nil {
		return fmt.Errorf("upsert operating result candidate: %w", err)
	}
	if !shouldNotify {
		return nil
	}
	return s.candidates.MarkNotified(ctx, record.TenantID, record.ID)
}

// MaybeResolveOperatingResultNegative cierra el insight cuando el resultado
// operativo total deja de ser negativo.
func (s *Service) MaybeResolveOperatingResultNegative(ctx context.Context, tenantID uuid.UUID, projectID string) error {
	if s == nil || s.resolver == nil {
		return nil
	}
	if strings.TrimSpace(projectID) == "" {
		return nil
	}
	_, err := s.resolver.ResolveByEntity(
		ctx,
		tenantID.String(),
		"ponti.report.operating_result.negative",
		"project",
		strings.TrimSpace(projectID),
		"system",
		time.Now().UTC(),
	)
	return err
}

// NotifyTentativePrices registra un insight cuando existen insumos con precio
// tentativo dentro del proyecto consultado.
func (s *Service) NotifyTentativePrices(ctx context.Context, tenantID uuid.UUID, actor string, issue TentativePricesIssue) error {
	if s == nil || s.candidates == nil {
		return nil
	}
	projectID := nonEmpty(issue.ProjectID, "")
	if projectID == "" {
		return nil
	}
	if issue.Count <= 0 {
		return s.MaybeResolveTentativePrices(ctx, tenantID, projectID)
	}

	now := time.Now().UTC()
	fingerprint := bucketedID("ponti.data_integrity.tentative_prices", projectID, s.config.DataIntegrityDedupWindow, now)
	body := fmt.Sprintf(
		"Hay %d insumos con precio tentativo. Los costos y reportes pueden estar distorsionados hasta confirmar esos precios.",
		issue.Count,
	)
	record, shouldNotify, err := s.candidates.Record(ctx, CandidateUpsert{
		TenantID:    tenantID.String(),
		Kind:        "integrity",
		EventType:   "ponti.data_integrity.tentative_prices",
		EntityType:  "project",
		EntityID:    projectID,
		Fingerprint: fingerprint,
		Severity:    "warning",
		Title:       "Precios tentativos en insumos",
		Body:        body,
		Evidence: map[string]any{
			"project_id":       projectID,
			"customer_id":      issue.CustomerID,
			"campaign_id":      issue.CampaignID,
			"field_id":         issue.FieldID,
			"count":            issue.Count,
			"sample_items":     issue.SampleItems,
			"suggested_action": "review_tentative_prices",
			"source_ref":       "ponti.data_integrity.tentative_prices",
			"workspace": map[string]any{
				"customer_id": issue.CustomerID,
				"project_id":  projectID,
				"campaign_id": issue.CampaignID,
				"field_id":    issue.FieldID,
			},
		},
		Actor: actor,
		Now:   now,
	})
	if err != nil {
		return fmt.Errorf("upsert tentative prices candidate: %w", err)
	}
	if !shouldNotify {
		return nil
	}
	return s.candidates.MarkNotified(ctx, record.TenantID, record.ID)
}

// MaybeResolveTentativePrices cierra el insight cuando ya no existen precios
// tentativos para el proyecto.
func (s *Service) MaybeResolveTentativePrices(ctx context.Context, tenantID uuid.UUID, projectID string) error {
	if s == nil || s.resolver == nil {
		return nil
	}
	if strings.TrimSpace(projectID) == "" {
		return nil
	}
	_, err := s.resolver.ResolveByEntity(
		ctx,
		tenantID.String(),
		"ponti.data_integrity.tentative_prices",
		"project",
		strings.TrimSpace(projectID),
		"system",
		time.Now().UTC(),
	)
	return err
}

// MarkRead persiste que el usuario leyo el candidato.
func (s *Service) MarkRead(ctx context.Context, tenantID uuid.UUID, candidateID, userID string) error {
	if s == nil || s.reads == nil {
		return nil
	}
	return s.reads.MarkRead(ctx, tenantID.String(), candidateID, userID, time.Now().UTC())
}

// MarkUnread borra la marca de lectura del usuario.
func (s *Service) MarkUnread(ctx context.Context, tenantID uuid.UUID, candidateID, userID string) error {
	if s == nil || s.reads == nil {
		return nil
	}
	return s.reads.MarkUnread(ctx, tenantID.String(), candidateID, userID)
}

// ResolveManual marca un candidato como resuelto por un usuario.
func (s *Service) ResolveManual(ctx context.Context, tenantID uuid.UUID, candidateID, actor string) error {
	if s == nil || s.resolver == nil {
		return nil
	}
	return s.resolver.ResolveByID(ctx, tenantID.String(), candidateID, actor, time.Now().UTC())
}

// Reopen reactiva un candidato resuelto y limpia las marcas de lectura.
func (s *Service) Reopen(ctx context.Context, tenantID uuid.UUID, candidateID, actor string) error {
	if s == nil || s.resolver == nil {
		return nil
	}
	return s.resolver.ReopenByID(ctx, tenantID.String(), candidateID, actor, time.Now().UTC())
}

// policyMatched indica si Nexus permitio la accion. En la primera respuesta
// el reason viene como "Policy '<name>'"; en replays idempotentes el reason
// llega vacio pero el decision/status sigue siendo allow/allowed. Nexus
// rechaza por default si no matchea policy, asi que `decision=allow` ya es
// señal suficiente para nuestro caso (1 sola policy por action_type).
func policyMatched(d governanceclient.SubmitResponse) bool {
	if d.Decision != "allow" {
		return false
	}
	if strings.HasPrefix(d.DecisionReason, "Policy '") {
		return true
	}
	// Replay idempotente: Nexus solo devuelve reason vacio si la request
	// original ya habia sido aprobada por una policy.
	return d.DecisionReason == ""
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

func dataIntegritySeverity(failedChecks int) string {
	if failedChecks >= 3 {
		return "critical"
	}
	return "warning"
}
