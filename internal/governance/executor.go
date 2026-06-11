package governance

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	nexusclient "github.com/devpablocristo/ponti-backend/internal/nexus"
)

const (
	// executorAttester identifica a Ponti como firmante de las attestations.
	executorAttester = "ponti-backend"
	// executorReportTimeout acota el ciclo async ReportResult/Attest/Evidence.
	executorReportTimeout = 10 * time.Second
)

// executorDraftTypes mapea los action types ejecutables al draft_type local.
// Los action types fuera de este set se loggean y marcan como error en la fila
// sin romper el callback (workorder drafts los ejecuta Axis llamando al
// endpoint HTTP, no el callback).
var executorDraftTypes = map[string]string{
	nexusclient.ActionTypeInsightResolve:  "insight_resolution",
	nexusclient.ActionTypeStockCountApply: "stock_count",
}

// ActionDispatcher aplica una acción aprobada; lo implementa el ActionExecutor
// de internal/ai y se conecta en bootstrap (mismo patrón que los notifiers).
type ActionDispatcher interface {
	DispatchApproved(ctx context.Context, tenantID uuid.UUID, actionType, nexusRequestID string, params map[string]any, actor string) (map[string]any, error)
}

// ExecutorNexusPort es el subset del cliente Nexus que usa el executor para
// cerrar el loop de governance (resultado, attestation y evidence pack).
type ExecutorNexusPort interface {
	ReportResult(ctx context.Context, requestID string, body nexusclient.ReportResultBody, opts ...nexusclient.RequestOption) (int, error)
	Attest(ctx context.Context, requestID string, body any, opts ...nexusclient.RequestOption) (int, []byte, error)
	GetEvidence(ctx context.Context, requestID string, opts ...nexusclient.RequestOption) ([]byte, int, error)
}

// ExecutorConfig agrupa la configuración del executor (subset de config.Nexus).
type ExecutorConfig struct {
	GovernedWritesEnabled bool
	AttestationHMACSecret string
}

// ApprovedExecutor ejecuta la acción aprobada cuando llega el callback
// approval_resolved(approved): despacha al ActionDispatcher según el
// action_type de la fila local, registra executed_at/result_json/error_message
// y luego, async y best-effort, reporta el resultado a Nexus (ReportResult →
// Attest si hay secret → GetEvidence cacheado localmente). Con
// AI_GOVERNED_WRITES_ENABLED=false mantiene semántica no-op (solo loggea).
type ApprovedExecutor struct {
	repo       RepositoryPort
	nexus      ExecutorNexusPort
	cfg        ExecutorConfig
	dispatcher ActionDispatcher

	reportTimeout time.Duration
	wg            sync.WaitGroup
}

func NewApprovedExecutor(repo RepositoryPort, nexus ExecutorNexusPort, cfg ExecutorConfig) *ApprovedExecutor {
	return &ApprovedExecutor{repo: repo, nexus: nexus, cfg: cfg, reportTimeout: executorReportTimeout}
}

// SetDispatcher conecta el dispatcher real después de wire (el ActionExecutor
// de ai depende del Service de businessinsights, que se arma en bootstrap).
func (e *ApprovedExecutor) SetDispatcher(d ActionDispatcher) {
	e.dispatcher = d
}

// Wait bloquea hasta que terminen los reportes async pendientes. Pensado para
// tests y shutdown ordenado; el callback HTTP nunca lo llama.
func (e *ApprovedExecutor) Wait() {
	e.wg.Wait()
}

// ExecuteApproved implementa Executor. Nunca devuelve error por fallas de
// ejecución (las registra en la fila local): el callback de Nexus ya aplicó el
// estado terminal y no debe reintentarse por un problema de ejecución local.
func (e *ApprovedExecutor) ExecuteApproved(ctx context.Context, row RequestRecord) error {
	if e == nil {
		return nil
	}
	if !e.cfg.GovernedWritesEnabled {
		log.Printf("[governance] executor disabled (AI_GOVERNED_WRITES_ENABLED=false): request %s action %q queda sin ejecutar", row.NexusRequestID, row.ActionType)
		return nil
	}
	if row.ExecutedAt != nil {
		// Idempotencia: callback replayed sobre una fila ya ejecutada.
		return nil
	}
	if _, supported := executorDraftTypes[row.ActionType]; !supported {
		log.Printf("[governance] executor: action type %q de request %s no es ejecutable localmente", row.ActionType, row.NexusRequestID)
		e.recordExecution(ctx, row, nil, fmt.Sprintf("unsupported action type %q", row.ActionType), false)
		return nil
	}
	if e.dispatcher == nil {
		log.Printf("[governance] executor: dispatcher no configurado, request %s queda sin ejecutar", row.NexusRequestID)
		e.recordExecution(ctx, row, nil, "action dispatcher not configured", false)
		return nil
	}

	started := time.Now()
	result, err := e.dispatcher.DispatchApproved(ctx, row.TenantID, row.ActionType, row.NexusRequestID, executionParams(row), row.DecidedBy)
	duration := time.Since(started)
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
		log.Printf("[governance] executor: ejecución de request %s (%s) falló: %v", row.NexusRequestID, row.ActionType, err)
	}
	e.recordExecution(ctx, row, result, errMsg, true)
	e.reportAsync(row, err == nil, duration, result, errMsg)
	return nil
}

