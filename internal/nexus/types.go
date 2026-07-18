package nexus

import (
	"encoding/json"

	"github.com/devpablocristo/platform/kernels/governance/go/governanceclient"
)

// Decision reusa el set canónico del kernel de governance para que los
// consumers no hardcodeen strings.
const (
	DecisionAllow           = governanceclient.DecisionAllow
	DecisionDeny            = governanceclient.DecisionDeny
	DecisionRequireApproval = governanceclient.DecisionRequireApproval
)

// Status reusa el set canónico del kernel (alineado a Nexus KnownStatuses).
const (
	StatusPending         = governanceclient.StatusPending
	StatusEvaluated       = governanceclient.StatusEvaluated
	StatusAllowed         = governanceclient.StatusAllowed
	StatusDenied          = governanceclient.StatusDenied
	StatusPendingApproval = governanceclient.StatusPendingApproval
	StatusApproved        = governanceclient.StatusApproved
	StatusRejected        = governanceclient.StatusRejected
	StatusExpired         = governanceclient.StatusExpired
	StatusExecuted        = governanceclient.StatusExecuted
	StatusFailed          = governanceclient.StatusFailed
	StatusCancelled       = governanceclient.StatusCancelled
)

// KnownStatuses lista todos los valores válidos del campo `status` de un
// request en Nexus (delegado al kernel para no duplicar el contrato).
var KnownStatuses = governanceclient.KnownStatuses

// ToolIntentSchemaVersion es la única versión de action binding soportada.
const ToolIntentSchemaVersion = "tool_intent.v1"

// Action types canónicos que Ponti registra en Nexus para sus herramientas
// gobernadas (Ola B): un action type por tool habilita policies y verificación
// per-tool. ActionTypeCapabilityInvoke queda como legacy global de transición.
const (
	ActionTypeCapabilityInvoke     = "agent.capability.invoke"
	ActionTypeWorkOrderDraftCreate = "ponti.workorder.draft.create"
	ActionTypeInsightResolve       = "ponti.insight.resolve"
	ActionTypeStockAdjust          = "ponti.stock.adjust"
	ActionTypeStockCountApply      = "ponti.stock.count.apply"
)

// ToolIntent es el action binding tool_intent.v1 que Nexus exige para writes
// gobernados: ata la decisión a la invocación exacta de la herramienta.
type ToolIntent struct {
	SchemaVersion    string         `json:"schema_version"`
	OrgID            string         `json:"org_id"`
	ActorID          string         `json:"actor_id"`
	ActorType        string         `json:"actor_type"`
	ProductSurface   string         `json:"product_surface"`
	TaskID           string         `json:"task_id,omitempty"`
	RunID            string         `json:"run_id"`
	ToolInvocationID string         `json:"tool_invocation_id"`
	ConnectorID      string         `json:"connector_id"`
	CapabilityID     string         `json:"capability_id"`
	Operation        string         `json:"operation"`
	TargetSystem     string         `json:"target_system"`
	TargetResource   string         `json:"target_resource"`
	PayloadHash      string         `json:"payload_hash"`
	IdempotencyKey   string         `json:"idempotency_key"`
	RiskHint         string         `json:"risk_hint,omitempty"`
	EvidenceContext  map[string]any `json:"evidence_context,omitempty"`
}

// SubmitRequestBody es el cuerpo de POST /v1/requests con action binding.
type SubmitRequestBody struct {
	IdempotencyKey string         `json:"idempotency_key,omitempty"`
	RequesterType  string         `json:"requester_type"`
	RequesterID    string         `json:"requester_id"`
	RequesterName  string         `json:"requester_name,omitempty"`
	ActionType     string         `json:"action_type"`
	TargetSystem   string         `json:"target_system,omitempty"`
	TargetResource string         `json:"target_resource,omitempty"`
	ActionBinding  *ToolIntent    `json:"action_binding,omitempty"`
	Params         map[string]any `json:"params,omitempty"`
	Reason         string         `json:"reason,omitempty"`
	Context        string         `json:"context,omitempty"`
}

