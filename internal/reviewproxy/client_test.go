package reviewproxy_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/devpablocristo/core/governance/go/reviewclient"
	"github.com/devpablocristo/ponti-backend/internal/reviewproxy"
)

func TestNewClient_SubmitRequestSendsAPIKey(t *testing.T) {
	var gotKey, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("X-API-Key")
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"request_id":"r-1","decision":"allow","risk_level":"low","status":"allowed"}`))
	}))
	t.Cleanup(srv.Close)

	client := reviewproxy.NewClient(srv.URL, "test-key")
	body := reviewclient.SubmitRequestBody{
		RequesterType: "service",
		RequesterID:   "ponti-backend",
		ActionType:    "stock.write",
		TargetSystem:  "ponti",
		Params:        map[string]any{"quantity": float64(-10)},
	}
	_, err := client.SubmitRequest(context.Background(), "idem-1", body)
	if err != nil {
		t.Fatalf("SubmitRequest: %v", err)
	}
	if gotKey != "test-key" {
		t.Fatalf("X-API-Key = %q, want %q", gotKey, "test-key")
	}
	if gotPath != "/v1/requests" {
		t.Fatalf("path = %q, want /v1/requests", gotPath)
	}
}
