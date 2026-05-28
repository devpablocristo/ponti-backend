package lifecycle

import "github.com/prometheus/client_golang/prometheus"

// rejectedArchivedRefCounter cuenta cuántas operaciones (Create/Update/restore/
// archive child) fueron rechazadas porque referenciaban un row archivado. Sirve
// como señal de calidad del invariante "archived = no existe": si este contador
// sube significa que la UI/import está dejando pasar referencias inválidas y
// la barrera BE las atrapa — gap de UX a corregir.
//
// Por default es nil (no-op) para no acoplar el paquete a Prometheus en tests
// ni forzar registro global. Se activa explícitamente vía RegisterMetrics
// durante el bootstrap del servicio.
var rejectedArchivedRefCounter *prometheus.CounterVec

// RegisterMetrics registra los counters del paquete lifecycle en la registry
// dada. Llamar UNA SOLA VEZ en el bootstrap del servicio. Si `registry` es nil
// o ya hubo registro previo, no hace nada.
//
// `namespace` debe coincidir con el de las demás métricas del servicio (ver
// observability.DefaultMetricsConfig) para mantener consistencia en dashboards.
func RegisterMetrics(registry *prometheus.Registry, namespace string) {
	if registry == nil || rejectedArchivedRefCounter != nil {
		return
	}
	rejectedArchivedRefCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "crudar_rejected_archived_ref_total",
			Help:      "Writes rejected because a referenced row is archived (lifecycle.RequireActive returned Conflict).",
		},
		[]string{"table"},
	)
	registry.MustRegister(rejectedArchivedRefCounter)
}

func observeRejectedArchivedRef(table string) {
	if rejectedArchivedRefCounter == nil {
		return
	}
	rejectedArchivedRefCounter.WithLabelValues(table).Inc()
}
