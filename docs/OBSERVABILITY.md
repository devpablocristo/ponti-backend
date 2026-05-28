# Observability

## TLDR

- **Logs**: `slog` JSON a stdout, enriquecido con `request_id` + `trace_id` + `span_id` + (post-auth) `user_id` + `tenant_id` + `role`.
- **Métricas**: Prometheus RED metrics HTTP + Go runtime expuestas en `GET /observability/metrics`. Counter custom `crudar_rejected_archived_ref_total{table}`.
- **Tracing**: OTel SDK con exporter configurable (`OTEL_EXPORTER`: `otlp` / `stdout` / `none`). Off por default; cero overhead cuando inactivo.

Todo el wiring vive en `platform/observability/go` (lib compartida). Ponti solo consume.

---

## Logs

### Setup
`cmd/api/main.go:16` construye el logger con `observability.NewJSONLogger("ponti-backend")` y lo registra como default vía `slog.SetDefault`. Cualquier `slog.Info(...)` desde cualquier paquete sale ya formateado.

### Atributos comunes del access log

```
{
  "time": "2026-05-22T12:35:11.892Z",
  "level": "INFO",
  "msg": "http request completed",
  "service": "ponti-backend",
  "request_id": "fe6b6e801b42356ee0178be5",  // siempre
  "trace_id": "edf6fcf0e9c31961...",          // si OTEL_EXPORTER != none
  "span_id": "25aa7039e863d90f",              // idem
  "user_id": "local:soalenadmin25@ponti.local", // si pasó por auth middleware
  "tenant_id": "652f757e-465a-419e-a529-04a0a81e2803", // idem
  "role": "admin",                             // idem
  "event": "http_request_completed",
  "method": "GET",
  "path": "/api/v1/insights",
  "route": "/api/v1/insights",
  "status": 200,
  "duration_ms": 11,
  "remote_addr": "172.18.0.1"
}
```

### Enriquecimiento progresivo
Cada middleware que conoce un nuevo atributo lo agrega al logger del context:

1. `pkgmwr.ObservabilityWithMetrics`: setea `request_id`, abre span OTel, agrega `trace_id`+`span_id`.
2. `pkgmwr.RequireIdentityPlatformAuthz` (o `RequireLocalDevAuthz` en dev): agrega `user_id`+`tenant_id`+`role`.

Handlers/repos llaman `observability.LoggerFromContext(ctx)` y obtienen el logger más enriquecido hasta ese punto. No hay que pasar el logger manualmente.

---

## Métricas (Prometheus)

### Endpoint
`GET /observability/metrics` — formato exposition Prometheus.

> Nota: hay `/metrics` distintos en handlers de negocio (`lot.GetMetrics`, `labor.GetMetrics`, `work-order.GetMetrics`) que devuelven KPIs del dominio. El path Prometheus está a propósito en `/observability/metrics` para no colisionar.

### Setup
`cmd/api/main.go:19` construye `metrics := observability.NewMetrics(observability.DefaultMetricsConfig("ponti_backend"))`. El gin engine usa `pkgmwr.ObservabilityWithMetrics(logger, metrics)` que cuenta cada request.

### Métricas built-in

| Métrica | Tipo | Labels |
|---|---|---|
| `ponti_backend_http_requests_total` | counter | `method`, `route`, `status_code` |
| `ponti_backend_http_request_errors_total` | counter | `method`, `route`, `status_code` (solo 4xx/5xx) |
| `ponti_backend_http_request_duration_seconds` | histogram | `method`, `route` |
| `go_goroutines`, `go_memstats_*`, `go_gc_*`, `process_*` | varias | runtime + process |

**Cardinalidad controlada**: el label `route` es el patrón Gin (`/api/v1/customers/:customer_id`), no la URL concreta. Eso se logra seteando `r.Pattern = METHOD + " " + c.FullPath()` antes de `metrics.ObserveHTTPRequest` en el middleware (Gin no setea `r.Pattern` automáticamente, así que `platform/observability/go.routeLabel` colapsaría todo en `"unmatched"` si no lo hicieramos).

### Métricas custom

| Métrica | Tipo | Labels | Cuándo se incrementa |
|---|---|---|---|
| `ponti_backend_crudar_rejected_archived_ref_total` | counter | `table` | `lifecycle.RequireActive` detecta un FK apuntando a un row archivado y retorna `domainerr.Conflict` |

**Interpretación**: subidas en esta métrica indican que la UI/import está enviando referencias inválidas que la barrera BE atrapa. Es señal de gap UX (selectores que no filtran archivados) a investigar.

