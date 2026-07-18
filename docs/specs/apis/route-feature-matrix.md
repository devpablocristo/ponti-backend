# Route Feature Matrix Baseline Specification

Specification type: baseline current-state API mapping specification.

| Feature IDs | Route Group |
|---|---|
| PL-01 | `/version`, `/healthz`, `/ping` |
| PL-02 | Global validation middleware |
| PL-03 | `/admin/*` |
| PL-08 | `/actors*` (identity actor registry / dedup, Identity Gate) |
| PL-11 | `/registry*` (unified entity search + actor aliases) |
| PL-09 | Tenant enforcement (cross-cutting; no dedicated REST route — flag-gated scoping on existing routes) |
| PL-10 | Fine-grained permissions / authz dual-check (cross-cutting; no dedicated REST route) |
| PL-04 | No REST route; migration command/runtime artifact |
| PL-05 | No REST route; CI/CD workflow artifact |
| PL-06 | No REST route verified; OpenAPI generation target only |
| PL-07 | No REST route; local development artifact |
| PF-01 | `/customers*` |
| PF-02 | `/campaigns*` (List accepts `?status=active|archived|all`; `?status=archived` routes to `/campaigns/archived`) |
| PF-03, PF-04 | `/projects*` |
| PF-05 | `/managers*` |
| PF-06 | `/investors*` |
| PF-07 | `/providers` |
| PF-08 | `/business-parameters*` |
| PF-09 | `/categories*` |
| PF-10 | `/types*` (List accepts `?status=active|archived|all`) |
| LC-01 | `/fields*` |
| LC-02 | `/lease-types*` (List accepts `?status=active|archived|all`) |
| LC-03, LC-04 | `/lots*` |
| LC-05 | `/crops*` (List accepts `?status=active|archived|all`) |
| OP-01, OP-02 | `/projects/:project_id/labors*`, `/labors*` |
| OP-03, OP-04, OP-05, OP-06 | `/work-orders*`; OP-03 includes `/work-orders/archived` |
| OP-07, OP-08, OP-09, OP-10 | `/work-order-drafts*` |
| IN-01, IN-02, IN-03, IN-04 | `/supplies*` |
| IN-05, IN-06, IN-07, IN-08 | `/projects/:project_id/supply-movements*`, `/projects/:project_id/stock-movements*` |
| IN-09, IN-10 | `/projects/:project_id/stocks*` |
| FN-01 | `/projects/:project_id/dollar-values` |
| FN-02 | `/projects/:project_id/commercializations` |
| FN-03 | `/invoices*` |
| FN-04 | Dedicated allocation route UNKNOWN |
| RP-01 | `/dashboard` |
| RP-02, RP-03, RP-04 | `/reports/:type` |
| RP-05, RP-06 | `/data-integrity/*` |
| AI-01, AI-02, AI-03 | `/ai/chat*` |
| AI-04, AI-05 | `/insights*` |

## Evidence

- `internal/*/handler.go`
- `docs/specs/features/feature-inventory.md`
