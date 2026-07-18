package axis

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_DoJSON_SendsDelegatedAxisHeaders(t *testing.T) {
	t.Parallel()
	var gotHeaders http.Header
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeaders = r.Header.Clone()
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := NewClient(Config{
		BaseURL:        server.URL,
		APIKey:         "axis-key",
		ProductSurface: "ponti",
		TimeoutMS:      1000,
	})
	status, raw, err := client.DoJSON(context.Background(), CallContext{
		OrgID:      "org-1",
		ActorID:    "user-1",
		OnBehalfOf: "user-1",
		Scopes:     []string{"companion:tasks:write", "ponti:insights:read"},
	}, http.MethodPost, "/v1/chat", map[string]any{"message": "hola"})
	if err != nil {
		t.Fatalf("DoJSON returned error: %v", err)
	}
	if status != http.StatusOK {
		t.Fatalf("status=%d raw=%s", status, string(raw))
	}
	if gotHeaders.Get("X-API-Key") != "axis-key" {
		t.Fatalf("missing X-API-Key header")
	}
	if gotHeaders.Get("X-Org-ID") != "org-1" {
		t.Fatalf("missing X-Org-ID header")
	}
	if gotHeaders.Get("X-User-ID") != "user-1" {
		t.Fatalf("missing X-User-ID header")
	}
	if gotHeaders.Get("X-On-Behalf-Of") != "user-1" {
		t.Fatalf("missing X-On-Behalf-Of header")
	}
	if gotHeaders.Get("X-Product-Surface") != "ponti" {
		t.Fatalf("missing X-Product-Surface header")
	}
	if gotHeaders.Get("X-Auth-Scopes") != "companion:tasks:write ponti:insights:read" {
		t.Fatalf("unexpected scopes: %q", gotHeaders.Get("X-Auth-Scopes"))
	}
	if gotBody["message"] != "hola" {
		t.Fatalf("unexpected body: %#v", gotBody)
	}
}
