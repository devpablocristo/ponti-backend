# dependencies — feature-001 · Platform tenancy refactor (Fase 7)

## Resumen de posición

feature-001 es la **base del BE**. No depende de otra feature de la lista para existir, pero sí del **ecosistema de módulos `platform/*`** (new-cns3) en `go.mod`. Bloquea (directa o indirectamente) a casi todas las features BE/BEFE posteriores.

## Depende de

### Plataforma (FUERTE, externa a la lista)
- `github.com/devpablocristo/platform/*` debe resolver en `go.mod`/`go.sum`: `errors/go/domainerr`, `security/go/contextkeys`, `security/go/tenant`, `http/gin/go`, `http/go/httperr`, `databases/postgres/go`, `authn/go/jwks`, `observability/go`, `persistence/gorm/go/tenancy`.
- Bumps go-jose/x/net: YA en develop (#124). Excluir de 021.

### Intra-repo (DÉBIL / contextual)
- **003 be-multitenant-db-hardening** (DÉBIL→FUERTE en runtime): el scoping `tenancy.Scope(ctx, db, alias)` filtra por columna `tenant_id`. Si esas columnas no existen, en modo transición no rompe build pero no acota nada. Para que el feature tenga efecto real necesita 003.
- **008 identity-tenant-context** (DÉBIL→FUERTE en runtime): `TenantFromContext` lee `contextkeys.OrgID` que lo setea el middleware de identidad/tenant de 008. Sin 008, el contexto nunca trae tenant (modo transición permanente).

### INCIERTAS (mezcla de código, NO dependencia lógica)
- En dp~1 los `repository.go` de customer/lot/etc. importan `internal/shared/lifecycle` (002/009) e `internal/actor` (007). Eso es **contaminación del diff**, no una dependencia de 001. Al extraer 001 hay que EXCLUIR esos hunks; si no se puede, el repo va con su feature dueña.

## Bloquea a

- **002 be-crudar-lifecycle-framework** (FUERTE): usa `authz`/`tenancy.Scope` y los repos ya migrados.
- **003 be-multitenant-db-hardening** (FUERTE): el código que consume `tenant_id` vive aquí.
- **007 actor-system** (FUERTE): los repos donde se inyecta actor-sync ya deben estar migrados a platform.
- **008 identity-tenant-context** (FUERTE): produce el tenant que `authz.TenantFromContext` consume.
- **009 crudar-archive-surface** (FUERTE): la cascada de archive opera sobre repos ya migrados.
- **004, 005, 013, 023, 025, 027** y demás features BE (DÉBIL): comparten imports `platform/*`; si 001 no fue antes, cada una tendría que migrar imports por su cuenta → conflictos.

## Archivos / tipos / config / APIs compartidos

| recurso | feature dueña | compartido con | nota |
|---|---|---|---|
| internal/shared/authz/* | 001 | 002,007,008,009 | núcleo de permisos/tenant |
| internal/shared/filters/workspace.go | 001 | 010 (projects), 013 (export) | ResolveProjectIDs/ValidateProjectAccess |
| internal/shared/models/base.go | 008/007 | 001 | en 001 solo el hunk de import contextkeys |
| internal/*/repository.go | 001 (import+scope) | 002,007,009 (lógica) | partial-hunks obligatorio |
| internal/report/repository.go | 027/013 | 001 (import+ctx) | el grueso del diff NO es de 001 |
| internal/dashboard/repository.go | 027/013 | 001 (import+ctx) | idem |
| internal/platform/files/excel/excelize/* | 013 | — | borrado NO es de 001 |
| go.mod / go.sum | 021/plataforma | todas | resolución platform/* |
| contextkeys.OrgID/Actor/Role/Scopes | 008 | 001 | claves de contexto |

## APIs / contratos

- Ninguno cambia (refactor interno). Sin impacto FE.

## Migraciones compartidas

- Ninguna en esta flist. Las columnas `tenant_id` vienen de 003.

## Recomendación de orden

1. **001** (este) — primero de todo el BE, tras tener `platform/*` en go.mod.
2. **003** + **008** — habilitan el efecto runtime del scoping y el tenant en contexto.
3. **002** / **007** / **009** — sobre repos ya migrados.
4. **004, 005, 013, 023, 025, 027** — el resto del BE.

Cross-repo: ninguno. El FE no se toca por 001.
