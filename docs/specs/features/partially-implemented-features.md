# Partially Implemented Features Baseline Specification

Specification type: baseline current-state feature status specification.

| Feature ID | Feature | Verified Current State | Verified Gap / UNKNOWN |
|---|---|---|---|
| PL-06 | OpenAPI generation target | Make target/reference exists | Generated OpenAPI artifact not verified |
| PF-02 | Campaign lookup | `GET /campaigns` exists | Create/get route surface not verified despite usecase/repository methods |
| FN-04 | Investor allocations | Allocation tables and project repository behavior exist | Dedicated API semantics and percentage invariants UNKNOWN |
| AI-01 | AI chat proxy | Backend proxy and dummy fallback exist | External AI implementation and persistence UNKNOWN |
| AI-02 | AI chat streaming proxy | Backend SSE proxy and fallback exist | External AI stream behavior UNKNOWN |
| AI-03 | AI conversations proxy | Backend proxy exists | Local conversation persistence UNKNOWN |

## Evidence

- `Makefile`
- `internal/campaign/handler.go`
- `internal/campaign/usecases.go`
- `internal/project/repository.go`
- `internal/ai/*`
- `migrations_v4/000070_investors_commercialization_tables.up.sql`
