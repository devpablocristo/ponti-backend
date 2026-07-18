package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
)

const capabilityExecutionSchemaVersion = "capability_execution.v1"

type capabilityExecutionRequest struct {
	SchemaVersion  string                   `json:"schema_version"`
	Operation      string                   `json:"operation"`
	ExecutorRef    string                   `json:"executor_ref,omitempty"`
	Payload        json.RawMessage          `json:"payload"`
	Workspace      map[string]any           `json:"workspace,omitempty"`
	IdempotencyKey string                   `json:"idempotency_key,omitempty"`
	TaskID         string                   `json:"task_id,omitempty"`
	RunID          string                   `json:"run_id,omitempty"`
	NexusRequestID string                   `json:"nexus_request_id,omitempty"`
	Actor          capabilityExecutionActor `json:"actor"`
	OrgID          string                   `json:"org_id"`
}

type capabilityExecutionActor struct {
	ActorID        string `json:"actor_id"`
	ActorType      string `json:"actor_type,omitempty"`
	OnBehalfOf     string `json:"on_behalf_of,omitempty"`
	ProductSurface string `json:"product_surface,omitempty"`
}

type capabilityExecutionResponse struct {
	Status      string          `json:"status"`
	ExternalRef string          `json:"external_ref,omitempty"`
	Result      json.RawMessage `json:"result"`
	Evidence    map[string]any  `json:"evidence"`
	Error       string          `json:"error,omitempty"`
}

type capabilityExecutionTarget struct {
	Method string
	Path   string
	Body   json.RawMessage
}

// CapabilityExecution implementa el contrato server-to-server de Axis:
// POST /api/v1/capability-executions con envelope capability_execution.v1.
// La ejecución queda dentro de Ponti: el dispatcher llama al router interno
// para reutilizar handlers, validaciones, tenancy y gates de Nexus existentes.
func (h *Handler) CapabilityExecution(c *gin.Context) {
	var req capabilityExecutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sharedhandlers.RespondError(c, domainerr.Validation("invalid request payload"))
		return
	}
	req.SchemaVersion = strings.TrimSpace(req.SchemaVersion)
	req.Operation = strings.TrimSpace(req.Operation)
	if req.SchemaVersion != capabilityExecutionSchemaVersion {
		sharedhandlers.RespondError(c, domainerr.Validation("schema_version must be capability_execution.v1"))
		return
	}
	if req.Operation == "" {
		sharedhandlers.RespondError(c, domainerr.Validation("operation is required"))
		return
	}
	if len(req.Payload) == 0 || string(req.Payload) == "null" {
		req.Payload = json.RawMessage(`{}`)
	}

	orgID, actor, err := requireActionContext(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	target, workspace, err := h.capabilityExecutionTarget(req)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}

	rec := h.dispatchCapabilityExecution(c, target, req)
	body := rec.Body.Bytes()
	result := json.RawMessage(`{}`)
	if len(strings.TrimSpace(string(body))) > 0 {
		result = append(json.RawMessage(nil), body...)
		if !json.Valid(result) {
			encoded, marshalErr := json.Marshal(map[string]any{"raw": string(body)})
			if marshalErr == nil {
				result = encoded
			}
		}
	}

	status := "success"
	errMsg := ""
	if rec.Code < http.StatusOK || rec.Code >= http.StatusMultipleChoices {
		status = "failure"
		errMsg = strings.TrimSpace(string(body))
	}

	resp := capabilityExecutionResponse{
		Status:      status,
		ExternalRef: capabilityExecutionExternalRef(req),
		Result:      result,
		Evidence: map[string]any{
			"source_ref":        req.Operation,
			"captured_at":       time.Now().UTC().Format(time.RFC3339),
			"tenant_scope":      orgID.String(),
			"actor_id":          actor,
			"workspace":         workspace,
			"executor_ref":      req.ExecutorRef,
			"internal_route":    target.Method + " " + target.Path,
			"http_status":       rec.Code,
			"schema_version":    capabilityExecutionSchemaVersion,
			"product_surface":   pontiProductSurface,
			"axis_org_id":       strings.TrimSpace(req.OrgID),
			"axis_run_id":       strings.TrimSpace(req.RunID),
			"axis_task_id":      strings.TrimSpace(req.TaskID),
			"idempotency_key":   strings.TrimSpace(req.IdempotencyKey),
			"nexus_request_id":  firstNonEmpty(strings.TrimSpace(req.NexusRequestID), strings.TrimSpace(c.GetHeader(nexusRequestIDHeader))),
			"on_behalf_of":      strings.TrimSpace(req.Actor.OnBehalfOf),
			"axis_actor_id":     strings.TrimSpace(req.Actor.ActorID),
			"axis_actor_type":   strings.TrimSpace(req.Actor.ActorType),
			"axis_product_hint": strings.TrimSpace(req.Actor.ProductSurface),
		},
		Error: errMsg,
	}
	c.JSON(capabilityExecutionHTTPStatus(rec.Code), resp)
}

