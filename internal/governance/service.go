// Package governance integra Ponti core con Nexus Governance: recibe los
// callbacks de approvals (HMAC-verificados), mantiene el espejo local en
// ai_governance_requests y expone el inbox de approvals tenant-scoped.
package governance

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	"github.com/google/uuid"

	nexusclient "github.com/devpablocristo/ponti-backend/internal/nexus"
)

const (
	EventApprovalPending  = "approval_pending"
	EventApprovalResolved = "approval_resolved"

	originAgent   = "agent"
	originWatcher = "watcher"

	expiringSoonWindow = 15 * time.Minute

	// callbackTimestampMaxSkew limita |now - X-Nexus-Callback-Timestamp| para
	// cortar el replay de callbacks capturados. Nexus re-firma cada retry del
	// outbox con timestamp fresco, así que no rompe los reintentos legítimos.
	callbackTimestampMaxSkew = 5 * time.Minute
)

// CallbackEvent es el payload de los webhooks approval_pending/approval_resolved
// de Nexus (espejo de nexus/internal/callbacks.ApprovalEvent).
type CallbackEvent struct {
	Event          string  `json:"event"`
	ApprovalID     string  `json:"approval_id,omitempty"`
	OrgID          string  `json:"org_id,omitempty"`
	RequestID      string  `json:"request_id"`
	Decision       string  `json:"decision,omitempty"`
	DecidedBy      string  `json:"decided_by,omitempty"`
	DecisionNote   string  `json:"decision_note,omitempty"`
	ActionType     string  `json:"action_type,omitempty"`
	TargetResource string  `json:"target_resource,omitempty"`
	Reason         string  `json:"reason,omitempty"`
	RiskLevel      string  `json:"risk_level,omitempty"`
	AISummary      *string `json:"ai_summary,omitempty"`
	CreatedAt      string  `json:"created_at,omitempty"`
	ExpiresAt      *string `json:"expires_at,omitempty"`
	DecidedAt      *string `json:"decided_at,omitempty"`
}

// ApprovalItem es el shape del inbox que consume el FE (/api/v1/ai/approvals).
type ApprovalItem struct {
	RequestID         string                         `json:"request_id"`
	ApprovalID        string                         `json:"approval_id,omitempty"`
	ActionType        string                         `json:"action_type"`
	Status            string                         `json:"status"`
	RiskLevel         string                         `json:"risk_level"`
	RequestedBy       string                         `json:"requested_by"`
	Reason            string                         `json:"reason"`
	CreatedAt         string                         `json:"created_at"`
	ExpiresAt         string                         `json:"expires_at,omitempty"`
	Params            map[string]any                 `json:"params,omitempty"`
	Decisions         []nexusclient.ApprovalDecision `json:"decisions"`
	RequiredApprovals int                            `json:"required_approvals,omitempty"`
	CurrentApprovals  int                            `json:"current_approvals,omitempty"`
}

// ApprovalsSummary alimenta el badge del inbox.
type ApprovalsSummary struct {
	PendingCount      int `json:"pending_count"`
	ExpiringSoonCount int `json:"expiring_soon_count"`
}

// NexusPort abstrae el cliente Nexus para testeo.
type NexusPort interface {
	Get(ctx context.Context, requestID string, opts ...nexusclient.RequestOption) (nexusclient.Request, int, error)
	ListRequests(ctx context.Context, query string, opts ...nexusclient.RequestOption) ([]nexusclient.Request, error)
	ListPendingApprovals(ctx context.Context, query string, opts ...nexusclient.RequestOption) ([]nexusclient.Approval, error)
	Approve(ctx context.Context, approvalID, decidedBy, note string, opts ...nexusclient.RequestOption) (int, []byte, error)
	Reject(ctx context.Context, approvalID, decidedBy, note string, opts ...nexusclient.RequestOption) (int, []byte, error)
	GetEvidence(ctx context.Context, requestID string, opts ...nexusclient.RequestOption) ([]byte, int, error)
}

// RepositoryPort abstrae la persistencia local para testeo.
type RepositoryPort interface {
	GetByNexusRequestID(ctx context.Context, tenantID uuid.UUID, nexusRequestID string) (RequestRecord, error)
	Create(ctx context.Context, row RequestRecord) error
	UpdateByNexusRequestID(ctx context.Context, tenantID uuid.UUID, nexusRequestID string, updates map[string]any) error
	ListHistory(ctx context.Context, tenantID uuid.UUID, limit int) ([]RequestRecord, error)
	GetEvidence(ctx context.Context, tenantID uuid.UUID, nexusRequestID string) (EvidenceRecord, error)
	SaveEvidence(ctx context.Context, row EvidenceRecord) error
}

