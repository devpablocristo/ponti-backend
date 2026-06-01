# risks.md — feature-012 (BE)

## Funcionales

| riesgo | impacto | mitigación |
|---|---|---|
| Pérdida de streaming token-por-token (Companion es síncrono; el adapter emite SSE `start`+`done` sintético) | UX degradada: el usuario ve la respuesta de golpe, no progresiva | Documentado en `companion_adapter.go`. FE debe leer `done`. Reescribir `DoStream` cuando Companion gane SSE real (no toca handler/FE) |
| `project_id` no se propaga a Companion (no tiene noción de project) | conversaciones no filtrables por proyecto en Companion | Validación tenant↔project local; filtrado por project futuro vía metadata de task |
| Cutover sin fallback: el binary no arranca sin `COMPANION_BASE_URL`/`SECRET` | downtime si la config no está en el deploy | Documentar env obligatorias; validar en CI/staging que arranca con config real |

## Técnicos

| riesgo | impacto | mitigación |
|---|---|---|
| `wire/wire_gen.go` no regenerado tras cambiar firma de `ProvideAIHandler` | no compila / mismatch de tipos | Regenerar con `wire` (no editar a mano); `go build ./...` antes del PR |
| JWT mal firmado (issuer/audience/scope) rechazado por Companion | 401 en todas las requests | Tests `client_test.go` validan claims; confirmar `iss`/`aud`/`scope` que espera Companion (`wire/auth.go::claimScopes` lee `scope`/`scp`) |
| Secret desincronizado entre Ponti y Companion | 401 silencioso en prod | `COMPANION_INTERNAL_JWT_SECRET` vía Secret Manager GCP; mismo valor en ambos lados |

## Integración / cross-repo

| riesgo | impacto | mitigación |
|---|---|---|
| Mergear BE sin FE feature-012 | el FE legacy puede no entender campos nuevos (aunque el shape es retrocompatible) | BE-first; el shape mantiene campos legacy (`request_id`, `tokens_used`, etc.). Validar el contrato `done` con el FE |
| Mergear FE sin BE | el FE apunta a un BE que aún usa cliente legacy ponti-ai | Orden estricto BE-first |
| DTOs de `axis/types.go` divergen del OpenAPI real de Companion | parsing roto / campos vacíos | Comparar `internal/axis/types.go` contra `axis/companion/openapi.yaml` y `nexus_types.go` contra `axis/nexus/openapi.yaml` |

## Datos / migración

- **Sin migraciones**. Único acceso a DB: `SELECT 1 FROM projects WHERE id=? AND tenant_id=? AND deleted_at IS NULL` (tenant-scope). Riesgo: si la tabla `projects` o columnas `tenant_id`/`deleted_at` no existen en develop, falla. Mitigación: confirmar el schema de `projects` en develop (depende de feature-010 projects).

## Archivos compartidos

| archivo | riesgo | mitigación |
|---|---|---|
| `wire/ai_providers.go` (feature-023) | sobrescribir o dejar la versión legacy rompe el wiring de Companion | Reescribir explícitamente; no traer ciegamente desde SOURCE ni dejar el de develop |
| `wire/wire_gen.go` (feature-023) | generado; merges manuales corrompen | Regenerar siempre |
| `cmd/config/loadconfig.go` (feature-005) | conflictos si 005 evoluciona | Coordinar orden; usar `git restore -p` para hunks puntuales |

## Extracción parcial

- Traer `internal/axis` y el adapter SIN reescribir `wire/ai_providers.go` → compila a medias o falla (símbolos no usados / `ai.Client` aún referenciado). Señal: `grep -rn "ProvideAIClient\|config.AI\b\|X-SERVICE-KEY"`.
- Olvidar borrar `internal/ai/client.go` → coexisten dos clientes; el legacy queda muerto pero referenciado por el `ai_providers.go` de develop. Señal: `git grep -n "ai.NewClient"`.
- Traer handler/usecases nuevos sin `shared/authz` (feature-001) → no compila.

## Riesgo de mergear solo un repo

- **Solo BE**: el contrato hacia el FE cambia poco (retrocompat salvo streaming). El FE legacy seguirá funcionando con `done`, pero perderá el streaming progresivo si lo esperaba. Riesgo MEDIO-BAJO si el shape se respeta.
- **Solo FE**: el FE apuntaría a endpoints cuyo BE aún usa ponti-ai → comportamiento inconsistente. Riesgo ALTO. NO mergear FE antes que BE.
