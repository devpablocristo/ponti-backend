# Entity Capabilities

This file records the current CRUDAR policy direction. It is intentionally not
a generic-table design; each entity keeps its own usecase/repository.

| Entity | Category | Decision | Notes |
| --- | --- | --- | --- |
| actors | identity master | APPLY CRUDAR + merge | `deleted_at` is lifecycle source; `archived_at` remains compatibility. |
| customers | operational root | APPLY CRUDAR | Cascades archive to active projects by batch cause. |
| projects | operational child | APPLY CRUDAR | Belongs to customer; restore requires active customer. |
| fields | operational child | APPLY CRUDAR | Archive cascades to active lots by batch cause; restore only restores lots from the same cause. |
| lots | operational child | APPLY CRUDAR | Restore requires active field; hard delete blocks existing work orders. |
| campaigns | operational child/catalog | APPLY CRUDAR | Archive/restore use lifecycle metadata; hard delete blocks project references. |
| work_orders | operation | APPLY CRUDAR | Archive cascades to owned items/splits by cause; restore only restores owned rows from the same cause. |
| labors | operation/catalog | APPLY CRUDAR | Archive/restore use lifecycle metadata; hard delete blocks work order usage. |
| supplies | catalog/operation | APPLY CRUDAR | Archive/restore use lifecycle metadata; hard delete blocks movements/stocks/work order usage. |
| supply_movements | operation | APPLY CRUDAR | Project-scoped; restore requires active project. |
| managers | operational actor role | APPLY CRUDAR | Legacy table remains actor-linked and syncs actor lifecycle. |
| investors | operational actor role | APPLY CRUDAR | Legacy table remains actor-linked and syncs actor lifecycle. |
| providers | operational actor role | APPLY CRUDAR | Backend CRUDAR completed; hard delete blocks supply movement usage. No dedicated frontend CRUD page exists yet. |
| stock | operational state | PARTIAL CRUDAR | Stock has updates and close-period behavior, not full CRUDAR today. |
| invoices | operational document | PARTIAL CRUDAR | Delete exists; archive/restore policy needs explicit design. |
| categories | mutable catalog | APPLY CRUDAR | Archive/restore use lifecycle metadata; hard delete blocks supply/labor category usage. |
| crops | mutable catalog | APPLY CRUDAR | Archive/restore use lifecycle metadata; hard delete blocks lot/work order/commercialization usage. |
| class_types | mutable catalog | APPLY CRUDAR | Archive/restore use lifecycle metadata; hard delete blocks category/supply type usage. |
| lease_types | mutable catalog | APPLY CRUDAR | Archive/restore use lifecycle metadata; hard delete blocks field usage. |
| business_parameters | mutable catalog | APPLY CRUDAR | Archive/restore use lifecycle metadata; no FK usage found, so hard delete requires archived only. |
| join tables without business identity | relation | RELATION_ONLY_DELETE | Archive only when history is required. |
| audit logs / insight events / migration tables | append-only/technical | NO CRUDAR | Do not hard delete through normal entity UI. |