// ApprovalRef referencia la approval creada cuando la decisión es require_approval.
type ApprovalRef struct {
	ID        string `json:"id"`
	ExpiresAt string `json:"expires_at"`
}

// SubmitResponse es la respuesta de POST /v1/requests.
type SubmitResponse struct {
	RequestID      string       `json:"request_id"`
	Decision       string       `json:"decision"`
	RiskLevel      string       `json:"risk_level"`
	DecisionReason string       `json:"decision_reason"`
	Status         string       `json:"status"`
	BindingHash    string       `json:"binding_hash,omitempty"`
	Approval       *ApprovalRef `json:"approval,omitempty"`
	AISummary      string       `json:"ai_summary,omitempty"`
}

// Request es el shape completo de GET /v1/requests/{id} y de cada item de
// GET /v1/requests. ActionBinding queda como mapa para no fallar si Nexus
// agrega campos al binding persistido.
type Request struct {
	ID             string         `json:"id"`
	OrgID          string         `json:"org_id,omitempty"`
	RequesterType  string         `json:"requester_type"`
	RequesterID    string         `json:"requester_id"`
	RequesterName  string         `json:"requester_name,omitempty"`
	ActionType     string         `json:"action_type"`
	TargetSystem   string         `json:"target_system,omitempty"`
	TargetResource string         `json:"target_resource,omitempty"`
	ActionBinding  map[string]any `json:"action_binding,omitempty"`
	BindingHash    string         `json:"binding_hash,omitempty"`
	Params         map[string]any `json:"params,omitempty"`
	Reason         string         `json:"reason,omitempty"`
	RiskLevel      string         `json:"risk_level"`
	Decision       string         `json:"decision"`
	DecisionReason string         `json:"decision_reason,omitempty"`
	Status         string         `json:"status"`
	CreatedAt      string         `json:"created_at"`
	UpdatedAt      string         `json:"updated_at"`
}

// ReportResultBody es el cuerpo de POST /v1/requests/{id}/result.
type ReportResultBody struct {
	ResultID     string         `json:"result_id,omitempty"`
	Success      bool           `json:"success"`
	Result       map[string]any `json:"result,omitempty"`
	DurationMS   int64          `json:"duration_ms,omitempty"`
	ErrorMessage string         `json:"error_message,omitempty"`
}

// ApprovalDecision es una decisión individual (multi-approver / break-glass).
type ApprovalDecision struct {
	ApproverID string `json:"approver_id"`
	Action     string `json:"action"`
	Note       string `json:"note,omitempty"`
	DecidedAt  string `json:"decided_at"`
}

// Approval es cada item de GET /v1/approvals/pending.
type Approval struct {
	ID                string             `json:"id"`
	OrgID             string             `json:"org_id,omitempty"`
	RequestID         string             `json:"request_id"`
	Status            string             `json:"status"`
	DecidedBy         string             `json:"decided_by,omitempty"`
	DecisionNote      string             `json:"decision_note,omitempty"`
	DecidedAt         *string            `json:"decided_at,omitempty"`
	ExpiresAt         string             `json:"expires_at"`
	CreatedAt         string             `json:"created_at"`
	BreakGlass        bool               `json:"break_glass"`
	RequiredApprovals int                `json:"required_approvals"`
	CurrentApprovals  int                `json:"current_approvals"`
	Decisions         []ApprovalDecision `json:"decisions,omitempty"`
}

// AuditIntegrity es la respuesta de GET /v1/requests/{id}/replay/verify.
type AuditIntegrity struct {
	Status        string `json:"status"`
	CheckedEvents int    `json:"checked_events"`
	FirstHash     string `json:"first_hash,omitempty"`
	LastHash      string `json:"last_hash,omitempty"`
	Error         string `json:"error,omitempty"`
}

// ParseErrorBody intenta extraer el mensaje de error de una respuesta de Nexus.
func ParseErrorBody(raw []byte) string {
	var eb struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if json.Unmarshal(raw, &eb) == nil && eb.Message != "" {
		return eb.Message
	}
	return string(raw)
}