// Executor es el hook de ejecución post-aprobación; Ola B lo implementa.
type Executor interface {
	ExecuteApproved(ctx context.Context, row RequestRecord) error
}

// NoopExecutor placeholder hasta el executor real de Ola B.
type NoopExecutor struct{}

func (NoopExecutor) ExecuteApproved(context.Context, RequestRecord) error { return nil }

// Config agrupa la configuración del módulo (subset de config.Nexus).
type Config struct {
	CallbackToken string
}

// Service orquesta callbacks de Nexus e inbox de approvals.
type Service struct {
	repo     RepositoryPort
	nexus    NexusPort
	cfg      Config
	executor Executor
}

func NewService(repo RepositoryPort, nx NexusPort, cfg Config, executor Executor) *Service {
	if executor == nil {
		executor = NoopExecutor{}
	}
	return &Service{repo: repo, nexus: nx, cfg: cfg, executor: executor}
}

// CallbackConfigured indica si NEXUS_CALLBACK_TOKEN está seteado.
func (s *Service) CallbackConfigured() bool {
	return strings.TrimSpace(s.cfg.CallbackToken) != ""
}

// VerifyCallbackSignature replica exactamente el cómputo del publisher de
// Nexus (callbacks/outbox.go signCallback): HMAC-SHA256(token, timestamp +
// "." + payload) en hex con prefijo "sha256=", comparado en tiempo constante.
func (s *Service) VerifyCallbackSignature(timestamp string, payload []byte, signature string) bool {
	if !s.CallbackConfigured() || signature == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(strings.TrimSpace(s.cfg.CallbackToken)))
	_, _ = mac.Write([]byte(timestamp))
	_, _ = mac.Write([]byte("."))
	_, _ = mac.Write(payload)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

// VerifyCallbackTimestamp valida la frescura de X-Nexus-Callback-Timestamp:
// Nexus firma time.Now().UTC().Format(time.RFC3339Nano) (RFC3339 también se
// acepta). Rechaza timestamps ausentes, no parseables o fuera de
// ±callbackTimestampMaxSkew; sin esto cualquier (timestamp, payload, firma)
// capturado validaría para siempre.
func (s *Service) VerifyCallbackTimestamp(timestamp string) bool {
	ts, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(timestamp))
	if err != nil {
		return false
	}
	skew := time.Since(ts)
	if skew < 0 {
		skew = -skew
	}
	return skew <= callbackTimestampMaxSkew
}

// HandleCallback procesa un evento de approval ya autenticado. Es idempotente
// por (approval_id, event): si el evento ya fue aplicado retorna nil sin efecto.
func (s *Service) HandleCallback(ctx context.Context, event CallbackEvent) error {
	requestID := strings.TrimSpace(event.RequestID)
	if requestID == "" {
		return domainerr.Validation("request_id is required")
	}
	tenantID, err := uuid.Parse(strings.TrimSpace(event.OrgID))
	if err != nil || tenantID == uuid.Nil {
		// Sin tenant mapeable no hay fila local posible: ack sin efecto para
		// no envenenar los reintentos del outbox de Nexus.
		return nil
	}
	switch event.Event {
	case EventApprovalPending:
		return s.applyPending(ctx, tenantID, requestID, event)
	case EventApprovalResolved:
		return s.applyResolved(ctx, tenantID, requestID, event)
	default:
		return nil
	}
}

func (s *Service) applyPending(ctx context.Context, tenantID uuid.UUID, requestID string, event CallbackEvent) error {
	row, err := s.getOrCreateRow(ctx, tenantID, requestID, event)
	if err != nil {
		return err
	}
	if isTerminalStatus(row.Status) {
		// Callback pending tardío o replayed: nunca degradar un estado terminal
		// a pending_approval (cierra la cadena replay -> revert -> re-ejecución).
		return nil
	}
	if row.ApprovalID == event.ApprovalID && row.Status == nexusclient.StatusPendingApproval {
		return nil // evento ya aplicado
	}
	updates := map[string]any{
		"status":      nexusclient.StatusPendingApproval,
		"approval_id": event.ApprovalID,
	}
	if event.RiskLevel != "" {
		updates["risk_level"] = event.RiskLevel
	}
	if event.ActionType != "" {
		updates["action_type"] = event.ActionType
	}
	return s.repo.UpdateByNexusRequestID(ctx, tenantID, requestID, updates)
}

