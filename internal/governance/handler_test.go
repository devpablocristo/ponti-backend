package governance_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/devpablocristo/ponti-backend/internal/governance"
	nexusclient "github.com/devpablocristo/ponti-backend/internal/nexus"
)

type testGinEngine struct{ router *gin.Engine }

func (e testGinEngine) GetRouter() *gin.Engine { return e.router }

type testAPIConfig struct{}

func (testAPIConfig) APIVersion() string { return "v1" }
func (testAPIConfig) APIBaseURL() string { return "/api/v1" }

type testMiddlewares struct{}

func (testMiddlewares) GetValidation() []gin.HandlerFunc { return nil }

func newCallbackTestRouter(t *testing.T, svc *governance.Service) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	h := governance.NewHandler(svc, testGinEngine{router: router}, testAPIConfig{}, testMiddlewares{})
	h.Routes()
	return router
}

// TestNexusCallback_TimestampAntiReplay cubre la ventana anti-replay del
// callback: una firma HMAC válida con timestamp ausente, no parseable o fuera
// de ±5 minutos debe rechazarse con 401 (un (timestamp, payload, firma)
// capturado no puede validar para siempre).
func TestNexusCallback_TimestampAntiReplay(t *testing.T) {
	const token = "nexus-callback-secret"
	tenantID := uuid.New()
	payload := []byte(fmt.Sprintf(`{"event":"approval_pending","approval_id":"appr-1","org_id":%q,"request_id":"req-1"}`, tenantID.String()))
	now := time.Now().UTC()

	cases := []struct {
		name      string
		timestamp string
		wantCode  int
	}{
		{"fresh RFC3339Nano accepted", now.Format(time.RFC3339Nano), http.StatusOK},
		{"fresh RFC3339 accepted", now.Format(time.RFC3339), http.StatusOK},
		{"stale rejected", now.Add(-10 * time.Minute).Format(time.RFC3339Nano), http.StatusUnauthorized},
		{"future skew rejected", now.Add(10 * time.Minute).Format(time.RFC3339Nano), http.StatusUnauthorized},
		{"missing rejected", "", http.StatusUnauthorized},
		{"garbage rejected", "not-a-timestamp", http.StatusUnauthorized},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := newStubRepo()
			repo.rows["req-1"] = governance.RequestRecord{
				TenantID:       tenantID,
				NexusRequestID: "req-1",
				ActionType:     "workorder.create",
				Origin:         "agent",
				Status:         nexusclient.StatusPendingApproval,
				ApprovalID:     "appr-1",
			}
			svc := governance.NewService(repo, &stubNexus{}, governance.Config{CallbackToken: token}, nil)
			router := newCallbackTestRouter(t, svc)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/governance/callbacks/nexus", bytes.NewReader(payload))
			if tc.timestamp != "" {
				req.Header.Set("X-Nexus-Callback-Timestamp", tc.timestamp)
			}
			// La firma siempre es válida para el timestamp enviado: lo único que
			// distingue los casos es la frescura del timestamp.
			req.Header.Set("X-Nexus-Callback-Signature", signCallback(token, tc.timestamp, payload))
			res := httptest.NewRecorder()
			router.ServeHTTP(res, req)

			if res.Code != tc.wantCode {
				t.Fatalf("status = %d, want %d (body=%s)", res.Code, tc.wantCode, res.Body.String())
			}
		})
	}
}

func TestNexusCallback_InvalidSignatureRejected(t *testing.T) {
	const token = "nexus-callback-secret"
	svc := governance.NewService(newStubRepo(), &stubNexus{}, governance.Config{CallbackToken: token}, nil)
	router := newCallbackTestRouter(t, svc)

	payload := []byte(`{"event":"approval_pending","request_id":"req-1"}`)
	timestamp := time.Now().UTC().Format(time.RFC3339Nano)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/governance/callbacks/nexus", bytes.NewReader(payload))
	req.Header.Set("X-Nexus-Callback-Timestamp", timestamp)
	req.Header.Set("X-Nexus-Callback-Signature", signCallback("wrong-token", timestamp, payload))
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401 (body=%s)", res.Code, res.Body.String())
	}
}