// recordExecution persiste el resultado del intento en la fila local. report
// indica si el intento llegó a despacharse (los action types no soportados
// solo marcan error_message, sin executed_at).
func (e *ApprovedExecutor) recordExecution(ctx context.Context, row RequestRecord, result map[string]any, errMsg string, executed bool) {
	if e.repo == nil {
		return
	}
	updates := map[string]any{}
	if executed {
		updates["executed_at"] = time.Now().UTC()
	}
	if errMsg != "" {
		updates["error_message"] = errMsg
	}
	if result != nil {
		updates["result_json"] = marshalJSON(result)
	}
	if len(updates) == 0 {
		return
	}
	if err := e.repo.UpdateByNexusRequestID(ctx, row.TenantID, row.NexusRequestID, updates); err != nil {
		log.Printf("[governance] executor: no se pudo registrar la ejecución de %s: %v", row.NexusRequestID, err)
	}
}

// reportAsync cierra el loop con Nexus en background sin bloquear la respuesta
// del callback: ReportResult, Attest (si hay NEXUS_ATTESTATION_HMAC_SECRET) y
// GetEvidence persistido en el cache local. Todo best-effort: cualquier falla
// solo se loggea.
func (e *ApprovedExecutor) reportAsync(row RequestRecord, success bool, duration time.Duration, result map[string]any, errMsg string) {
	if e.nexus == nil {
		return
	}
	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), e.reportTimeout)
		defer cancel()
		tenant := nexusclient.WithTenantID(row.TenantID.String())

		if _, err := e.nexus.ReportResult(ctx, row.NexusRequestID, nexusclient.ReportResultBody{
			Success:      success,
			Result:       result,
			DurationMS:   duration.Milliseconds(),
			ErrorMessage: errMsg,
		}, tenant); err != nil {
			log.Printf("[governance] executor: report result de %s falló: %v", row.NexusRequestID, err)
		}

		if secret := strings.TrimSpace(e.cfg.AttestationHMACSecret); secret != "" {
			canonical := canonicalResultJSON(row.NexusRequestID, success, result, errMsg)
			mac := hmac.New(sha256.New, []byte(secret))
			_, _ = mac.Write(canonical)
			attestation := map[string]any{
				"attester":  executorAttester,
				"algorithm": "hmac-sha256",
				"payload":   json.RawMessage(canonical),
				"signature": hex.EncodeToString(mac.Sum(nil)),
			}
			if _, _, err := e.nexus.Attest(ctx, row.NexusRequestID, attestation, tenant); err != nil {
				log.Printf("[governance] executor: attestation de %s falló: %v", row.NexusRequestID, err)
			}
		}

		raw, st, err := e.nexus.GetEvidence(ctx, row.NexusRequestID, tenant)
		if err != nil || st != http.StatusOK {
			log.Printf("[governance] executor: evidence de %s no disponible (status %d, err %v)", row.NexusRequestID, st, err)
			return
		}
		if !ownedByTenant(extractPackOrgID(raw), row.TenantID) {
			// Pack de otra org (o sin org): nunca cachearlo bajo este tenant.
			return
		}
		record := EvidenceRecord{TenantID: row.TenantID, NexusRequestID: row.NexusRequestID, PackJSON: raw}
		record.Signature, record.SignatureKeyID = extractPackSignature(raw)
		if e.repo != nil {
			if serr := e.repo.SaveEvidence(ctx, record); serr != nil {
				log.Printf("[governance] executor: cache de evidence de %s falló: %v", row.NexusRequestID, serr)
			}
		}
	}()
}

// canonicalResultJSON produce el JSON canónico del resultado para firmar la
// attestation (json.Marshal de map ordena las keys, es determinístico).
func canonicalResultJSON(requestID string, success bool, result map[string]any, errMsg string) []byte {
	canonical, err := json.Marshal(map[string]any{
		"request_id":    requestID,
		"success":       success,
		"result":        result,
		"error_message": errMsg,
	})
	if err != nil {
		return []byte("{}")
	}
	return canonical
}

// executionParams extrae los parámetros de la acción desde la fila local:
// params_json primero, payload_json.params como fallback (filas hidratadas
// desde el Get completo de Nexus).
func executionParams(row RequestRecord) map[string]any {
	if params := unmarshalMap(row.ParamsJSON); len(params) > 0 {
		return params
	}
	payload := unmarshalMap(row.PayloadJSON)
	if payload == nil {
		return map[string]any{}
	}
	if params, ok := payload["params"].(map[string]any); ok {
		return params
	}
	return map[string]any{}
}