func (s *Service) applyResolved(ctx context.Context, tenantID uuid.UUID, requestID string, event CallbackEvent) error {
	status, err := resolvedStatus(event.Decision)
	if err != nil {
		return err
	}
	row, err := s.getOrCreateRow(ctx, tenantID, requestID, event)
	if err != nil {
		return err
	}
	decidedBy := strings.TrimSpace(event.DecidedBy)
	if row.Status == status && row.DecidedBy == decidedBy {
		return nil // evento ya aplicado
	}
	updates := map[string]any{
		"status":     status,
		"decided_by": decidedBy,
	}
	if event.ApprovalID != "" {
		updates["approval_id"] = event.ApprovalID
	}
	if err := s.repo.UpdateByNexusRequestID(ctx, tenantID, requestID, updates); err != nil {
		return err
	}
	if status == nexusclient.StatusApproved {
		row.Status = status
		row.DecidedBy = decidedBy
		// Hook Ola B: el executor real ejecutará la acción aprobada (hoy no-op).
		return s.executor.ExecuteApproved(ctx, row)
	}
	return nil
}

// ownedByTenant indica si un item devuelto por Nexus pertenece al tenant.
// Defensa en profundidad: con una API key cross_org Nexus puede devolver datos
// de otras orgs, y Ponti nunca debe exponerlos. OrgID vacío se trata como NO
// propio (se refuta en vez de asumir ownership).
func ownedByTenant(orgID string, tenantID uuid.UUID) bool {
	return orgID == tenantID.String()
}

// isTerminalStatus indica si un status local ya es definitivo y no debe ser
// pisado por un approval_pending fuera de orden o replayed.
func isTerminalStatus(status string) bool {
	switch status {
	case nexusclient.StatusApproved, nexusclient.StatusRejected, nexusclient.StatusExpired,
		nexusclient.StatusExecuted, nexusclient.StatusFailed, nexusclient.StatusCancelled:
		return true
	default:
		return false
	}
}

// resolvedStatus mapea el decision del evento resolved al status local.
func resolvedStatus(decision string) (string, error) {
	switch strings.TrimSpace(decision) {
	case nexusclient.StatusApproved:
		return nexusclient.StatusApproved, nil
	case nexusclient.StatusRejected:
		return nexusclient.StatusRejected, nil
	case nexusclient.StatusExpired:
		return nexusclient.StatusExpired, nil
	default:
		return "", domainerr.Validation(fmt.Sprintf("unknown approval decision %q", decision))
	}
}

// getOrCreateRow devuelve la fila local del request; si no existe la crea con
// origin=agent, hidratada con los detalles de Nexus (Get con WithTenantID).
func (s *Service) getOrCreateRow(ctx context.Context, tenantID uuid.UUID, requestID string, event CallbackEvent) (RequestRecord, error) {
	row, err := s.repo.GetByNexusRequestID(ctx, tenantID, requestID)
	if err == nil {
		return row, nil
	}
	if !domainerr.IsNotFound(err) {
		return RequestRecord{}, err
	}
	row = RequestRecord{
		TenantID:       tenantID,
		NexusRequestID: requestID,
		ActionType:     event.ActionType,
		Origin:         originAgent,
		Status:         nexusclient.StatusPendingApproval,
		RiskLevel:      event.RiskLevel,
	}
	if s.nexus != nil {
		req, st, ferr := s.nexus.Get(ctx, requestID, nexusclient.WithTenantID(tenantID.String()))
		if ferr != nil {
			// Nexus inalcanzable: devolver error para que el outbox reintente.
			return RequestRecord{}, domainerr.UpstreamError("nexus get request failed")
		}
		if st == http.StatusOK && ownedByTenant(req.OrgID, tenantID) {
			row.ActionType = req.ActionType
			row.RequesterID = req.RequesterID
			row.Status = req.Status
			row.Decision = req.Decision
			row.RiskLevel = req.RiskLevel
			row.BindingHash = req.BindingHash
			row.ActionBindingJSON = marshalJSON(req.ActionBinding)
			row.ParamsJSON = marshalJSON(req.Params)
			row.PayloadJSON = marshalJSON(req)
		}
	}
	if row.ActionType == "" {
		row.ActionType = "unknown"
	}
	if err := s.repo.Create(ctx, row); err != nil {
		return RequestRecord{}, err
	}
	return row, nil
}

