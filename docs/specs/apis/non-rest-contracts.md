# Non-REST Contracts Baseline Specification

Specification type: baseline current-state API contract specification.

## GraphQL

Status: UNKNOWN / not verified in current runtime.

No GraphQL server route or schema was verified in the audited backend.

## gRPC

Status: UNKNOWN / not verified in current runtime.

No gRPC service registration or `.proto` contract was verified as current backend runtime API.

## Events And Queues

Status: UNKNOWN / not verified in current runtime.

No queue producer, queue consumer, topic contract, or broker runtime configuration was verified.

## Scheduled Jobs

Status: UNKNOWN.

GitHub workflows exist for operational tasks, but no in-app scheduler contract was verified.

## Evidence

- `cmd/api/http_server.go`
- `internal/*/handler.go`
- `go.mod`
- `.github/workflows/*`

## Rule

Do not model GraphQL, gRPC, queue, event, or scheduled-job behavior as implemented until current code/config verifies it.
