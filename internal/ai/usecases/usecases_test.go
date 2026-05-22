package usecases

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
)

// fakeClient implementa ClientPort. Permite scriptear (status, body, err) por llamada y
// capturar el path que el caller pidió.
type fakeClient struct {
	doStatus int
	doBody   []byte
	doErr    error

	calledMethod    string
	calledPath      string
	calledBody      any
	calledUserID    string
	calledTenantID  string
	calledProjectID string
}

func (f *fakeClient) Do(_ context.Context, method, path string, body any, userID, tenantID, projectID string) (int, []byte, error) {
	f.calledMethod = method
	f.calledPath = path
	f.calledBody = body
	f.calledUserID = userID
	f.calledTenantID = tenantID
	f.calledProjectID = projectID
	return f.doStatus, f.doBody, f.doErr
}

func (f *fakeClient) DoStream(_ context.Context, _, _ string, _ io.Reader, _, _, _, _ string) (*http.Response, error) {
	return nil, errors.New("not implemented in fake")
}

func TestIsAIServiceNotConfigured(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil err", nil, false},
		{"url not configured", errors.New("ai service url not configured"), true},
		{"key not configured", errors.New("ai service key not configured"), true},
		{"wrapping ok", errors.New("client failed: ai service url not configured: tcp"), true},
		{"random db err", errors.New("connection refused"), false},
		{"empty msg", errors.New(""), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isAIServiceNotConfigured(tc.err); got != tc.want {
				t.Fatalf("isAIServiceNotConfigured(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

func TestChat_HappyPath_PassesThroughClient(t *testing.T) {
	wantBody := []byte(`{"reply":"hola"}`)
	client := &fakeClient{doStatus: 200, doBody: wantBody}
	uc := NewUseCases(client)

	status, raw, err := uc.Chat(context.Background(), "user-1", "tenant-1", "proj-1", map[string]string{"msg": "hello"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != 200 {
		t.Fatalf("expected status 200, got %d", status)
	}
	if string(raw) != string(wantBody) {
		t.Fatalf("expected body %q, got %q", wantBody, raw)
	}
	if client.calledMethod != "POST" || client.calledPath != "/v1/chat" {
		t.Fatalf("expected POST /v1/chat, got %s %s", client.calledMethod, client.calledPath)
	}
	if client.calledUserID != "user-1" || client.calledTenantID != "tenant-1" || client.calledProjectID != "proj-1" {
		t.Fatalf("expected identity headers forwarded, got user=%q tenant=%q proj=%q",
			client.calledUserID, client.calledTenantID, client.calledProjectID)
	}
}

func TestChat_AINotConfigured_ReturnsDummy(t *testing.T) {
	client := &fakeClient{doErr: errors.New("ai service url not configured")}
	uc := NewUseCases(client)

	status, raw, err := uc.Chat(context.Background(), "u", "t", "", nil)
	if err != nil {
		t.Fatalf("expected nil error for dummy fallback, got %v", err)
	}
	if status != 200 {
		t.Fatalf("expected dummy status 200, got %d", status)
	}

	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("dummy body should be JSON, got %q (%v)", raw, err)
	}
	if payload["output_kind"] != "chat_reply" {
		t.Fatalf("expected output_kind=chat_reply in dummy, got %v", payload["output_kind"])
	}
	if payload["routing_source"] != "read_fallback" {
		t.Fatalf("expected routing_source=read_fallback, got %v", payload["routing_source"])
	}
}

func TestChat_OtherError_PropagatesUnchanged(t *testing.T) {
	client := &fakeClient{doErr: errors.New("upstream 500")}
	uc := NewUseCases(client)

	_, _, err := uc.Chat(context.Background(), "u", "t", "", nil)
	if err == nil || err.Error() != "upstream 500" {
		t.Fatalf("expected propagated 'upstream 500', got %v", err)
	}
}

func TestListChatConversations_ClampsLimit(t *testing.T) {
	client := &fakeClient{doStatus: 200, doBody: []byte(`{"items":[]}`)}
	uc := NewUseCases(client)
	ctx := context.Background()

	cases := []struct {
		name     string
		input    int
		wantPath string
	}{
		{"zero defaults to 50", 0, "/v1/chat/conversations?limit=50"},
		{"negative defaults to 50", -5, "/v1/chat/conversations?limit=50"},
		{"valid passes through", 25, "/v1/chat/conversations?limit=25"},
		{"too high clamps to 200", 500, "/v1/chat/conversations?limit=200"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := uc.ListChatConversations(ctx, "u", "t", "", tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if client.calledPath != tc.wantPath {
				t.Fatalf("expected path %q, got %q", tc.wantPath, client.calledPath)
			}
		})
	}
}

func TestListChatConversations_DummyFallback(t *testing.T) {
	client := &fakeClient{doErr: errors.New("ai service key not configured")}
	uc := NewUseCases(client)

	_, raw, err := uc.ListChatConversations(context.Background(), "u", "t", "", 100)
	if err != nil {
		t.Fatalf("expected nil for dummy fallback, got %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("dummy body should be JSON, got %q", raw)
	}
	items, ok := payload["items"].([]any)
	if !ok || len(items) != 0 {
		t.Fatalf("expected dummy items=[], got %v", payload["items"])
	}
}

func TestGetChatConversation_TrimsAndForwardsID(t *testing.T) {
	client := &fakeClient{doStatus: 200, doBody: []byte(`{}`)}
	uc := NewUseCases(client)

	_, _, err := uc.GetChatConversation(context.Background(), "u", "t", "", "  conv-42  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantPath := "/v1/chat/conversations/conv-42"
	if client.calledPath != wantPath {
		t.Fatalf("expected path %q, got %q", wantPath, client.calledPath)
	}
}

func TestGetChatConversation_DummyFallback(t *testing.T) {
	client := &fakeClient{doErr: errors.New("ai service url not configured")}
	uc := NewUseCases(client)

	_, raw, err := uc.GetChatConversation(context.Background(), "u", "t", "", "abc")
	if err != nil {
		t.Fatalf("expected nil for dummy fallback, got %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("dummy body should be JSON, got %q", raw)
	}
	if payload["id"] != "abc" {
		t.Fatalf("expected id=abc in dummy, got %v", payload["id"])
	}
}