// --- Inbox ---

// ListApprovals devuelve el inbox: pending lee LIVE de Nexus, history lee el
// espejo local resuelto.
func (s *Service) ListApprovals(ctx context.Context, tenantID uuid.UUID, status string, limit int) ([]ApprovalItem, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	switch strings.TrimSpace(status) {
	case "", "pending":
		return s.listPending(ctx, tenantID, limit)
	case "history":
		return s.listHistory(ctx, tenantID, limit)
	default:
		return nil, domainerr.Validation("status must be pending or history")
	}
}

func (s *Service) listPending(ctx context.Context, tenantID uuid.UUID, limit int) ([]ApprovalItem, error) {
	if s.nexus == nil {
		return nil, domainerr.Unavailable("nexus client not configured")
	}
	tenant := nexusclient.WithTenantID(tenantID.String())
	reqs, err := s.nexus.ListRequests(ctx, fmt.Sprintf("status=%s&limit=%d", nexusclient.StatusPendingApproval, limit), tenant)
	if err != nil {
		return nil, domainerr.UpstreamError("nexus list requests failed")
	}
	approvals, err := s.nexus.ListPendingApprovals(ctx, fmt.Sprintf("limit=%d", limit), tenant)
	if err != nil {
		return nil, domainerr.UpstreamError("nexus list pending approvals failed")
	}
	byRequest := make(map[string]nexusclient.Approval, len(approvals))
	for _, a := range approvals {
		if !ownedByTenant(a.OrgID, tenantID) {
			continue
		}
		byRequest[a.RequestID] = a
	}
	items := make([]ApprovalItem, 0, len(reqs))
	for _, req := range reqs {
		if !ownedByTenant(req.OrgID, tenantID) {
			// Nexus puede devolver items de otras orgs con una key cross_org:
			// nunca exponerlos en el inbox del tenant.
			continue
		}
		item := approvalItemFromRequest(req)
		if a, ok := byRequest[req.ID]; ok {
			applyApproval(&item, a)
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *Service) listHistory(ctx context.Context, tenantID uuid.UUID, limit int) ([]ApprovalItem, error) {
	rows, err := s.repo.ListHistory(ctx, tenantID, limit)
	if err != nil {
		return nil, err
	}
	items := make([]ApprovalItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, approvalItemFromRecord(row))
	}
	return items, nil
}

// Summary cuenta pendientes y los que expiran en menos de 15 minutos.
func (s *Service) Summary(ctx context.Context, tenantID uuid.UUID) (ApprovalsSummary, error) {
	items, err := s.listPending(ctx, tenantID, 200)
	if err != nil {
		return ApprovalsSummary{}, err
	}
	out := ApprovalsSummary{PendingCount: len(items)}
	now := time.Now().UTC()
	for _, item := range items {
		if item.ExpiresAt == "" {
			continue
		}
		expires, perr := time.Parse(time.RFC3339, item.ExpiresAt)
		if perr != nil {
			continue
		}
		if expires.After(now) && expires.Sub(now) < expiringSoonWindow {
			out.ExpiringSoonCount++
		}
	}
	return out, nil
}

// GetApproval devuelve el detalle de una request gobernada, preferentemente
// desde Nexus, complementado con el espejo local.
func (s *Service) GetApproval(ctx context.Context, tenantID uuid.UUID, requestID string) (ApprovalItem, error) {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return ApprovalItem{}, domainerr.Validation("request_id is required")
	}
	if s.nexus != nil {
		req, st, err := s.nexus.Get(ctx, requestID, nexusclient.WithTenantID(tenantID.String()))
		if err == nil && st == http.StatusOK && !ownedByTenant(req.OrgID, tenantID) {
			// Request de otra org (o sin org): not found para no revelar existencia.
			return ApprovalItem{}, domainerr.NotFound("governance request not found")
		}
		if err == nil && st == http.StatusOK {
			item := approvalItemFromRequest(req)
			if item.Status == nexusclient.StatusPendingApproval {
				if approvals, aerr := s.nexus.ListPendingApprovals(ctx, "request_id="+url.QueryEscape(requestID), nexusclient.WithTenantID(tenantID.String())); aerr == nil {
					for _, a := range approvals {
						if a.RequestID == requestID && ownedByTenant(a.OrgID, tenantID) {
							applyApproval(&item, a)
							break
						}
					}
				}
			}
			if row, rerr := s.repo.GetByNexusRequestID(ctx, tenantID, requestID); rerr == nil {
				if item.ApprovalID == "" {
					item.ApprovalID = row.ApprovalID
				}
				if len(item.Decisions) == 0 {
					item.Decisions = decisionsFromRecord(row)
				}
			}
			return item, nil
		}
		// 404 o Nexus caído: fallback al espejo local antes de fallar.
	}
	row, err := s.repo.GetByNexusRequestID(ctx, tenantID, requestID)
	if err != nil {
		return ApprovalItem{}, err
	}
	return approvalItemFromRecord(row), nil
}

