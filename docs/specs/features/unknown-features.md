# Unknown Features Baseline Specification

Specification type: baseline current-state uncertainty catalog.

## UNKNOWN Feature Areas

| Area | UNKNOWN |
|---|---|
| Frontend/BFF features | Source code unavailable in this repository |
| Mobile application | No mobile app code verified |
| External AI internals | Ponti AI implementation and storage unavailable |
| Review/Nexus policy behavior | Governance service policy definitions unavailable |
| Full tenant isolation for business data | Physical `tenant_id` absent/unverified on most business tables |
| Allocation percentage invariants | Complete percentage sum rules for project/field/admin cost investor allocations not verified |
| Monitoring/alerting features | No alert rules, dashboards, tracing, or metrics exporter verified |
| Queue/event-driven features | No runtime queue/event contracts verified |
| GraphQL/gRPC features | No runtime GraphQL/gRPC API verified |

## Evidence

- `internal/*`
- `migrations_v4/*`
- `go.mod`
- `.github/workflows/*`
- absence of verified runtime contracts during audit
