package governance

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	"github.com/google/uuid"

	nexusclient "github.com/devpablocristo/ponti-backend/internal/nexus"
)

// verifyCacheTTL limita la ventana en la que una verificación positiva se
// sirve desde memoria sin reconsultar Nexus.
const verifyCacheTTL = 30 * time.Second

// NotApprovedError indica que el request de Nexus no habilita la ejecución:
// no existe, no está aprobado, el action type no coincide o pertenece a otro
// tenant. Los handlers lo mapean a HTTP 412 (precondición de governance).
type NotApprovedError struct {
	Detail string
}

func (e *NotApprovedError) Error() string { return e.Detail }

// IsNotApproved reporta si el error (o su cadena de wrapping) es un
// NotApprovedError del verifier.
func IsNotApproved(err error) bool {
	var target *NotApprovedError
	return errors.As(err, &target)
}

// VerifierNexusPort es el subset del cliente Nexus que necesita el verifier.
type VerifierNexusPort interface {
	Get(ctx context.Context, requestID string, opts ...nexusclient.RequestOption) (nexusclient.Request, int, error)
}

// Verifier valida contra Nexus que un request id esté aprobado para el action
// type esperado antes de ejecutar un write gobernado. Fail closed: sin cliente
// o con Nexus inalcanzable la verificación falla con error 502/503, nunca
// permite el write.
type Verifier struct {
	nexus VerifierNexusPort

	mu    sync.Mutex
	cache map[string]verifiedRequest
}

type verifiedRequest struct {
	request   nexusclient.Request
	expiresAt time.Time
}

func NewVerifier(nexus VerifierNexusPort) *Verifier {
	return &Verifier{nexus: nexus, cache: map[string]verifiedRequest{}}
}

// VerifyApproved verifica que el request exista en Nexus, esté en estado
// allowed/approved, su action_type coincida con expectedActionType (con
// ActionTypeCapabilityInvoke aceptado como fallback legacy durante la
// transición a action types per-tool) y pertenezca al tenant. Solo los
// resultados positivos se cachean (TTL verifyCacheTTL) y cada hit revalida
// tenant y action type contra el request cacheado.
func (v *Verifier) VerifyApproved(ctx context.Context, tenantID uuid.UUID, nexusRequestID, expectedActionType string) error {
	nexusRequestID = strings.TrimSpace(nexusRequestID)
	if nexusRequestID == "" {
		return &NotApprovedError{Detail: "nexus request id is required"}
	}
	if v == nil || v.nexus == nil {
		// Sin cliente Nexus no hay forma de verificar: fail closed.
		return domainerr.Unavailable("nexus client not configured")
	}
	req, ok := v.cached(nexusRequestID)
	if !ok {
		fresh, st, err := v.nexus.Get(ctx, nexusRequestID, nexusclient.WithTenantID(tenantID.String()))
		if err != nil {
			return domainerr.UpstreamError("nexus verification failed")
		}
		if st == http.StatusNotFound {
			return &NotApprovedError{Detail: "nexus request not found"}
		}
		if st != http.StatusOK {
			return domainerr.UpstreamError(fmt.Sprintf("nexus verification status %d", st))
		}
		req = fresh
	}
	if err := validateApprovedRequest(req, tenantID, expectedActionType); err != nil {
		return err
	}
	v.store(nexusRequestID, req)
	return nil
}

func validateApprovedRequest(req nexusclient.Request, tenantID uuid.UUID, expectedActionType string) error {
	if req.Status != nexusclient.StatusAllowed && req.Status != nexusclient.StatusApproved {
		return &NotApprovedError{Detail: fmt.Sprintf("nexus request status %q is not approved", req.Status)}
	}
	expectedActionType = strings.TrimSpace(expectedActionType)
	if req.ActionType != expectedActionType && req.ActionType != nexusclient.ActionTypeCapabilityInvoke {
		return &NotApprovedError{Detail: fmt.Sprintf("nexus request action type %q does not match %q", req.ActionType, expectedActionType)}
	}
	if !ownedByTenant(req.OrgID, tenantID) {
		return &NotApprovedError{Detail: "nexus request does not belong to tenant"}
	}
	return nil
}

func (v *Verifier) cached(nexusRequestID string) (nexusclient.Request, bool) {
	v.mu.Lock()
	defer v.mu.Unlock()
	entry, ok := v.cache[nexusRequestID]
	if !ok {
		return nexusclient.Request{}, false
	}
	if time.Now().After(entry.expiresAt) {
		delete(v.cache, nexusRequestID)
		return nexusclient.Request{}, false
	}
	return entry.request, true
}

func (v *Verifier) store(nexusRequestID string, req nexusclient.Request) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.cache[nexusRequestID] = verifiedRequest{request: req, expiresAt: time.Now().Add(verifyCacheTTL)}
}