// Evidence devuelve el evidence pack crudo de Nexus, cacheado en ai_evidence_packs.
func (s *Service) Evidence(ctx context.Context, tenantID uuid.UUID, requestID string) (json.RawMessage, error) {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return nil, domainerr.Validation("request_id is required")
	}
	if cached, err := s.repo.GetEvidence(ctx, tenantID, requestID); err == nil {
		return cached.PackJSON, nil
	} else if !domainerr.IsNotFound(err) {
		return nil, err
	}
	if s.nexus == nil {
		return nil, domainerr.Unavailable("nexus client not configured")
	}
	raw, st, err := s.nexus.GetEvidence(ctx, requestID, nexusclient.WithTenantID(tenantID.String()))
	if err != nil {
		return nil, domainerr.UpstreamError("nexus get evidence failed")
	}
	if st == http.StatusNotFound {
		return nil, domainerr.NotFound("evidence pack not found")
	}
	if st != http.StatusOK {
		return nil, domainerr.UpstreamError(fmt.Sprintf("nexus get evidence status %d", st))
	}
	if !ownedByTenant(extractPackOrgID(raw), tenantID) {
		// Pack de otra org (o sin org): not found para no revelar existencia,
		// y nunca cachearlo bajo este tenant.
		return nil, domainerr.NotFound("evidence pack not found")
	}
	record := EvidenceRecord{TenantID: tenantID, NexusRequestID: requestID, PackJSON: raw}
	record.Signature, record.SignatureKeyID = extractPackSignature(raw)
	if serr := s.repo.SaveEvidence(ctx, record); serr != nil {
		// Cache best-effort: servir el pack aunque falle la persistencia.
		_ = serr
	}
	return raw, nil
}

// Decide aprueba/rechaza la approval asociada al request en nombre del usuario
// autenticado, propagando 409/403/404 de Nexus con su mensaje.
func (s *Service) Decide(ctx context.Context, tenantID uuid.UUID, requestID, actor, action, note string) error {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return domainerr.Validation("request_id is required")
	}
	if s.nexus == nil {
		return domainerr.Unavailable("nexus client not configured")
	}
	approvalID := ""
	row, rowErr := s.repo.GetByNexusRequestID(ctx, tenantID, requestID)
	if rowErr == nil {
		approvalID = row.ApprovalID
	}
	if approvalID == "" {
		// Lookup filtrado por request_id: el listado pleno de Nexus está capeado
		// y perdería approvals fuera del cap. El filtro client-side se mantiene
		// como fallback si la versión de Nexus ignora los query params.
		approvals, err := s.nexus.ListPendingApprovals(ctx, "request_id="+url.QueryEscape(requestID), nexusclient.WithTenantID(tenantID.String()))
		if err != nil {
			return domainerr.UpstreamError("nexus list pending approvals failed")
		}
		for _, a := range approvals {
			if a.RequestID == requestID && ownedByTenant(a.OrgID, tenantID) {
				approvalID = a.ID
				break
			}
		}
	}
	if approvalID == "" {
		return domainerr.NotFound("pending approval not found for request")
	}

	decidedBy := decidedByActor(actor)
	var st int
	var raw []byte
	var err error
	switch action {
	case "approve":
		st, raw, err = s.nexus.Approve(ctx, approvalID, decidedBy, note, nexusclient.WithTenantID(tenantID.String()))
	case "reject":
		st, raw, err = s.nexus.Reject(ctx, approvalID, decidedBy, note, nexusclient.WithTenantID(tenantID.String()))
	default:
		return domainerr.Validation("action must be approve or reject")
	}
	if err != nil {
		return domainerr.UpstreamError("nexus approval call failed")
	}
	switch st {
	case http.StatusOK:
		if rowErr == nil {
			// Persistir solo approval_id: el status terminal lo escribe únicamente
			// el callback approval_resolved (single writer). Nexus responde 200 en
			// aprobaciones parciales multi-approver/break-glass con la request aún
			// pending_approval, y sintetizar approved acá rompería el display y el
			// dedup de applyResolved (saltearía el executor en el callback genuino).
			_ = s.repo.UpdateByNexusRequestID(ctx, tenantID, requestID, map[string]any{
				"approval_id": approvalID,
			})
		}
		return nil
	case http.StatusConflict:
		return domainerr.Conflict(nexusclient.ParseErrorBody(raw))
	case http.StatusForbidden:
		return domainerr.Forbidden(nexusclient.ParseErrorBody(raw))
	case http.StatusNotFound:
		return domainerr.NotFound(nexusclient.ParseErrorBody(raw))
	default:
		return domainerr.UpstreamError(fmt.Sprintf("nexus approval status %d", st))
	}
}

