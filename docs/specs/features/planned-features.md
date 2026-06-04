# Planned Features Baseline Specification

Specification type: baseline current-state planned/historical feature catalog.

This file records planned or historical feature artifacts discovered in `docs/features-develop-problematico/`. Current verified behavior belongs in `feature-inventory.md` and the status-specific current-state catalogs, not in this file.

| Historical Feature | Baseline Treatment | Evidence |
|---|---|---|
| Platform tenancy refactor | Historical planned scope remains unresolved for business data; current auth context behavior is catalogued as PL-02 and PL-03 | `docs/features-develop-problematico/feature-001-*`; current business tables do not verify physical tenant isolation |
| CRU/DAR lifecycle framework | Historical planned scope remains unverified | `internal/shared/lifecycle` not verified |
| Multitenant DB hardening | Historical planned scope remains unverified | referenced migrations not verified in current migration set |
| Proper-name normalization | Historical planned scope remains unverified | `internal/shared/text` not verified |
| Config modularization | Historical planned scope remains broader than verified current config loading | current `cmd/config/*` exists; planned companion/nexus/reporting/security shape not fully verified |
| FE design system | Historical planned scope is external/UNKNOWN | frontend code unavailable |
| Actor system | Historical planned scope remains unverified | `internal/actor` not verified |
| Identity tenant context `/me` | Historical planned scope remains unverified | explicit `/me` context endpoint not verified |
| Shared archive surface | Historical planned shared lifecycle abstraction remains unverified; per-domain archive routes are catalogued in current feature specs | archive routes exist per domain; shared lifecycle abstraction not verified |
| Project actor/lifecycle version | Historical planned actor/shared lifecycle scope remains unverified | project CRUD exists; actor/shared lifecycle version not verified |
| Campaign DTO project_id contract | Historical planned DTO contract remains unresolved | current campaign list exists; planned DTO contract not fully verified |
| AI companion integration | Historical planned companion/nexus contract remains unresolved; current backend AI proxy behavior is catalogued as AI-01 through AI-03 | backend AI proxy exists; companion/nexus internal contract not verified |
| CSV export framework | Historical planned scope remains unverified; current exports are XLSX where verified | current exports verified as XLSX, not CSV framework |
| Data integrity admin surface | Historical planned admin/frontend surface remains external/UNKNOWN | backend endpoints exist; admin/frontend surface external/UNKNOWN |
| Lefthook git hooks | Historical planned scope remains unverified | `lefthook.yml` not verified |
| OpenAPI and docs | Historical planned scope remains unresolved; current generation target is catalogued as PL-06 | Make target exists; generated OpenAPI output UNKNOWN |
| Test coverage expansion | Historical planned coverage expansion remains unresolved | tests exist and pass; planned broader coverage not fully verified |
| Domain purity cleanup | Historical planned cleanup remains unresolved | historical docs identify cleanup; current code still includes mixed tags in domain/report shapes |

## Evidence

- `docs/features-develop-problematico/index.md`
- `docs/features-develop-problematico/feature-*`
- current source tree checks performed during audit
