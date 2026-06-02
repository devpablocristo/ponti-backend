# implementation-status.md — feature-024 · openapi-and-docs (BE)

## Estado global

**COMPLETA** como documentación. Los 20 archivos del flist existen y son coherentes en `777e5f6a`. El único "incompleto" intencional es el spec OpenAPI (piloto de 2 endpoints, declarado como tal).

- **% completitud (como paquete de docs)**: ~95%.
  - 17 docs nuevos: 100% presentes y legibles.
  - 3 modificados: 100% presentes (con caveat de overlap de hunks con 019/021).
  - OpenAPI spec: ~4% de cobertura de endpoints (2 de ~50), pero ESTO ES EL ESTADO DECLARADO, no un bug del paquete.

## Estado en este repo (BE)

- Todo el contenido existe en el SOURCE. Verificado con `git show 777e5f6a:<path>`.
- Los 17 nuevos NO existen en develop → copia limpia, sin conflicto.
- Los 3 modificados existen en develop con contenido viejo → requieren merge/partial.
- `docs/openapi/docs.go` (output `.go` de swag) NO está presente ni en SOURCE ni en flist (correcto: se regenera).

## Estado en el otro repo (FE feature-024)

- No verificado desde aquí (paquete separado). Según la NOTA de feature: FE aporta `docs/`, `docs/audit` (visual regression, posible generado), `RESPONSIVE_GUIDELINES`, `PR-92.md`. El FE consume `swagger.yaml` con `yarn codegen:openapi`. Coordinar con el agente FE-024.

## Tests

- Ninguno (docs). No hay nada que correr.
- Validación = lectura/render de markdown + comparación byte a byte con SOURCE.

## Pendientes (del propio contenido, declarados en los docs)

- **`docs/OPENAPI.md`**: anotar los ~48 handlers restantes con comentarios swaggo y regenerar spec. (Backlog, NO de 024.)
- **`CRUDAR_PLAN.md`**: fases B/C/E/F/G parciales o pendientes (es plan FE; informativo).
- **`docs/projects-archive-audit.md`**: deuda BE conocida — `RunCascadeRestore` no es base segura todavía; data cleanup histórico pendiente (IA-4..IA-14). NO es de 024 arreglarlo.

## Bugs

- Ninguno funcional (docs).
- **Riesgo cosmético**: enlaces relativos y referencias a line numbers (`audit-custom-errors.md` cita `internal/report/repository.go:45`, etc.) pueden haber drifteado respecto al código actual de develop. No rompe nada; sólo puede confundir.

## Clasificación de pendientes

### BLOQUEANTE para mergear
- Ninguno. El paquete puede mergear tal cual.
- (Condicional) Resolver el overlap de hunks de README/docs.README con 019/021 ANTES de incluir esos archivos, para no generar conflicto/doble fuente. Si se omiten esos 2 archivos del PR, deja de ser bloqueante.

### Mejora futura
- Completar anotaciones swaggo (cobertura OpenAPI de ~50 endpoints).
- Actualizar conteos/fechas en `MULTI_TENANT_100_EVIDENCE.md` y `BACKEND_CLEANUP_AUDIT.md` para reflejar develop actual.

### Deuda aceptable
- Snapshots históricos (`*_AUDIT.md`, `*_EVIDENCE.md`) con datos de 2026-05-12/05-26. Aceptable como registro de auditoría fechado.
- Spec OpenAPI piloto. Aceptable porque está declarado.
- `CRUDAR_PLAN.md` FE-céntrico dentro del repo BE.

### Duda humana
- ¿README.md/docs.README.md van en 024 o en 019/021? (overlap de tooling rename).
- ¿`CRUDAR_PLAN.md` debería vivir en BE o FE? (el flist lo pone en BE).
- ¿Mantener los snapshots de auditoría con datos viejos o actualizarlos? (decisión de producto/docs).
