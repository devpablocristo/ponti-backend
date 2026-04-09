package ai

import (
	"context"
	"fmt"
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
