// Package nexus implementa el cliente HTTP de Ponti hacia Nexus Governance
// con los DTOs extendidos de governance enterprise (action_binding ToolIntent,
// binding_hash, approvals on-behalf-of, attestations y evidence packs).
// Complementa al kernel governanceclient, que expone el subset genérico.
package nexus

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/devpablocristo/platform/http/go/httpclient"
)

// Client cliente HTTP hacia Nexus Governance.
type Client struct {
	caller *httpclient.Caller
}

// NewClient crea el cliente con baseURL, API key y timeout en milisegundos.
func NewClient(baseURL, apiKey string, timeoutMS int) *Client {
	if timeoutMS <= 0 {
		timeoutMS = 30000
	}
	h := make(http.Header)
	h.Set("X-API-Key", apiKey)
	return &Client{
		caller: &httpclient.Caller{
			BaseURL:     baseURL,
			Header:      h,
			HTTP:        &http.Client{Timeout: time.Duration(timeoutMS) * time.Millisecond},
			MaxBodySize: 4 << 20, // 4MB: evidence packs pueden ser grandes
		},
	}
}

// RequestOption modifica una llamada individual sin tocar la config global
// del Client (mismo patrón que governanceclient/options.go).
type RequestOption func(*requestOptions)

type requestOptions struct {
	tenantID       string
	idempotencyKey string
	onBehalfOf     string
}

// WithTenantID scopea la llamada a un tenant (header X-Org-ID). Requiere que
// la API key tenga scope nexus:cross_org.
func WithTenantID(tenantID string) RequestOption {
	return func(o *requestOptions) { o.tenantID = tenantID }
}

// WithIdempotencyKey setea el header Idempotency-Key en la llamada.
func WithIdempotencyKey(key string) RequestOption {
	return func(o *requestOptions) { o.idempotencyKey = key }
}

func collectOptions(opts ...RequestOption) []httpclient.RequestOption {
	var ro requestOptions
	for _, o := range opts {
		if o != nil {
			o(&ro)
		}
	}
	var out []httpclient.RequestOption
	if ro.tenantID != "" {
		out = append(out, httpclient.WithHeader("X-Org-ID", ro.tenantID))
	}
	if ro.idempotencyKey != "" {
		out = append(out, httpclient.WithIdempotencyKey(ro.idempotencyKey))
	}
	if ro.onBehalfOf != "" {
		out = append(out, httpclient.WithHeader("X-On-Behalf-Of", ro.onBehalfOf))
	}
	return out
}

// --- Requests ---

// Submit envía POST /v1/requests con action binding opcional.
func (c *Client) Submit(ctx context.Context, body SubmitRequestBody, opts ...RequestOption) (SubmitResponse, error) {
	var out SubmitResponse
	st, raw, err := c.caller.DoJSON(ctx, http.MethodPost, "/v1/requests", body, collectOptions(opts...)...)
	if err != nil {
		return out, fmt.Errorf("nexus submit: %w", err)
	}
	if st != http.StatusCreated {
		return out, fmt.Errorf("nexus submit: status %d body %s", st, ParseErrorBody(raw))
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return out, fmt.Errorf("decode submit response: %w", err)
	}
	return out, nil
}

// Get consulta GET /v1/requests/{id}. Devuelve status HTTP para distinguir 404.
func (c *Client) Get(ctx context.Context, requestID string, opts ...RequestOption) (Request, int, error) {
	var out Request
	st, raw, err := c.caller.DoJSON(ctx, http.MethodGet, "/v1/requests/"+url.PathEscape(requestID), nil, collectOptions(opts...)...)
	if err != nil {
		return out, 0, fmt.Errorf("nexus get request: %w", err)
	}
	if st == http.StatusNotFound {
		return out, st, nil
	}
	if st != http.StatusOK {
		return out, st, fmt.Errorf("nexus get request: status %d body %s", st, ParseErrorBody(raw))
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return out, st, fmt.Errorf("decode get response: %w", err)
	}
	return out, st, nil
}