func (h *Handler) dispatchCapabilityExecution(c *gin.Context, target capabilityExecutionTarget, req capabilityExecutionRequest) *httptest.ResponseRecorder {
	var body *bytes.Reader
	if target.Method == http.MethodGet {
		body = bytes.NewReader(nil)
	} else {
		if len(target.Body) == 0 {
			target.Body = json.RawMessage(`{}`)
		}
		body = bytes.NewReader(target.Body)
	}
	internalReq := httptest.NewRequestWithContext(c.Request.Context(), target.Method, target.Path, body)
	for _, header := range []string{
		"Authorization",
		"X-Ponti-Axis-Api-Key",
		"X-Tenant-Id",
		"X-API-KEY",
		"X-USER-ID",
		"X-PROJECT-ID",
		nexusRequestIDHeader,
	} {
		if value := strings.TrimSpace(c.GetHeader(header)); value != "" {
			internalReq.Header.Set(header, value)
		}
	}
	if req.NexusRequestID != "" && internalReq.Header.Get(nexusRequestIDHeader) == "" {
		internalReq.Header.Set(nexusRequestIDHeader, req.NexusRequestID)
	}
	internalReq.Header.Set("Accept", "application/json")
	if target.Method != http.MethodGet {
		internalReq.Header.Set("Content-Type", "application/json")
	}

	rec := httptest.NewRecorder()
	h.gsv.GetRouter().ServeHTTP(rec, internalReq)
	return rec
}

func (h *Handler) capabilityExecutionTarget(req capabilityExecutionRequest) (capabilityExecutionTarget, map[string]any, error) {
	payload, err := executionPayloadObject(req.Payload)
	if err != nil {
		return capabilityExecutionTarget{}, nil, err
	}
	workspace := executionWorkspace(req, payload)
	base := h.acf.APIBaseURL()
	query := url.Values{}

	switch req.Operation {
	case "ponti.insights.list":
		addPayloadQuery(query, payload, "limit", "include_resolved")
		return getTarget(base+"/insights", query), workspace, nil
	case "ponti.insights.summary":
		return getTarget(base+"/insights/summary", query), workspace, nil
	case "ponti.insights.explain":
		insightID := stringValue(payload["insight_id"])
		if insightID == "" {
			return capabilityExecutionTarget{}, nil, domainerr.Validation("insight_id is required")
		}
		return getTarget(base+"/insights/"+url.PathEscape(insightID)+"/explain", query), workspace, nil
	case "ponti.dashboard.summary":
		addWorkspaceQuery(query, workspace)
		return getTarget(base+"/dashboard", query), workspace, nil
	case "ponti.stock.summary":
		projectID := firstNonEmpty(stringValue(payload["project_id"]), stringValue(workspace["project_id"]))
		if projectID == "" {
			return capabilityExecutionTarget{}, nil, domainerr.Validation("project_id is required")
		}
		addWorkspaceQuery(query, workspace)
		addPayloadQuery(query, payload, "cutoff_date")
		return getTarget(base+"/projects/"+url.PathEscape(projectID)+"/stocks/summary", query), workspace, nil
	case "ponti.workorders.list":
		addWorkspaceQuery(query, workspace)
		addPayloadQuery(query, payload, "status", "supply_id", "is_digital")
		addLimitAsPerPage(query, payload)
		return getTarget(base+"/work-orders", query), workspace, nil
	case "ponti.workorders.metrics":
		addWorkspaceQuery(query, workspace)
		addPayloadQuery(query, payload, "status", "supply_id", "is_digital")
		return getTarget(base+"/work-orders/metrics", query), workspace, nil
	case "ponti.lots.summary":
		addWorkspaceQuery(query, workspace)
		addPayloadQuery(query, payload, "crop_id")
		addLimitAsPerPage(query, payload)
		return getTarget(base+"/lots", query), workspace, nil
	case "ponti.supplies.summary":
		addWorkspaceQuery(query, workspace)
		addPayloadQuery(query, payload, "mode")
		addLimitAsPerPage(query, payload)
		return getTarget(base+"/supplies", query), workspace, nil
	case "ponti.reports.field_crop.summary":
		addWorkspaceQuery(query, workspace)
		return getTarget(base+"/reports/field-crop", query), workspace, nil
	case "ponti.reports.investor_contribution.summary":
		addWorkspaceQuery(query, workspace)
		return getTarget(base+"/reports/investor-contribution", query), workspace, nil
	case "ponti.reports.summary_results.summary":
		addWorkspaceQuery(query, workspace)
		return getTarget(base+"/reports/summary-results", query), workspace, nil
	case "ponti.data_integrity.summary":
		addWorkspaceQuery(query, workspace)
		return getTarget(base+"/data-integrity/summary", query), workspace, nil
	case "ponti.insight.resolve.prepare":
		return postTarget(base+"/ai/actions/insight-resolve/prepare", req.Payload), workspace, nil
	case "ponti.workorder.draft.prepare":
		return postTarget(base+"/ai/actions/workorder-draft/prepare", req.Payload), workspace, nil
	case "ponti.stock_adjustment.prepare":
		return postTarget(base+"/ai/actions/stock-adjustment/prepare", req.Payload), workspace, nil
	case "ponti.workorder_draft.create":
		return postTarget(base+"/work-order-drafts/digital", req.Payload), workspace, nil
	case "ponti.insight_resolution.draft":
		return postTarget(base+"/ai/actions/insight-resolution/draft", req.Payload), workspace, nil
	case "ponti.stock_count.draft":
		return postTarget(base+"/ai/actions/stock-count/draft", req.Payload), workspace, nil
	default:
		return capabilityExecutionTarget{}, nil, domainerr.Validation(fmt.Sprintf("unknown operation: %s", req.Operation))
	}
}

