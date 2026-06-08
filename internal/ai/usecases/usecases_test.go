package usecases

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	ctxkeys "github.com/devpablocristo/platform/security/go/contextkeys"
	"github.com/google/uuid"

	"github.com/devpablocristo/ponti-backend/internal/axis"
)

type fakeLegacyClient struct {
	calls int
}

func (f *fakeLegacyClient) Do(context.Context, string, string, any, string, string) (int, []byte, error) {
	f.calls++
	return http.StatusOK, []byte(`{"request_id":"legacy","output_kind":"chat_reply","content_language":"es","chat_id":"","reply":"legacy","tokens_used":0,"tool_calls":[],"pending_confirmations":[],"blocks":[],"routed_agent":"general","routing_source":"legacy"}`), nil
}

func (f *fakeLegacyClient) DoStream(context.Context, string, string, io.Reader, string, string, string) (*http.Response, error) {
	f.calls++
	return nil, nil
}

type fakeAxisClient struct {
	status int
	raw    []byte
	err    error

	calls int
	call  axis.CallContext
	body  map[string]any
}

func (f *fakeAxisClient) ProductSurface() string { return "ponti" }

func (f *fakeAxisClient) DoJSON(_ context.Context, call axis.CallContext, _ string, _ string, body any) (int, []byte, error) {
	f.calls++
	f.call = call
	raw, _ := json.Marshal(body)
	_ = json.Unmarshal(raw, &f.body)
	return f.status, f.raw, f.err
}

func TestUseCases_ChatAxis_AdaptsResponseAndThreadsIdentity(t *testing.T) {
	t.Parallel()
	orgID := uuid.New()
	ctx := context.WithValue(context.Background(), ctxkeys.OrgID, orgID)
	ctx = context.WithValue(ctx, ctxkeys.Scopes, []string{"api.read"})
	axisClient := &fakeAxisClient{
		status: http.StatusOK,
		raw: []byte(`{
			"chat_id":"11111111-1111-1111-1111-111111111111",
			"task_id":"task-1",
			"run_id":"run-1",
			"reply":"respuesta axis",
			"blocks":[{"type":"text","text":"respuesta axis"}],
			"tool_calls":[{"name":"ponti.insights.list"}],
			"agent_id":"ponti_insights"
		}`),
	}
	uc := NewUseCases(&fakeLegacyClient{}, axisClient, Config{
		Provider:       "axis",
		AxisEnabled:    true,
		ProductSurface: "ponti",
	})
	status, raw, err := uc.Chat(ctx, "user-1", "project-1", map[string]any{
		"message": "hola",
		"workspace": map[string]any{
			"project_id": 1,
		},
	})
	if err != nil {
		t.Fatalf("Chat returned error: %v", err)
	}
	if status != http.StatusOK {
		t.Fatalf("status=%d raw=%s", status, string(raw))
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if out["request_id"] != "run-1" {
		t.Fatalf("request_id=%v", out["request_id"])
	}
	if out["routing_source"] != "axis" {
		t.Fatalf("routing_source=%v", out["routing_source"])
	}
	if out["reply"] != "respuesta axis" {
		t.Fatalf("reply=%v", out["reply"])
	}
	if axisClient.call.OrgID != orgID.String() {
		t.Fatalf("org id not threaded: %#v", axisClient.call)
	}
	if axisClient.call.ProductSurface != "ponti" {
		t.Fatalf("product surface not threaded: %#v", axisClient.call)
	}
	if _, exists := axisClient.body["workspace"]; exists {
		t.Fatalf("workspace must not be sent top-level to Axis current chat contract: %#v", axisClient.body)
	}
	if _, exists := axisClient.body["handoff"]; !exists {
		t.Fatalf("workspace should be carried inside handoff for future Axis contract: %#v", axisClient.body)
	}
}

func TestUseCases_ChatAxis_ForbiddenDoesNotFallbackToLegacy(t *testing.T) {
	t.Parallel()
	orgID := uuid.New()
	ctx := context.WithValue(context.Background(), ctxkeys.OrgID, orgID)
	legacy := &fakeLegacyClient{}
	axisClient := &fakeAxisClient{
		status: http.StatusForbidden,
		raw:    []byte(`{"code":"FORBIDDEN","message":"blocked"}`),
	}
	uc := NewUseCases(legacy, axisClient, Config{
		Provider:       "axis",
		AxisEnabled:    true,
		ProductSurface: "ponti",
	})
	status, raw, err := uc.Chat(ctx, "user-1", "project-1", map[string]any{"message": "hola"})
	if err != nil {
		t.Fatalf("Chat returned error: %v", err)
	}
	if status != http.StatusForbidden {
		t.Fatalf("status=%d raw=%s", status, string(raw))
	}
	if legacy.calls != 0 {
		t.Fatalf("legacy fallback must not run for Axis 4xx, calls=%d", legacy.calls)
	}
}
