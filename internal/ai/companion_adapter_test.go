package ai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/devpablocristo/ponti-backend/internal/axis"
)

const adapterTestSecret = "test-internal-jwt-secret-aaaaaaaaaaaaaaaaaaaa"

func newAdapter(t *testing.T, baseURL string) *CompanionAdapter {
	t.Helper()
	client, err := axis.NewCompanionClient(axis.Config{
		BaseURL:     baseURL,
		JWTSecret:   adapterTestSecret,
		JWTIssuer:   "ponti-test",
		JWTAudience: "companion",
		Timeout:     2 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewCompanionClient: %v", err)
	}
	return NewCompanionAdapter(client)
}

func TestCompanionAdapter_Chat_RoundtripsResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat" {
			t.Fatalf("expected /v1/chat, got %s", r.URL.Path)
		}
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"chat_id":        "chat-1",
			"reply":          "hola desde companion",
			"routed_agent":   "general",
			"routing_source": "orchestrator",
			"task":           map[string]any{"id": "task-1", "org_id": "org-1"},
			"messages":       []any{},
		})
	}))
	defer srv.Close()

	a := newAdapter(t, srv.URL)
	status, body, err := a.Do(context.Background(), "POST", "/v1/chat",
		map[string]any{"message": "hola"},
		"user-1", "org-1", "proj-99",
	)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
	var out map[string]any
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out["reply"] != "hola desde companion" {
		t.Fatalf("expected reply roundtripped, got %v", out["reply"])
	}
	if out["routed_agent"] != "general" || out["routing_source"] != "orchestrator" {
		t.Fatalf("routing fields missing: %+v", out)
	}
	if out["request_id"] != "task-1" {
		t.Fatalf("request_id should default to task.id, got %v", out["request_id"])
	}
}

func TestCompanionAdapter_Chat_RejectsEmptyMessage(t *testing.T) {
	a := newAdapter(t, "http://localhost:0")
	status, _, err := a.Do(context.Background(), "POST", "/v1/chat",
		map[string]any{"message": ""},
		"u", "o", "")
	if err == nil {
		t.Fatal("expected error")
	}
	if status != 400 {
		t.Fatalf("expected 400, got %d", status)
	}
}

// Nota: Ponti mantiene `route_hint` como detalle del FE, pero Companion todavía
// no consume project context en el contrato conversacional.
func TestCompanionAdapter_Chat_DoesNotSendRouteHint(t *testing.T) {
	var hasRouteHint bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var b map[string]any
		_ = json.Unmarshal(raw, &b)
		_, hasRouteHint = b["route_hint"]
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]any{"reply": "ok", "task": map[string]any{"id": "t"}, "messages": []any{}})
	}))
	defer srv.Close()

	a := newAdapter(t, srv.URL)
	_, _, err := a.Do(context.Background(), "POST", "/v1/chat",
		map[string]any{"message": "x"},
		"u", "o", "proj-77")
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if hasRouteHint {
		t.Fatal("expected adapter NOT to send route_hint")
	}
}

func TestCompanionAdapter_Chat_SendsChatIDAsChatID(t *testing.T) {
	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &got)
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"chat_id":  got["chat_id"],
			"reply":    "ok",
			"task":     map[string]any{"id": "task-1"},
			"messages": []any{},
		})
	}))
	defer srv.Close()

	a := newAdapter(t, srv.URL)
	_, _, err := a.Do(context.Background(), "POST", "/v1/chat",
		map[string]any{"message": "seguimos", "chat_id": "8ee190ab-80c8-4242-a03b-f00bd185a956"},
		"u", "o", "proj-77")
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if got["chat_id"] != "8ee190ab-80c8-4242-a03b-f00bd185a956" {
		t.Fatalf("expected chat_id forwarded, got body %+v", got)
	}
	if _, ok := got["task_id"]; ok {
		t.Fatalf("did not expect chat_id to be sent as task_id: %+v", got)
	}
}

