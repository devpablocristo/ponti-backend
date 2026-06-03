# Plans store

Planes/diseños que se planificaron pero se dejaron para implementar más adelante.

**Convención de naming:** `NNNN-<slug>.md` (número incremental de 4 dígitos). Cada archivo lleva un
**Estado** al inicio: `Draft` · `Approved` · `Deferred` · `Implemented` · `Rejected`.

| # | Plan | Estado | Resumen |
|---|------|--------|---------|
| 0001 | [github-flow-previews-cloudrun.md](0001-github-flow-previews-cloudrun.md) | Deferred | Migración a GitHub Flow + Preview Environments por PR en Cloud Run, con instancia Cloud SQL dedicada para previews |
| 0002 | [workorder-metrics-digital-and-filters.md](0002-workorder-metrics-digital-and-filters.md) | Draft | KPIs de Órdenes de Trabajo: reflejar borradores digitales y respetar `is_digital`/`status` vía vista dedicada (migración), sin tocar la SSOT de costo de data-integrity |
