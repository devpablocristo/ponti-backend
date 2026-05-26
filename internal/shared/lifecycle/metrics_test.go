package lifecycle

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestRegisterMetrics_RegistersCounter(t *testing.T) {
	resetRejectedCounterForTest(t)

	reg := prometheus.NewRegistry()
	RegisterMetrics(reg, "ponti_backend")

	if rejectedArchivedRefCounter == nil {
		t.Fatal("expected counter to be registered, got nil")
	}

	observeRejectedArchivedRef("actors")
	observeRejectedArchivedRef("actors")
	observeRejectedArchivedRef("fields")

	expected := `
# HELP ponti_backend_crudar_rejected_archived_ref_total Writes rejected because a referenced row is archived (lifecycle.RequireActive returned Conflict).
# TYPE ponti_backend_crudar_rejected_archived_ref_total counter
ponti_backend_crudar_rejected_archived_ref_total{table="actors"} 2
ponti_backend_crudar_rejected_archived_ref_total{table="fields"} 1
`
	if err := testutil.CollectAndCompare(rejectedArchivedRefCounter, strings.NewReader(expected)); err != nil {
		t.Fatalf("counter mismatch: %v", err)
	}
}

func TestRegisterMetrics_IdempotentAndNilSafe(t *testing.T) {
	resetRejectedCounterForTest(t)

	// nil registry: no-op, sin panic.
	RegisterMetrics(nil, "ponti_backend")
	if rejectedArchivedRefCounter != nil {
		t.Fatal("nil registry should not initialize counter")
	}

	// Doble llamada: la segunda no debe re-registrar ni romper.
	reg := prometheus.NewRegistry()
	RegisterMetrics(reg, "ponti_backend")
	first := rejectedArchivedRefCounter
	RegisterMetrics(reg, "ponti_backend")
	if rejectedArchivedRefCounter != first {
		t.Fatal("second RegisterMetrics call should be no-op")
	}
}

func TestObserveRejectedArchivedRef_NilCounterIsNoop(t *testing.T) {
	resetRejectedCounterForTest(t)
	// No panic cuando RegisterMetrics no se llamó (production-safe en tests).
	observeRejectedArchivedRef("anything")
}

// resetRejectedCounterForTest pone el counter en nil para aislar tests.
// El package-level singleton es intencional en prod (registro único en
// bootstrap), pero acá necesitamos restart por test.
func resetRejectedCounterForTest(t *testing.T) {
	t.Helper()
	rejectedArchivedRefCounter = nil
}