func TestCompanionAdapter_ListConversations_PassesLimit(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path + "?" + r.URL.RawQuery
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"items":[]}`))
	}))
	defer srv.Close()

	a := newAdapter(t, srv.URL)
	_, _, err := a.Do(context.Background(), "GET", "/v1/chat/conversations?limit=7", nil, "u", "o", "")
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if gotPath != "/v1/chat/conversations?limit=7" {
		t.Fatalf("expected limit=7 forwarded, got %q", gotPath)
	}
}

func TestCompanionAdapter_GetConversation_RoutesByPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"id":"conv-1","messages":[]}`))
	}))
	defer srv.Close()

	a := newAdapter(t, srv.URL)
	status, _, err := a.Do(context.Background(), "GET", "/v1/chat/conversations/conv-1", nil, "u", "o", "")
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if status != 200 || gotPath != "/v1/chat/conversations/conv-1" {
		t.Fatalf("unexpected status=%d path=%q", status, gotPath)
	}
}

func TestCompanionAdapter_GetConversation_MapsCanonicalMessagesForFrontend(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{
			"id":"8ee190ab-80c8-4242-a03b-f00bd185a956",
			"title":"historial",
			"created_at":"2026-05-25T18:00:00Z",
			"updated_at":"2026-05-25T18:01:00Z",
			"messages":[
				{"role":"user","content":"hola","timestamp":"2026-05-25T18:00:00Z"},
				{"role":"assistant","content":"respuesta","timestamp":"2026-05-25T18:01:00Z"}
			]
		}`))
	}))
	defer srv.Close()

	a := newAdapter(t, srv.URL)
	status, body, err := a.Do(context.Background(), "GET", "/v1/chat/conversations/8ee190ab-80c8-4242-a03b-f00bd185a956", nil, "u", "o", "")
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
	var out struct {
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
			TS      string `json:"ts"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(out.Messages) != 2 || out.Messages[0].Content != "hola" || out.Messages[1].Content != "respuesta" {
		t.Fatalf("messages not mapped for frontend: %+v", out.Messages)
	}
	if out.Messages[0].TS == "" {
		t.Fatalf("expected timestamp mapped to ts: %+v", out.Messages[0])
	}
}

func TestCompanionAdapter_UnsupportedRoute_Returns404(t *testing.T) {
	a := newAdapter(t, "http://localhost:0")
	status, _, err := a.Do(context.Background(), "POST", "/v1/insights/compute", nil, "u", "o", "")
	if err == nil {
		t.Fatal("expected error")
	}
	if status != 404 {
		t.Fatalf("expected 404, got %d", status)
	}
}

func TestCompanionAdapter_DoStream_EmitsTwoSSEEvents(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"chat_id":  "chat-stream",
			"reply":    "respuesta sincrónica",
			"task":     map[string]any{"id": "t"},
			"messages": []any{},
		})
	}))
	defer srv.Close()

	a := newAdapter(t, srv.URL)
	body := strings.NewReader(`{"message":"hola"}`)
	resp, err := a.DoStream(context.Background(), "POST", "/v1/chat/stream", body, "application/json", "u", "o", "")
	if err != nil {
		t.Fatalf("DoStream: %v", err)
	}
	if got := resp.Header.Get("Content-Type"); !strings.Contains(got, "text/event-stream") {
		t.Fatalf("expected SSE content-type, got %q", got)
	}
	raw, _ := io.ReadAll(resp.Body)
	stream := string(raw)
	if !strings.Contains(stream, "event: start") {
		t.Errorf("missing start event in stream: %q", stream)
	}
	if !strings.Contains(stream, "event: done") {
		t.Errorf("missing done event in stream: %q", stream)
	}
	if !strings.Contains(stream, "respuesta sincrónica") {
		t.Errorf("done event missing reply: %q", stream)
	}
}