Wiring: `lifecycle.RegisterMetrics(metrics.Registry(), "ponti_backend")` en `cmd/api/main.go`.

---

## Tracing (OpenTelemetry)

### Configuración (env vars)

| Variable | Default | Valores |
|---|---|---|
| `OTEL_EXPORTER` | `none` | `otlp` (HTTP), `stdout` (JSON local), `none` (no-op) |
| `OTEL_OTLP_ENDPOINT` | `` | host:port, e.g. `tempo:4318` o `localhost:4318` |
| `OTEL_OTLP_INSECURE` | `true` | `true`/`false` (HTTP vs HTTPS) |
| `OTEL_SAMPLE_RATIO` | `1.0` | `0.0`–`1.0` (parent-based ratio sampler) |
| `ENVIRONMENT` | `local` | resource attr `deployment.environment.name` |
| `SERVICE_VERSION` | `0.0.0` | resource attr `service.version` |

### Modos por entorno

- **Local dev**: dejar `OTEL_EXPORTER` unset → no-op silencioso, cero overhead, logs no llevan `trace_id`. Para debugging local con visibilidad, setear `OTEL_EXPORTER=stdout` y los spans salen JSON a stdout.
- **Staging/prod**: `OTEL_EXPORTER=otlp` + `OTEL_OTLP_ENDPOINT=<collector>:4318`. Sample ratio puede bajarse (e.g. `0.1` para 10% en prod).

### Propagación
W3C TraceContext (`traceparent`) + Baggage entrantes se respetan vía `propagation.NewCompositeTextMapPropagator`. Servicios upstream que ya tracean propagan a Ponti, y Ponti propaga downstream cuando llame a otros servicios (cuando se cableen).

### Span por request
El middleware Gin abre un span server-kind con name `"METHOD /api/v1/route"` y atributos semconv (`http.request.method`, `url.path`, `http.route`, `user_agent.original`, `net.peer.ip`, `http.response.status_code`). Status `Error` cuando `status >= 500`.

### Stack local opcional
Para ver spans en Jaeger/Tempo localmente, agregar al docker-compose un OTLP collector. No incluido por default para no inflar el stack de dev.

---

## Operación

### Cómo activar tracing en dev sin tocar código
```bash
# en .env del core
OTEL_EXPORTER=stdout
OTEL_SAMPLE_RATIO=1.0
ENVIRONMENT=local
```
Restart: `docker compose -f core/docker-compose.yml up -d --force-recreate ponti-api` (force-recreate es necesario para releer .env).

### Cómo confirmar que está activo
```bash
curl http://localhost:8080/api/v1/ping
docker logs core-ponti-api-1 | grep http_request_completed | tail -1
```
Si el log tiene `"trace_id":"..."`, está activo. Si no, está en no-op.

### Cómo ver métricas custom de CRUDAR
```bash
# Provocar un rechazo (intentar crear algo con FK archivada) y luego:
curl http://localhost:8080/observability/metrics | grep crudar_rejected
```
La métrica solo aparece después del primer rechazo (Prometheus expone `CounterVec` con label lazy).

---

## Decisiones de diseño

- **Logger lazy en context, no global**: el logger se enriquece y se re-setea en `context.Context`. Eso evita variables globales mutables y permite que tests aíslen el logger por request.
- **Default OFF para tracing**: el cost de OTel SDK init no es cero. Por default queda no-op (`OTEL_EXPORTER=none`) para no penalizar dev/CI. Se activa explícitamente.
- **Métricas Prometheus reusando `platform/observability/go`**: la lib compartida ya provee `NewMetrics`, `MiddlewareWithMetrics`, `ObserveHTTPRequest`. Ponti solo instancia y cablea, no reinventa.
- **Endpoint Prometheus en `/observability/metrics` (no `/metrics`)**: evita colisión con endpoints de KPIs de dominio que históricamente usan `/metrics`.
- **Métricas custom van con namespace del servicio**: `crudar_rejected_archived_ref_total` se registra con `Namespace: "ponti_backend"` para alineación con dashboards.

---

## Referencias
- Lib compartida: `github.com/devpablocristo/platform/observability/go` v0.2.1+
- Middleware: [`internal/platform/http/middlewares/gin/observability.go`](../internal/platform/http/middlewares/gin/observability.go)
- Counter CRUDAR: [`internal/shared/lifecycle/metrics.go`](../internal/shared/lifecycle/metrics.go)
- Bootstrap: [`cmd/api/main.go`](../cmd/api/main.go)
