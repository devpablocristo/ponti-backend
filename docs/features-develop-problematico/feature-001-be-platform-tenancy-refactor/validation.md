# validation — feature-001 · Platform tenancy refactor (Fase 7)

## Checklist pre-PR (BE)

- [ ] `git grep -n "devpablocristo/core/" internal/` → 0 resultados en los paths de la flist.
- [ ] `git grep -n "MaybeTenantScope\|TenantWhere\|RequireJWT" internal/` → solo comentarios históricos (no código activo local).
- [ ] El diff del PR NO contiene: `lifecycle.RunCascadeArchive`, `lifecycle.ArchiveUpdates`, `lifecycle.RestoreScopedRows`, `lifecycle.Cause`, `actorsync.*`, `legacy_actor_map`, `EnsureCustomerFromActor`, `RefreshProjectActorMirrors`.
- [ ] El diff NO borra `internal/platform/files/excel/excelize/*` (eso es 013).
- [ ] El diff de `report`/`dashboard` (si entran) es solo import-swap + ctx threading, sin `fmt.Errorf`→domainerr ni `project_id IN ?`.
- [ ] `internal/shared/authz/authz.go` exporta: `Principal`, `PrincipalFromContext`, `RequireTenant`, `TenantFromContext`, `OptionalTenantOrStrict`, `TenantStrictModeEnabled`, `HasPermission`, `RequirePermission`, y las constantes `Permission*`.

## Tests sugeridos (BE)

```bash
go build ./...
go vet ./...
go test ./internal/shared/authz/...
go test ./internal/shared/filters/...
go test ./internal/platform/http/middlewares/gin/...
# y los repos tocados (si tienen suite):
go test ./internal/customer/... ./internal/lot/... ./internal/field/... ./internal/work-order/...
```

Tests que deben pasar (presentes en dp~1):
- authz: `TestTenantOwnerIsNotGlobalWildcard`, `TestSaaSSuperadminIsGlobalWildcard`, `TestTenantFromContextAllowsTransitionMode`.
- filters: `TestResolveProjectIDsScopesByTenant`.
- mw: `TestValidateIdentityClaimsRejectsIssuerMismatch`, `TestValidateIdentityClaimsRejectsAudienceMismatch`, `TestValidateIdentityClaimsAcceptsIssuerAndAudience`, `TestRejectUnsafeLocalAuthzBlocksProductionEnv`.

## Validación manual

### API
- [ ] Con header de tenant válido (`cfg.Auth.TenantHeader`): listados (customers, lots, fields, workorders) devuelven solo datos del tenant.
- [ ] Sin Auth.Enabled en entorno local-like: requests pasan (RequireLocalDevAuthz).
- [ ] Sin Auth.Enabled en entorno NO local-like: requests rechazados (RejectUnsafeLocalAuthz).
- [ ] Token con issuer/audience incorrecto → 401/403.

### DB
- [ ] Las queries de los repos incluyen `WHERE <alias>.tenant_id = ?` cuando hay tenant en contexto (revisar SQL logs).
- [ ] Sin tenant en contexto y strict-mode off: queries sin filtro tenant (compat).

### Env / config
- [ ] Confirmar `Auth.Enabled`, `Auth.Environment`, `Auth.TenantHeader`, issuer/audience por entorno.
- [ ] `go.mod` resuelve `platform/*` (no `core/*`).

### Logs / observabilidad
- [ ] Access log JSON con `request_id`; header `X-Request-Id` propagado.
- [ ] Headers `Authorization`/`Cookie` aparecen `<redacted>` en logs.
- [ ] `godotenv`/`gorm`/`server` loguean por slog (no fmt.Printf/log.*).

## Casos borde

- Contexto sin tenant + strict-mode ON → `domainerr.TenantMissing()` (cierra). Verificar `platform/security/go/tenant.StrictModeEnabled()`.
- `ResolveProjectIDs` con 0 proyectos en modo tenant → devuelve `[]int64{0}` (centinela, no nil) → consumidores deben tratar `{0}` como "ningún proyecto".
- `ResolveProjectIDs` sin filtros de negocio en modo tenant → todos los proyectos del tenant (no DB completa).
- Rol `saas_superadmin` → wildcard de permisos (HasPermission true para cualquier permiso).

## Qué validar en el otro repo (FE)

- Nada. Solo-BE. Confirmar que el `cross-repo-map` del FE marca feature-001 como "sin cambios FE".

## Señales de incompletitud / incompatibilidad

- Build falla por `undefined: lifecycle.*`/`actorsync.*` → faltan símbolos de 002/007 (cortaste de menos) o trajiste imports que no van.
- `goimports`/`vet` reporta import sin usar → swap incompleto.
- Listados vacíos inesperados → semántica `[]int64{0}` mal manejada aguas arriba, o tenant ausente en contexto.
- Quedó `devpablocristo/core/` en algún archivo → swap incompleto.
