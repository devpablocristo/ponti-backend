# Migrations baseline (Strategy B: dual-support)

## Strategy
- This baseline keeps **v4_report** as the primary schema when `REPORT_SCHEMA=v4_report`.
- It also provides a **minimal v3 fallback** set of views in `public`.
- Required v3 function schemas (`v3_calc`, `v3_core_ssot`, `v3_lot_ssot`, `v3_workorder_ssot`, `v3_dashboard_ssot`, `v3_report_ssot`) are included only to support those fallback views.

## Order of migrations
1. Extensions
2. Base tables
3. Core triggers/functions
4. Domain tables
5. Constraints/FKs/Indexes
6. v4 schemas
7. v4 core functions
8. v4 ssot functions
9. v4 calc views
10. v4 report views
11. v3 minimal fallback views

## Notes
- SQL is in English; comments in Spanish when needed.
- All migrations are transactional (`BEGIN/COMMIT`).
- No historical migrations are preserved; this is a clean baseline.
- Required extension `unaccent` is declared explicitly.
