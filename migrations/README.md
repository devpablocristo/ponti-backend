# Migrations baseline (v4 only)

## Strategy
- This baseline keeps **v4_report** as the primary schema.
- Only v4 schemas and views are included.

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
11. v4 report views extensions

## Notes
- SQL is in English; comments in Spanish when needed.
- All migrations are transactional (`BEGIN/COMMIT`).
- No historical migrations are preserved; this is a clean baseline.
- Required extension `unaccent` is declared explicitly.