func getTarget(path string, query url.Values) capabilityExecutionTarget {
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	return capabilityExecutionTarget{Method: http.MethodGet, Path: path}
}

func postTarget(path string, body json.RawMessage) capabilityExecutionTarget {
	return capabilityExecutionTarget{Method: http.MethodPost, Path: path, Body: body}
}

func executionPayloadObject(raw json.RawMessage) (map[string]any, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return map[string]any{}, nil
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, domainerr.Validation("payload must be a JSON object")
	}
	if payload == nil {
		payload = map[string]any{}
	}
	return payload, nil
}

func executionWorkspace(req capabilityExecutionRequest, payload map[string]any) map[string]any {
	out := copyStringAnyMap(req.Workspace)
	if len(out) == 0 {
		if nested, ok := payload["workspace"].(map[string]any); ok {
			out = copyStringAnyMap(nested)
		}
	}
	for _, key := range []string{"customer_id", "project_id", "campaign_id", "field_id"} {
		if _, exists := out[key]; !exists {
			if value, ok := payload[key]; ok {
				out[key] = value
			}
		}
	}
	return out
}

func copyStringAnyMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func addWorkspaceQuery(query url.Values, workspace map[string]any) {
	for _, key := range []string{"customer_id", "project_id", "campaign_id", "field_id"} {
		if value := stringValue(workspace[key]); value != "" {
			query.Set(key, value)
		}
	}
}

func addPayloadQuery(query url.Values, payload map[string]any, keys ...string) {
	for _, key := range keys {
		if value := stringValue(payload[key]); value != "" {
			query.Set(key, value)
		}
	}
}

func addLimitAsPerPage(query url.Values, payload map[string]any) {
	if value := stringValue(payload["limit"]); value != "" {
		query.Set("per_page", value)
	}
}

func stringValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(typed)
	case bool:
		return strconv.FormatBool(typed)
	case float64:
		if typed == float64(int64(typed)) {
			return strconv.FormatInt(int64(typed), 10)
		}
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case int:
		return strconv.Itoa(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	case json.Number:
		return typed.String()
	default:
		return strings.TrimSpace(fmt.Sprint(typed))
	}
}

func capabilityExecutionExternalRef(req capabilityExecutionRequest) string {
	if req.IdempotencyKey != "" {
		return "ponti:" + req.Operation + ":" + req.IdempotencyKey
	}
	if req.RunID != "" {
		return "ponti:" + req.Operation + ":" + req.RunID
	}
	return "ponti:" + req.Operation
}

func capabilityExecutionHTTPStatus(internalStatus int) int {
	if internalStatus == 0 {
		return http.StatusInternalServerError
	}
	if internalStatus >= http.StatusBadRequest {
		return internalStatus
	}
	return http.StatusOK
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
