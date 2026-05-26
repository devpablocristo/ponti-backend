package usecases

import (
	"context"
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

func TestChat_PropagatesUpstreamError(t *testing.T) {
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
