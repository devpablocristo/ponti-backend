package axis

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var ErrNotConfigured = errors.New("axis companion not configured")

type Config struct {
	BaseURL        string
	APIKey         string
	ProductSurface string
	TimeoutMS      int
}

type CallContext struct {
	OrgID          string
	ActorID        string
	OnBehalfOf     string
	ProductSurface string
	Scopes         []string
}

type Client struct {
	baseURL        string
	apiKey         string
	productSurface string
	http           *http.Client
}

func NewClient(cfg Config) *Client {
	timeoutMS := cfg.TimeoutMS
	if timeoutMS <= 0 {
		timeoutMS = 10000
	}
	productSurface := strings.TrimSpace(cfg.ProductSurface)
	if productSurface == "" {
		productSurface = "ponti"
	}
	return &Client{
		baseURL:        strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/"),
		apiKey:         strings.TrimSpace(cfg.APIKey),
		productSurface: productSurface,
		http:           &http.Client{Timeout: time.Duration(timeoutMS) * time.Millisecond},
	}
}

func (c *Client) ProductSurface() string {
	if c == nil || c.productSurface == "" {
		return "ponti"
	}
	return c.productSurface
}

func (c *Client) DoJSON(ctx context.Context, call CallContext, method, path string, body any) (int, []byte, error) {
	if c == nil || c.baseURL == "" || c.apiKey == "" {
		return 0, nil, ErrNotConfigured
	}
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return 0, nil, err
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+normalizePath(path), reader)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)
	if strings.TrimSpace(call.OrgID) != "" {
		req.Header.Set("X-Org-ID", strings.TrimSpace(call.OrgID))
	}
	if strings.TrimSpace(call.ActorID) != "" {
		req.Header.Set("X-User-ID", strings.TrimSpace(call.ActorID))
	}
	onBehalfOf := strings.TrimSpace(call.OnBehalfOf)
	if onBehalfOf == "" {
		onBehalfOf = strings.TrimSpace(call.ActorID)
	}
	if onBehalfOf != "" {
		req.Header.Set("X-On-Behalf-Of", onBehalfOf)
	}
	productSurface := strings.TrimSpace(call.ProductSurface)
	if productSurface == "" {
		productSurface = c.ProductSurface()
	}
	req.Header.Set("X-Product-Surface", productSurface)
	if scopes := joinScopes(call.Scopes); scopes != "" {
		req.Header.Set("X-Auth-Scopes", scopes)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("axis companion request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return resp.StatusCode, nil, err
	}
	return resp.StatusCode, raw, nil
}

func normalizePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return "/"
	}
	if strings.HasPrefix(path, "/") {
		return path
	}
	return "/" + path
}

func joinScopes(scopes []string) string {
	if len(scopes) == 0 {
		return ""
	}
	seen := make(map[string]struct{}, len(scopes))
	out := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		scope = strings.TrimSpace(scope)
		if scope == "" {
			continue
		}
		if _, ok := seen[scope]; ok {
			continue
		}
		seen[scope] = struct{}{}
		out = append(out, scope)
	}
	return strings.Join(out, " ")
}