// --- Helpers ---

func decidedByActor(actor string) string {
	actor = strings.TrimSpace(actor)
	if actor == "" || strings.Contains(actor, ":") {
		return actor
	}
	return "user:" + actor
}

func approvalItemFromRequest(req nexusclient.Request) ApprovalItem {
	return ApprovalItem{
		RequestID:   req.ID,
		ActionType:  req.ActionType,
		Status:      req.Status,
		RiskLevel:   req.RiskLevel,
		RequestedBy: req.RequesterID,
		Reason:      req.Reason,
		CreatedAt:   req.CreatedAt,
		Params:      req.Params,
		Decisions:   []nexusclient.ApprovalDecision{},
	}
}

func applyApproval(item *ApprovalItem, a nexusclient.Approval) {
	item.ApprovalID = a.ID
	item.ExpiresAt = a.ExpiresAt
	item.RequiredApprovals = a.RequiredApprovals
	item.CurrentApprovals = a.CurrentApprovals
	if len(a.Decisions) > 0 {
		item.Decisions = a.Decisions
	}
}

func approvalItemFromRecord(row RequestRecord) ApprovalItem {
	item := ApprovalItem{
		RequestID:   row.NexusRequestID,
		ApprovalID:  row.ApprovalID,
		ActionType:  row.ActionType,
		Status:      row.Status,
		RiskLevel:   row.RiskLevel,
		RequestedBy: row.RequesterID,
		CreatedAt:   row.CreatedAt.UTC().Format(time.RFC3339),
		Params:      unmarshalMap(row.ParamsJSON),
		Decisions:   decisionsFromRecord(row),
	}
	if payload := unmarshalMap(row.PayloadJSON); payload != nil {
		if reason, ok := payload["reason"].(string); ok {
			item.Reason = reason
		}
	}
	return item
}

// decisionsFromRecord sintetiza la decisión registrada localmente (el detalle
// multi-approver vive en Nexus; history se sirve con lo que haya local).
func decisionsFromRecord(row RequestRecord) []nexusclient.ApprovalDecision {
	if row.DecidedBy == "" {
		return []nexusclient.ApprovalDecision{}
	}
	action := "approve"
	if row.Status == nexusclient.StatusRejected {
		action = "reject"
	}
	return []nexusclient.ApprovalDecision{{
		ApproverID: row.DecidedBy,
		Action:     action,
		DecidedAt:  row.UpdatedAt.UTC().Format(time.RFC3339),
	}}
}

// extractPackOrgID lee request.org_id del evidence pack para validar que el
// pack pertenece al tenant antes de servirlo o cachearlo.
func extractPackOrgID(raw []byte) string {
	var pack struct {
		Request struct {
			OrgID string `json:"org_id"`
		} `json:"request"`
	}
	if json.Unmarshal(raw, &pack) != nil {
		return ""
	}
	return pack.Request.OrgID
}

func extractPackSignature(raw []byte) (signature, keyID string) {
	var pack struct {
		Signature      string `json:"signature"`
		SignatureKeyID string `json:"signature_key_id"`
		KeyID          string `json:"key_id"`
	}
	if json.Unmarshal(raw, &pack) != nil {
		return "", ""
	}
	keyID = pack.SignatureKeyID
	if keyID == "" {
		keyID = pack.KeyID
	}
	return pack.Signature, keyID
}

func marshalJSON(v any) []byte {
	if v == nil {
		return nil
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return raw
}

func unmarshalMap(raw []byte) map[string]any {
	if len(raw) == 0 {
		return nil
	}
	out := map[string]any{}
	if json.Unmarshal(raw, &out) != nil {
		return nil
	}
	return out
}
