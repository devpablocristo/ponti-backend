# REST API Inventory Baseline Specification

Specification type: baseline current-state REST API specification.

Base path: `/api/v1`

## Platform, Identity, And Admin

- `GET /version`
- `GET /healthz`
- `GET /ping`

- `GET /admin/tenants`
- `POST /admin/tenants`
- `GET /admin/users`
- `POST /admin/users`
- `POST /admin/memberships`

## Portfolio And Master Data

- `/customers`: `POST`, `GET`
- `/customers/archived`: `GET`
- `/customers/:customer_id`: `GET`, `PUT`, `DELETE`
- `/customers/:customer_id/archive`: `POST`
- `/customers/:customer_id/restore`: `POST`
- `/campaigns`: `GET`
- `/projects`: `POST`, `GET`
- `/projects/archived`: `GET`
- `/projects/:project_id/fields`: `GET`
- `/projects/dropdown`: `GET`
- `/projects/customer/:customer_id`: `GET`
- `/projects/customers/:customer_id`: `GET`
- `/projects/:project_id`: `GET`, `PUT`, `DELETE`
- `/projects/:project_id/archive`: `POST`
- `/projects/:project_id/restore`: `POST`
- `/projects/search`: `GET`
- `/managers`: `POST`, `GET`
- `/managers/:manager_id`: `GET`, `PUT`, `DELETE`
- `/managers/:manager_id/archive`: `POST`
- `/managers/:manager_id/restore`: `POST`
- `/investors`: `POST`, `GET`
- `/investors/:investor_id`: `GET`, `PUT`, `DELETE`
- `/investors/:investor_id/archive`: `POST`
- `/investors/:investor_id/restore`: `POST`
- `/providers`: `GET`
- `/business-parameters`: `GET`, `POST`
- `/business-parameters/category/:category`: `GET`
- `/business-parameters/:parameter_key`: `GET`
- `/business-parameters/:parameter_id`: `PUT`, `DELETE`
- `/categories`: `POST`, `GET`
- `/categories/:category_id`: `GET`, `PUT`, `DELETE`
- `/types`: `POST`, `GET`
- `/types/:class_type_id`: `GET`, `PUT`, `DELETE`

## Land And Crops

- `/fields`: `POST`, `GET`
- `/fields/:field_id`: `GET`, `PUT`, `DELETE`
- `/fields/:field_id/archive`: `POST`
- `/fields/:field_id/restore`: `POST`
- `/lease-types`: `POST`, `GET`
- `/lease-types/:lease_type_id`: `GET`, `PUT`, `DELETE`
- `/lots`: `POST`, `GET`
- `/lots/metrics`: `GET`
- `/lots/:lot_id/tons`: `PUT`
- `/lots/:lot_id`: `GET`, `PUT`, `DELETE`
- `/lots/export`: `GET`
- `/crops`: `POST`, `GET`
- `/crops/:crop_id`: `GET`, `PUT`, `DELETE`

## Field Operations

- `/projects/:project_id/labors`: `POST`, `GET`
- `/projects/:project_id/labors/:labor_id`: `DELETE`, `PUT`
- `/projects/:project_id/labors/:labor_id/workorders-count`: `GET`
- `/projects/:project_id/labors/labor-categories/:type_id`: `GET`
- `/projects/:project_id/labors/export`: `GET`
- `/labors/:labor_id`: `DELETE`
- `/labors/:work_order_id`: `GET`
- `/labors/group/:project_id`: `GET`
- `/labors/export/:project_id`: `GET`
- `/labors/export/all`: `GET`
- `/labors/metrics`: `GET`
- `/work-orders`: `POST`, `GET`
- `/work-orders/filter-rows`: `GET`
- `/work-orders/metrics`: `GET`
- `/work-orders/export`: `GET`
- `/work-orders/:work_order_id`: `GET`, `PUT`, `DELETE`
- `/work-orders/:work_order_id/archive`: `POST`
- `/work-orders/:work_order_id/restore`: `POST`
- `/work-orders/:work_order_id/investors/:investor_id/payment-status`: `PATCH`
- `/work-orders/:work_order_id/duplicate`: `POST`
- `/work-order-drafts`: `POST`, `GET`
- `/work-order-drafts/digital`: `POST`, `GET`
- `/work-order-drafts/digital/batch`: `POST`
- `/work-order-drafts/digital/preview-number`: `POST`
- `/work-order-drafts/digital/batch/preview-number`: `POST`
- `/work-order-drafts/digital/groups`: `GET`
- `/work-order-drafts/:work_order_draft_id/group`: `GET`, `PUT`
- `/work-order-drafts/:work_order_draft_id`: `GET`, `PUT`, `DELETE`
- `/work-order-drafts/:work_order_draft_id/pdf-data`: `GET`
- `/work-order-drafts/:work_order_draft_id/group-pdf-data`: `GET`
- `/work-order-drafts/:work_order_draft_id/publish`: `POST`

## Inventory And Stock

- `/supplies`: `POST`, `GET`
- `/supplies/pending`: `POST`, `GET`
- `/supplies/bulk`: `POST`, `PUT`
- `/supplies/export/all`: `GET`
- `/supplies/:supply_id`: `GET`, `PUT`, `DELETE`
- `/supplies/pending/:supply_id/complete`: `PUT`
- `/supplies/:supply_id/archive`: `POST`
- `/supplies/:supply_id/restore`: `POST`
- `/supplies/:supply_id/workorders-count`: `GET`
- `/projects/:project_id/supply-movements`: `POST`, `GET`
- `/projects/:project_id/supply-movements/import`: `POST`
- `/projects/:project_id/supply-movements/export`: `GET`
- `/projects/:project_id/supply-movements/providers`: `GET`
- `/projects/:project_id/supply-movements/:supply_movement_id`: `PUT`, `DELETE`
- `/projects/:project_id/stock-movements`: `POST`, `GET`
- `/projects/:project_id/stock-movements/export`: `GET`
- `/projects/:project_id/stock-movements/providers`: `GET`
- `/projects/:project_id/stock-movements/:stock_movement_id`: `PUT`, `DELETE`
- `/projects/:project_id/stocks/summary`: `GET`
- `/projects/:project_id/stocks/periods`: `GET`
- `/projects/:project_id/stocks/export`: `GET`
- `/projects/:project_id/stocks/close-date`: `PUT`
- `/projects/:project_id/stocks/real-stock/:stock_id`: `PUT`

## Finance And Investor Accounting

- `/projects/:project_id/dollar-values`: `GET`, `PUT`
- `/projects/:project_id/commercializations`: `GET`, `POST`
- `/invoices`: `GET`
- `/invoices/:work_order_id`: `GET`, `POST`, `PUT`, `DELETE`

## Reporting And Data Integrity

- `/dashboard`: `GET`
- `/reports/:type`: `GET`
- `/data-integrity/costs-check`: `GET`
- `/data-integrity/tentative-prices`: `GET`

## AI And Business Insights

- `/ai/chat`: `POST`
- `/ai/chat/stream`: `POST`
- `/ai/chat/conversations`: `GET`
- `/ai/chat/conversations/:conversation_id`: `GET`
- `/insights`: `GET`
- `/insights/:id/read`: `POST`, `DELETE`
- `/insights/:id/resolve`: `POST`, `DELETE`

## Evidence

- `internal/*/handler.go`
- `cmd/api/http_server.go`
