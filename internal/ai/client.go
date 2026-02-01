package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client maneja llamadas al AI Copilot Service.
type Client struct {
	baseURL    string
	serviceKey string
	httpClient *http.Client
}

// NewClient crea un cliente de AI Copilot Service.
func NewClient(baseURL, serviceKey string, timeoutMS int) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		serviceKey: serviceKey,
		httpClient: &http.Client{Timeout: time.Duration(timeoutMS) * time.Millisecond},
	}
}

func (c *Client) Do(
	ctx context.Context,
	method string,
	path string,
	body any,
	userID string,
	projectID string,
) (int, []byte, error) {
	if strings.TrimSpace(c.baseURL) == "" {
		return 0, nil, fmt.Errorf("ai service url not configured")
	}
	if strings.TrimSpace(c.serviceKey) == "" {
		return 0, nil, fmt.Errorf("ai service key not configured")
	}
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return 0, nil, fmt.Errorf("failed to marshal body: %w", err)
		}
		reader = bytes.NewBuffer(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-SERVICE-KEY", c.serviceKey)
	req.Header.Set("X-USER-ID", userID)
	req.Header.Set("X-PROJECT-ID", projectID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("failed to read response: %w", err)
	}

	return resp.StatusCode, raw, nil
}
