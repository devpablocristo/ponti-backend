# Domain Feature Matrix Baseline Specification

Specification type: baseline current-state mapping specification.

| Domain | Feature IDs |
|---|---|
| Platform, Identity, And Admin | PL-01, PL-02, PL-03 |
| Runtime, Migration, And Delivery | PL-04, PL-05, PL-06, PL-07 |
| Portfolio And Master Data | PF-01, PF-02, PF-03, PF-04, PF-05, PF-06, PF-07, PF-08, PF-09, PF-10 |
| Land And Crops | LC-01, LC-02, LC-03, LC-04, LC-05 |
| Field Operations | OP-01, OP-02, OP-03, OP-04, OP-05, OP-06, OP-07, OP-08, OP-09, OP-10 |
| Inventory And Stock | IN-01, IN-02, IN-03, IN-04, IN-05, IN-06, IN-07, IN-08, IN-09, IN-10 |
| Finance And Investor Accounting | FN-01, FN-02, FN-03, FN-04 |
| Reporting And Data Integrity | RP-01, RP-02, RP-03, RP-04, RP-05, RP-06 |
| AI And Business Insights | AI-01, AI-02, AI-03, AI-04, AI-05 |

## Evidence

- `cmd/api/http_server.go`
- `internal/*/handler.go`
- `internal/*/usecases*.go`
- `migrations_v4/*`
