package axis

import "time"

// NexusSubmitRequest matchea `SubmitRequest` del OpenAPI de axis/nexus.
// El caller arma una "intención de acción" que Nexus evalúa contra policies
// y devuelve `allow`/`deny`/`require_approval` + `binding_hash`.
type NexusSubmitRequest struct {
	IdempotencyKey string         `json:"idempotency_key,omitempty"`
	RequesterType  string         `json:"requester_type"`
	RequesterID    string         `json:"requester_id"`
	RequesterName  string         `json:"requester_name,omitempty"`
	ActionType     string         `json:"action_type"`
	TargetSystem   string         `json:"target_system,omitempty"`
	TargetResource string         `json:"target_resource,omitempty"`
	ActionBinding  map[string]any `json:"action_binding,omitempty"`
	Params         map[string]any `json:"params,omitempty"`
	Reason         string         `json:"reason,omitempty"`
	Context        string         `json:"context,omitempty"`
}

// NexusDecision son los valores posibles de `SubmitResponse.decision`.
type NexusDecision string

const (
	DecisionAllow           NexusDecision = "allow"
	DecisionDeny            NexusDecision = "deny"
	DecisionRequireApproval NexusDecision = "require_approval"
)

// NexusSubmitResponse matchea `SubmitResponse` del OpenAPI.
type NexusSubmitResponse struct {
	RequestID      string         `json:"request_id"`
	Decision       NexusDecision  `json:"decision"`
	RiskLevel      string         `json:"risk_level"`
	DecisionReason string         `json:"decision_reason,omitempty"`
	Status         string         `json:"status"`
	BindingHash    string         `json:"binding_hash,omitempty"`
	Approval       *NexusApproval `json:"approval,omitempty"`
}

// NexusApproval es el handle de aprobación cuando `decision=require_approval`.
type NexusApproval struct {
	ID        string    `json:"id"`
	ExpiresAt time.Time `json:"expires_at"`
}

// NexusReportResult es el body de `POST /v1/requests/{id}/result` que el caller
// envía después de ejecutar la acción aprobada.
type NexusReportResult struct {
	ResultID     string         `json:"result_id,omitempty"`
	Success      bool           `json:"success"`
	Result       map[string]any `json:"result,omitempty"`
	DurationMS   int            `json:"duration_ms,omitempty"`
	ErrorMessage string         `json:"error_message,omitempty"`
}
