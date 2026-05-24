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

// Defaults para el wrapper. Mantienen el cliente operable sin tener que cablear
// cada knob en el config — solo `BaseURL` y `JWTSecret` son obligatorios.
const (
	defaultTimeout    = 30 * time.Second
	defaultMaxRetries = 2
	defaultRetryDelay = 250 * time.Millisecond
	defaultMaxBody    = 8 * 1024 * 1024 // 8 MiB
	defaultJWTTTL     = 5 * time.Minute
)

// Default scopes que firmamos en cada JWT. Companion los mira para decidir si
// permite operaciones de chat. Si más adelante hace falta dividir lectura/escritura
// por endpoint, mover este array a por-método.
var defaultChatScopes = []string{
	"companion:tasks:read",
	"companion:tasks:write",
}

// Config agrupa lo que el wrapper necesita. Se llena desde `cmd/config/companion.go`
// vía env vars (`COMPANION_BASE_URL`, `COMPANION_INTERNAL_JWT_SECRET`, etc.).
type Config struct {
	BaseURL      string
	JWTSecret    string
	JWTIssuer    string
	JWTAudience  string
	JWTTTL       time.Duration
	Timeout      time.Duration
	MaxRetries   int
}

// CompanionClient envuelve `httpclient.Caller` con la firma JWT requerida por
// Companion. Es seguro para uso concurrente (Caller y signer son inmutables
// post-construcción).
type CompanionClient struct {
	caller *httpclient.Caller
	signer *jwtSigner
}

// NewCompanionClient construye el cliente. Si `BaseURL` está vacío, retorna
// `ErrNotConfigured` — los usecases deben hacer fallback (no romper el handler).
func NewCompanionClient(cfg Config) (*CompanionClient, error) {
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

	signer, err := newJWTSigner(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTAudience, ttl)
	if err != nil {
		return nil, err
	}

	caller := &httpclient.Caller{
		HTTP:           newHTTPClient(timeout),
		BaseURL:        strings.TrimRight(cfg.BaseURL, "/"),
		MaxRetries:     maxRetries,
		RetryBaseDelay: defaultRetryDelay,
		MaxBodySize:    defaultMaxBody,
	}

	return &CompanionClient{caller: caller, signer: signer}, nil
}

// Chat invoca `POST /v1/chat` y devuelve el response parseado.
// Firma un JWT short-lived con el `CallContext` (org_id + actor + scopes).
func (c *CompanionClient) Chat(ctx context.Context, call CallContext, req ChatRequest) (*ChatResponse, error) {
	token, err := c.signer.sign(callWithDefaultScopes(call))
	if err != nil {
		return nil, fmt.Errorf("sign jwt: %w", err)
	}
	status, body, err := c.caller.DoJSON(ctx, "POST", "/v1/chat", req,
		httpclient.WithHeader("Authorization", "Bearer "+token),
	)
	if err != nil {
		return nil, err
	}
	if status != 200 {
		return nil, mapHTTPError(status, body)
	}
	var out ChatResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode chat response: %w", err)
	}
	return &out, nil
}

// ListConversations invoca `GET /v1/chat/conversations?limit=N`.
func (c *CompanionClient) ListConversations(ctx context.Context, call CallContext, limit int) (*ConversationList, error) {
	token, err := c.signer.sign(callWithDefaultScopes(call))
	if err != nil {
		return nil, fmt.Errorf("sign jwt: %w", err)
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	path := "/v1/chat/conversations?limit=" + fmt.Sprintf("%d", limit)
	status, body, err := c.caller.DoJSON(ctx, "GET", path, nil,
		httpclient.WithHeader("Authorization", "Bearer "+token),
	)
	if err != nil {
		return nil, err
	}
	if status != 200 {
		return nil, mapHTTPError(status, body)
	}
	var out ConversationList
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode conversations list: %w", err)
	}
	return &out, nil
}

// GetConversation invoca `GET /v1/chat/conversations/{id}`.
// `id` debe estar URL-encoded por el caller; aquí solo validamos no-vacío.
func (c *CompanionClient) GetConversation(ctx context.Context, call CallContext, id string) (*ConversationDetail, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("axis: conversation id required")
	}
	token, err := c.signer.sign(callWithDefaultScopes(call))
	if err != nil {
		return nil, fmt.Errorf("sign jwt: %w", err)
	}
	path := "/v1/chat/conversations/" + url.PathEscape(id)
	status, body, err := c.caller.DoJSON(ctx, "GET", path, nil,
		httpclient.WithHeader("Authorization", "Bearer "+token),
	)
	if err != nil {
		return nil, err
	}
	if status != 200 {
		return nil, mapHTTPError(status, body)
	}
	var out ConversationDetail
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode conversation detail: %w", err)
	}
	return &out, nil
}

// callWithDefaultScopes asegura que el JWT siempre lleve los scopes mínimos
// que Companion requiere para chat, incluso si el caller no los pasó.
func callWithDefaultScopes(call CallContext) CallContext {
	if len(call.Scopes) == 0 {
		call.Scopes = defaultChatScopes
	}
	return call
}
