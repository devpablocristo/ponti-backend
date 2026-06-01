# feature-018 · data-integrity-admin · risks (BE)

## Riesgos funcionales

- **Cambio de contrato hacia el FE**: el DTO suma `check_type`, `severity`, `recommendation` y
  reduce de 9/17 a 5 checks. Si el FE asume cantidad/forma fija, puede romper render.
  - *Mitigación*: coordinar BE-first; verificar en el paquete FE que `useDatabase` mapea por
    `control_number` y tolera campos extra y `recommendation` vacío.
- **Semántica de severidad incompleta**: `buildCheck` solo emite `Status ∈ {OK, ERROR}` y
  `Severity ∈ {INFO, ERROR}`, pero el dominio documenta `WARNING/SKIPPED`. Un FE que ramifique
  por `WARNING` nunca lo verá.
  - *Mitigación*: documentar el set real emitido; no exponer `WARNING/SKIPPED` en la UI hasta que el BE los produzca.

## Riesgos técnicos

- **No compila en aislamiento (ALTO)**: `usecases.go` referencia `GetRawNetIncome /
  GetRawSupplyInvestment / GetRawAdminCostTotal / GetRawLeaseExecuted`, inexistentes en develop;
  y los RAW methods usan `internal/shared/authz`, también ausente.
  - *Mitigación*: traer authz (001/003) primero; partial-hunks de los 4 RAW en el mismo PR; `go build ./...` como gate.
- **Wire desincronizado (ALTO)**: la firma de `ProvideDataIntegrityUseCases` cambió (6 args,
  sin `Stock`, con `Project`, repos concretos). Copiar `wire_gen.go` a mano o dejar providers
  viejos rompe `Initialize()`.
  - *Mitigación*: traer `wire/data_integrity_providers.go` por hunks y **regenerar** `wire_gen.go` con `go generate ./wire/...`; verificar idempotencia.
- **Import paths core→platform**: `handler.go` cambia `core/errors/go/domainerr` →
  `platform/errors/go/domainerr`. En develop ya predomina platform, así que OK; pero si quedara
  algún `core/...` colgando habría doble import.
  - *Mitigación*: `grep -rn "devpablocristo/core/errors" internal/data-integrity` debe dar 0.

## Riesgos de integración

- **Dependencia con dashboard SSOT (`v4_*`)**: los `SystemValue` salen de
  `dashboard.GetDashboard`. Si las vistas `v4_report/v4_ssot/v4_calc` no están desplegadas en el
  entorno, el control no es comparable (RAW ≠ SSOT por datos faltantes, no por bug).
  - *Mitigación*: validar en un entorno con las vistas v4 presentes y datos reales.

## Riesgos cross-repo

- **Merge desfasado FE/BE**: si FE 018 mergea antes que BE 018, el FE espera campos que el BE
  viejo no envía → posibles `undefined` en la UI.
  - *Mitigación*: orden BE-first; feature flag o tolerancia de campos opcionales en el FE.

## Riesgos de datos / migración

- **Sin migraciones** en 018-BE → riesgo de datos = bajo. Las queries RAW son SELECT puros.
- **Filtro tenant inconsistente**: 4 RAW filtran por `tenant_id` (modo estricto opcional);
  `GetRawDirectCost` en develop NO. Resultado: el control 1 podría sumar datos de otros tenants
  si no se tenantiza, dando falsos ERROR/OK.
  - *Mitigación*: tenantizar `GetRawDirectCost` (traer su hunk) o documentar la limitación.

## Riesgos de archivos compartidos (partial-hunks)

- **`internal/project/repository.go`** es enorme (+1642/-431 en rango). Tomar el hunk equivocado
  arrastra cambios de 001/003/010 que no corresponden a 018.
  - *Mitigación*: en `restore -p`, aceptar SOLO el bloque `func (r *Repository) GetRawAdminCostTotal`. Revisar `git diff --check` y `git diff HEAD` antes de commit.
- **Conflictos de imports**: los hunks RAW pueden requerir `authz`, `domainerr`, `fmt`,
  `decimal` en el bloque de imports; si el hunk no los trae, el archivo no compila.
  - *Mitigación*: tras pegar hunks, `goimports`/`go build` por paquete.

## Riesgos de extracción parcial

- **Falsa sensación de completitud**: `go test ./internal/data-integrity/...` puede pasar (usa
  mocks) aunque el repo entero no compile.
  - *Mitigación*: gate obligatorio = `go build ./...` del repo completo, no solo el test del paquete.

## Riesgo de mergear SOLO un repo

- **Solo BE**: endpoint nuevo funcional, pero la UI admin del FE sigue vieja → el usuario no ve
  `severity/recommendation`. Aceptable temporalmente.
- **Solo FE**: la UI espera campos que el BE viejo no entrega → la página admin puede mostrar
  vacíos o romper. **Peor escenario**; evitar FE-first.