// ListRequests trae GET /v1/requests?query (ej: "status=pending_approval&limit=50")
// y decodifica los items del envelope {"data":[...]}.
func (c *Client) ListRequests(ctx context.Context, query string, opts ...RequestOption) ([]Request, error) {
	path := "/v1/requests"
	if query != "" {
		path += "?" + query
	}
	st, raw, err := c.caller.DoJSON(ctx, http.MethodGet, path, nil, collectOptions(opts...)...)
	if err != nil {
		return nil, fmt.Errorf("nexus list requests: %w", err)
	}
	if st != http.StatusOK {
		return nil, fmt.Errorf("nexus list requests: status %d body %s", st, ParseErrorBody(raw))
	}
	var out struct {
		Data []Request `json:"data"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decode list response: %w", err)
	}
	return out.Data, nil
}

// ReportResult reporta resultado de ejecución POST /v1/requests/{id}/result.
// Devuelve el status HTTP para que el caller distinga 409 (no ejecutable).
func (c *Client) ReportResult(ctx context.Context, requestID string, body ReportResultBody, opts ...RequestOption) (int, error) {
	st, raw, err := c.caller.DoJSON(ctx, http.MethodPost, "/v1/requests/"+url.PathEscape(requestID)+"/result", body, collectOptions(opts...)...)
	if err != nil {
		return 0, fmt.Errorf("nexus report result: %w", err)
	}
	if st != http.StatusNoContent && st != http.StatusOK && st != http.StatusConflict {
		return st, fmt.Errorf("nexus report result: status %d body %s", st, ParseErrorBody(raw))
	}
	return st, nil
}

// Attest registra una attestation firmada POST /v1/requests/{id}/attest.
func (c *Client) Attest(ctx context.Context, requestID string, body any, opts ...RequestOption) (int, []byte, error) {
	st, raw, err := c.caller.DoJSON(ctx, http.MethodPost, "/v1/requests/"+url.PathEscape(requestID)+"/attest", body, collectOptions(opts...)...)
	if err != nil {
		return 0, nil, fmt.Errorf("nexus attest: %w", err)
	}
	return st, raw, nil
}

// GetEvidence trae el evidence pack firmado GET /v1/requests/{id}/evidence.
// Devuelve el JSON crudo sin decodificar para preservar la firma.
func (c *Client) GetEvidence(ctx context.Context, requestID string, opts ...RequestOption) ([]byte, int, error) {
	st, raw, err := c.caller.DoJSON(ctx, http.MethodGet, "/v1/requests/"+url.PathEscape(requestID)+"/evidence", nil, collectOptions(opts...)...)
	if err != nil {
		return nil, 0, fmt.Errorf("nexus get evidence: %w", err)
	}
	return raw, st, nil
}

// GetReplayVerify verifica integridad tamper-evident GET /v1/requests/{id}/replay/verify.
func (c *Client) GetReplayVerify(ctx context.Context, requestID string, opts ...RequestOption) (AuditIntegrity, int, error) {
	var out AuditIntegrity
	st, raw, err := c.caller.DoJSON(ctx, http.MethodGet, "/v1/requests/"+url.PathEscape(requestID)+"/replay/verify", nil, collectOptions(opts...)...)
	if err != nil {
		return out, 0, fmt.Errorf("nexus replay verify: %w", err)
	}
	if st != http.StatusOK {
		return out, st, fmt.Errorf("nexus replay verify: status %d body %s", st, ParseErrorBody(raw))
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return out, st, fmt.Errorf("decode replay verify response: %w", err)
	}
	return out, st, nil
}

// --- Approvals ---

// ListPendingApprovals trae GET /v1/approvals/pending?query (org-scoped via
// WithTenantID) y decodifica los items del envelope {"data":[...]}. query es
// opcional (ej: "request_id=req-1&limit=50"); versiones de Nexus que ignoren
// los params devuelven la lista completa, por lo que el caller debe filtrar
// client-side como fallback.
func (c *Client) ListPendingApprovals(ctx context.Context, query string, opts ...RequestOption) ([]Approval, error) {
	path := "/v1/approvals/pending"
	if query != "" {
		path += "?" + query
	}
	st, raw, err := c.caller.DoJSON(ctx, http.MethodGet, path, nil, collectOptions(opts...)...)
	if err != nil {
		return nil, fmt.Errorf("nexus list pending approvals: %w", err)
	}
	if st != http.StatusOK {
		return nil, fmt.Errorf("nexus list pending approvals: status %d body %s", st, ParseErrorBody(raw))
	}
	var out struct {
		Data []Approval `json:"data"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decode pending approvals response: %w", err)
	}
	return out.Data, nil
}

// Approve aprueba una approval POST /v1/approvals/{id}/approve en nombre del
// usuario real: manda header X-On-Behalf-Of y body decided_by (Nexus prefiere
// el principal autenticado; el header habilita la delegación de service keys).
// Devuelve status y body crudos para propagar 409/403/404 con el mensaje de Nexus.
func (c *Client) Approve(ctx context.Context, approvalID, decidedBy, note string, opts ...RequestOption) (int, []byte, error) {
	return c.decide(ctx, approvalID, "approve", decidedBy, note, opts...)
}

// Reject rechaza una approval POST /v1/approvals/{id}/reject. Misma semántica
// on-behalf-of que Approve.
func (c *Client) Reject(ctx context.Context, approvalID, decidedBy, note string, opts ...RequestOption) (int, []byte, error) {
	return c.decide(ctx, approvalID, "reject", decidedBy, note, opts...)
}

func (c *Client) decide(ctx context.Context, approvalID, action, decidedBy, note string, opts ...RequestOption) (int, []byte, error) {
	body := map[string]any{"decided_by": decidedBy}
	if note != "" {
		body["note"] = note
	}
	allOpts := opts
	if decidedBy != "" {
		allOpts = append(allOpts, func(o *requestOptions) { o.onBehalfOf = decidedBy })
	}
	st, raw, err := c.caller.DoJSON(ctx, http.MethodPost, "/v1/approvals/"+url.PathEscape(approvalID)+"/"+action, body, collectOptions(allOpts...)...)
	if err != nil {
		return 0, nil, fmt.Errorf("nexus %s approval: %w", action, err)
	}
	return st, raw, nil
}
