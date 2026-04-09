package ai

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/devpablocristo/core/http/go/httpclient"
)

// Client maneja llamadas a Ponti AI (`InsightService` + `CopilotAgent`).
type Client struct {
	caller *httpclient.Caller
}

// NewClient crea un cliente de Ponti AI.
func NewClient(baseURL, serviceKey string, timeoutMS int) *Client {
	if timeoutMS <= 0 {
		timeoutMS = 10000
	}
	h := make(http.Header)
	h.Set("X-SERVICE-KEY", strings.TrimSpace(serviceKey))
	return &Client{
		caller: &httpclient.Caller{
			BaseURL:     strings.TrimRight(baseURL, "/"),
			Header:      h,
			HTTP:        &http.Client{Timeout: time.Duration(timeoutMS) * time.Millisecond},
			MaxBodySize: 1 << 20,
		},
	}
}

// Do ejecuta una petición al AI service con headers dinámicos por request.
func (c *Client) Do(
	ctx context.Context,
	method string,
	path string,
	body any,
	userID string,
	projectID string,
) (int, []byte, error) {
	if c.caller.BaseURL == "" {
		return 0, nil, fmt.Errorf("ai service url not configured")
	}
	return c.caller.DoJSON(ctx, method, path, body,
		httpclient.WithHeader("X-USER-ID", userID),
		httpclient.WithHeader("X-PROJECT-ID", projectID),
	)
}

// DoStream reenvía el body tal cual (p. ej. JSON del chat) y devuelve la respuesta sin tope de timeout (SSE).
func (c *Client) DoStream(
	ctx context.Context,
	method string,
	path string,
	body io.Reader,
	contentType string,
	userID string,
	projectID string,
) (*http.Response, error) {
	if c.caller.BaseURL == "" {
		return nil, fmt.Errorf("ai service url not configured")
	}
	u := strings.TrimSuffix(c.caller.BaseURL, "/")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	u += path
	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return nil, err
	}
	for k, vals := range c.caller.Header {
		for _, v := range vals {
			req.Header.Add(k, v)
		}
	}
	req.Header.Set("X-USER-ID", userID)
	req.Header.Set("X-PROJECT-ID", projectID)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	cli := &http.Client{Timeout: 0}
	return cli.Do(req)
}
