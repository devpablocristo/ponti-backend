package axis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/devpablocristo/platform/http/go/httpclient"
)

// Scopes que Ponti envía al pedir decisión a Nexus. La key admin del docker-compose
// de axis tiene todos los scopes; en prod cada producto declara sólo los necesarios.
var defaultNexusScopes = []string{
	"nexus:requests:read",
	"nexus:requests:write",
	"nexus:requests:result",
}

// NexusConfig agrupa lo necesario para construir el cliente de Nexus.
// El JWT se firma con el mismo secret pattern que Companion (HS256 + claims).
type NexusConfig struct {
	BaseURL     string
	JWTSecret   string
	JWTIssuer   string
	JWTAudience string
	JWTTTL      time.Duration
	Timeout     time.Duration
	MaxRetries  int
}

// NexusClient envuelve `httpclient.Caller` con la firma JWT requerida por Nexus.
// API documentada en `axis/nexus/openapi.yaml`.
type NexusClient struct {
	caller *httpclient.Caller
	signer *jwtSigner
}

// NewNexusClient construye el cliente. Si `BaseURL` está vacío, retorna
// `ErrNotConfigured` — los callers pueden tratar Nexus como "no integrado"
// y saltarse el gating (válido para MVP solo-chat).
func NewNexusClient(cfg NexusConfig) (*NexusClient, error) {
	if strings.TrimSpace(cfg.BaseURL) == "" {
		return nil, ErrNotConfigured
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	maxRetries := cfg.MaxRetries
	if maxRetries < 0 {
		maxRetries = defaultMaxRetries
	}
	ttl := cfg.JWTTTL
	if ttl <= 0 {
		ttl = defaultJWTTTL
	}
	if cfg.JWTAudience == "" {
		cfg.JWTAudience = "nexus"
	}

	signer, err := newJWTSigner(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTAudience, ttl)
	if err != nil {
		return nil, err
	}
	return &NexusClient{
		caller: &httpclient.Caller{
			HTTP:           newHTTPClient(timeout),
			BaseURL:        strings.TrimRight(cfg.BaseURL, "/"),
			MaxRetries:     maxRetries,
			RetryBaseDelay: defaultRetryDelay,
			MaxBodySize:    defaultMaxBody,
		},
		signer: signer,
	}, nil
}

func (c *NexusClient) bearer(call CallContext) (string, error) {
	if len(call.Scopes) == 0 {
		call.Scopes = defaultNexusScopes
	}
	return c.signer.sign(call)
}

// SubmitRequest envía una acción para evaluación. Nexus responde con la decisión
// + binding_hash. El caller compara el binding_hash con el que él calculó local
// para evitar mismatches entre lo aprobado y lo que va a ejecutar.
func (c *NexusClient) SubmitRequest(ctx context.Context, call CallContext, req NexusSubmitRequest) (*NexusSubmitResponse, error) {
	token, err := c.bearer(call)
	if err != nil {
		return nil, fmt.Errorf("sign jwt: %w", err)
	}
	status, body, err := c.caller.DoJSON(ctx, "POST", "/v1/requests", req,
		httpclient.WithHeader("Authorization", "Bearer "+token),
	)
	if err != nil {
		return nil, err
	}
	if status != 200 && status != 201 {
		return nil, mapHTTPError(status, body)
	}
	var out NexusSubmitResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode nexus submit: %w", err)
	}
	return &out, nil
}

// GetRequest obtiene el estado actual de una request previa.
func (c *NexusClient) GetRequest(ctx context.Context, call CallContext, id string) (*NexusSubmitResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("axis: request id required")
	}
	token, err := c.bearer(call)
	if err != nil {
		return nil, fmt.Errorf("sign jwt: %w", err)
	}
	path := "/v1/requests/" + url.PathEscape(id)
	status, body, err := c.caller.DoJSON(ctx, "GET", path, nil,
		httpclient.WithHeader("Authorization", "Bearer "+token),
	)
	if err != nil {
		return nil, err
	}
	if status != 200 {
		return nil, mapHTTPError(status, body)
	}
	var out NexusSubmitResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode nexus get: %w", err)
	}
	return &out, nil
}

// ReportResult notifica a Nexus el resultado de la acción ejecutada.
// Sirve para auditoría + evidence pack.
func (c *NexusClient) ReportResult(ctx context.Context, call CallContext, id string, result NexusReportResult) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.New("axis: request id required")
	}
	token, err := c.bearer(call)
	if err != nil {
		return fmt.Errorf("sign jwt: %w", err)
	}
	path := "/v1/requests/" + url.PathEscape(id) + "/result"
	status, body, err := c.caller.DoJSON(ctx, "POST", path, result,
		httpclient.WithHeader("Authorization", "Bearer "+token),
	)
	if err != nil {
		return err
	}
	if status != 200 && status != 204 {
		return mapHTTPError(status, body)
	}
	return nil
}
