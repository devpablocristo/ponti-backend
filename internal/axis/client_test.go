package axis

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-internal-jwt-secret-aaaaaaaaaaaaaaaaaaaa"

func newTestClient(t *testing.T, baseURL string) *CompanionClient {
	t.Helper()
	c, err := NewCompanionClient(Config{
		BaseURL:     baseURL,
		JWTSecret:   testSecret,
		JWTIssuer:   "ponti-test",
		JWTAudience: "companion",
		Timeout:     2 * time.Second,
		MaxRetries:  0,
	})
	if err != nil {
		t.Fatalf("NewCompanionClient: %v", err)
	}
	return c
}

func TestNewCompanionClient_ReturnsErrNotConfiguredWhenBaseURLEmpty(t *testing.T) {
	_, err := NewCompanionClient(Config{BaseURL: ""})
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("expected ErrNotConfigured, got %v", err)
	}
}

func TestNewCompanionClient_FailsWithoutSecret(t *testing.T) {
	_, err := NewCompanionClient(Config{BaseURL: "http://x"})
	if err == nil || errors.Is(err, ErrNotConfigured) {
		t.Fatalf("expected secret-required error, got %v", err)
	}
}

func TestChat_SignsJWTAndForwardsBody(t *testing.T) {
	var seenAuth string
	var seenBody ChatRequest

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat" || r.Method != http.MethodPost {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		seenAuth = r.Header.Get("Authorization")
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &seenBody)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(ChatResponse{
			ChatID: "chat-1",
			Reply:  "hi",
			Task:   Task{ID: "task-1", OrgID: "org-99", Status: "new"},
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	resp, err := c.Chat(context.Background(),
		CallContext{OrgID: "org-99", Actor: "user@ponti.local", Scopes: []string{"companion:tasks:write"}},
		ChatRequest{Message: "hola", ProductSurface: "ponti"},
	)
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if resp.Reply != "hi" || resp.Task.OrgID != "org-99" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if !strings.HasPrefix(seenAuth, "Bearer ") {
		t.Fatalf("expected Bearer token, got %q", seenAuth)
	}

	// Verificar firma + claims del JWT enviado.
	tokenStr := strings.TrimPrefix(seenAuth, "Bearer ")
	parsed, err := jwt.Parse(tokenStr, func(*jwt.Token) (any, error) { return []byte(testSecret), nil })
	if err != nil || !parsed.Valid {
		t.Fatalf("token invalid: %v", err)
	}
	claims := parsed.Claims.(jwt.MapClaims)
	if claims["org_id"] != "org-99" {
		t.Fatalf("expected org_id=org-99 in claims, got %v", claims["org_id"])
	}
	if claims["actor"] != "user@ponti.local" {
		t.Fatalf("expected actor=user@ponti.local, got %v", claims["actor"])
	}
	if claims["iss"] != "ponti-test" || claims["aud"] != "companion" {
		t.Fatalf("expected iss/aud, got iss=%v aud=%v", claims["iss"], claims["aud"])
	}

	if seenBody.Message != "hola" || seenBody.ProductSurface != "ponti" {
		t.Fatalf("body not forwarded as expected: %+v", seenBody)
	}
}

func TestChat_DefaultsScopesWhenEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		parsed, _ := jwt.Parse(token, func(*jwt.Token) (any, error) { return []byte(testSecret), nil })
		claims := parsed.Claims.(jwt.MapClaims)
		scopesAny, ok := claims["scopes"].([]any)
		if !ok || len(scopesAny) == 0 {
			t.Errorf("expected default scopes in JWT, got %v", claims["scopes"])
		}
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(ChatResponse{Task: Task{}})
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, err := c.Chat(context.Background(),
		CallContext{OrgID: "o", Actor: "a"},
		ChatRequest{Message: "x"})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
}

func TestChat_MapsHTTPErrors(t *testing.T) {
	cases := []struct {
		name     string
		status   int
		wantMsg  string
	}{
		{"unauthorized", 401, "authentication failed"},
		{"forbidden", 403, "forbidden"},
		{"not found", 404, "not found"},
		{"upstream 500", 500, "upstream error"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(`{"error":"x"}`))
			}))
			defer srv.Close()

			c := newTestClient(t, srv.URL)
			_, err := c.Chat(context.Background(), CallContext{OrgID: "o", Actor: "a"}, ChatRequest{Message: "x"})
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tc.wantMsg) {
				t.Fatalf("expected %q in error, got %q", tc.wantMsg, err.Error())
			}
		})
	}
}

func TestListConversations_ClampsLimit(t *testing.T) {
	cases := []struct {
		in       int
		wantPath string
	}{
		{0, "/v1/chat/conversations?limit=50"},
		{-5, "/v1/chat/conversations?limit=50"},
		{25, "/v1/chat/conversations?limit=25"},
		{500, "/v1/chat/conversations?limit=200"},
	}
	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			var gotPath string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path + "?" + r.URL.RawQuery
				w.WriteHeader(200)
				_, _ = w.Write([]byte(`{"items":[]}`))
			}))
			defer srv.Close()

			c := newTestClient(t, srv.URL)
			_, err := c.ListConversations(context.Background(), CallContext{OrgID: "o", Actor: "a"}, tc.in)
			if err != nil {
				t.Fatalf("ListConversations: %v", err)
			}
			if gotPath != tc.wantPath {
				t.Fatalf("expected path %q, got %q", tc.wantPath, gotPath)
			}
		})
	}
}

func TestGetConversation_RequiresID(t *testing.T) {
	c := newTestClient(t, "http://localhost:0")
	_, err := c.GetConversation(context.Background(), CallContext{OrgID: "o", Actor: "a"}, "  ")
	if err == nil || !strings.Contains(err.Error(), "id required") {
		t.Fatalf("expected id required error, got %v", err)
	}
}

func TestGetConversation_PathEscapesID(t *testing.T) {
	var rawPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// EscapedPath() preserva el `%2F`; r.URL.Path lo decodifica.
		rawPath = r.URL.EscapedPath()
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"id":"x","messages":[]}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, err := c.GetConversation(context.Background(), CallContext{OrgID: "o", Actor: "a"}, "abc/def")
	if err != nil {
		t.Fatalf("GetConversation: %v", err)
	}
	if rawPath != "/v1/chat/conversations/abc%2Fdef" {
		t.Fatalf("expected path-escaped, got %q", rawPath)
	}
}
