# Stubbed Features Baseline Specification

Specification type: baseline current-state feature status specification.

| Feature ID | Feature | Stub Evidence |
|---|---|---|
| OP-06 | Work order duplicate | `POST /api/v1/work-orders/:work_order_id/duplicate` route exists; `DuplicateWorkOrder` usecase returns empty string and nil error without duplicating persistence |

## Evidence

- `internal/work-order/handler.go`
- `internal/work-order/usecases.go`

## Required Status Treatment

This feature must remain classified as `Stubbed` until current code verifies real duplication behavior.
