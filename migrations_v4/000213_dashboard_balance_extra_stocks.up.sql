-- Agrega columnas labores_stock_usd, arriendo_stock_usd y estructura_stock_usd
-- a las views de Balance de Gestión, calculadas como invertido - ejecutado
-- (consistente con las otras categorías).

BEGIN;

DROP VIEW IF EXISTS v4_report.dashboard_management_balance;
DROP VIEW IF EXISTS v4_report.dashboard_management_balance_field;

CREATE VIEW v4_report.dashboard_management_balance AS
SELECT p.id AS project_id,
    COALESCE(sum(v4_ssot.income_net_total_for_lot(l.id)), 0::numeric) AS income_usd,
    v4_ssot.operating_result_total_for_project(p.id) AS operating_result_usd,
    v4_ssot.renta_pct(v4_ssot.operating_result_total_for_project(p.id), v4_ssot.total_costs_for_project(p.id)) AS operating_result_pct,
    v4_ssot.direct_costs_total_for_project(p.id) AS costos_directos_ejecutados_usd,
    v4_ssot.supply_movements_invested_total_for_project(p.id) + COALESCE(sum(v4_ssot.labor_cost_pre_harvest_for_lot(l.id)), 0::numeric) AS costos_directos_invertidos_usd,
    v4_ssot.supply_movements_invested_total_for_project(p.id) + COALESCE(sum(v4_ssot.labor_cost_pre_harvest_for_lot(l.id)), 0::numeric) - v4_ssot.direct_costs_total_for_project(p.id) AS costos_directos_stock_usd,
    COALESCE(sc.semillas_ejecutados_usd, 0::numeric) AS semillas_ejecutados_usd,
    v4_ssot.seeds_invested_for_project_mb(p.id) AS semillas_invertidos_usd,
    v4_ssot.seeds_invested_for_project_mb(p.id) - COALESCE(sc.semillas_ejecutados_usd, 0::numeric) AS semillas_stock_usd,
    COALESCE(sc.agroquimicos_ejecutados_usd, 0::numeric) AS agroquimicos_ejecutados_usd,
    v4_ssot.agrochemicals_invested_for_project_mb(p.id) AS agroquimicos_invertidos_usd,
    v4_ssot.agrochemicals_invested_for_project_mb(p.id) - COALESCE(sc.agroquimicos_ejecutados_usd, 0::numeric) AS agroquimicos_stock_usd,
    COALESCE(sc.fertilizantes_ejecutados_usd, 0::numeric) AS fertilizantes_ejecutados_usd,
    COALESCE(fi.fertilizantes_invertidos_usd, 0::numeric) AS fertilizantes_invertidos_usd,
    COALESCE(fi.fertilizantes_invertidos_usd, 0::numeric) - COALESCE(sc.fertilizantes_ejecutados_usd, 0::numeric) AS fertilizantes_stock_usd,
    COALESCE(sum(v4_ssot.labor_cost_for_lot(l.id)), 0::numeric) AS labores_ejecutados_usd,
    COALESCE(sum(v4_ssot.labor_cost_pre_harvest_for_lot(l.id)), 0::numeric) AS labores_invertidos_usd,
    COALESCE(sum(v4_ssot.labor_cost_pre_harvest_for_lot(l.id)), 0::numeric) - COALESCE(sum(v4_ssot.labor_cost_for_lot(l.id)), 0::numeric) AS labores_stock_usd,
    v4_ssot.lease_executed_for_project(p.id) AS arriendo_ejecutados_usd,
    v4_ssot.lease_invested_for_project(p.id) AS arriendo_invertidos_usd,
    v4_ssot.lease_invested_for_project(p.id) - v4_ssot.lease_executed_for_project(p.id) AS arriendo_stock_usd,
    v4_ssot.admin_cost_total_for_project(p.id) AS estructura_ejecutados_usd,
    v4_ssot.admin_cost_total_for_project(p.id) AS estructura_invertidos_usd,
    0::numeric AS estructura_stock_usd,
    COALESCE(sc.semillas_ejecutados_usd, 0::numeric) AS semilla_cost,
    COALESCE(sc.agroquimicos_ejecutados_usd, 0::numeric) AS insumos_cost,
    COALESCE(sum(v4_ssot.labor_cost_for_lot(l.id)), 0::numeric) AS labores_cost,
    COALESCE(sc.fertilizantes_ejecutados_usd, 0::numeric) AS fertilizantes_cost
   FROM projects p
     LEFT JOIN fields f ON f.project_id = p.id AND f.deleted_at IS NULL
     LEFT JOIN lots l ON l.field_id = f.id AND l.deleted_at IS NULL
     LEFT JOIN v4_calc.dashboard_supply_costs_by_project sc ON sc.project_id = p.id
     LEFT JOIN v4_calc.dashboard_fertilizers_invested_by_project fi ON fi.project_id = p.id
  WHERE p.deleted_at IS NULL
  GROUP BY p.id, sc.semillas_ejecutados_usd, sc.agroquimicos_ejecutados_usd, sc.fertilizantes_ejecutados_usd, fi.fertilizantes_invertidos_usd;

CREATE VIEW v4_report.dashboard_management_balance_field AS
WITH lots_base AS (
    SELECT p.id AS project_id, p.customer_id, p.campaign_id, f.id AS field_id, l.id AS lot_id, l.hectares
    FROM projects p
    JOIN fields f ON f.project_id = p.id AND f.deleted_at IS NULL
    JOIN lots l ON l.field_id = f.id AND l.deleted_at IS NULL
    WHERE p.deleted_at IS NULL
), income_totals AS (
    SELECT project_id, field_id, sum(v4_ssot.income_net_total_for_lot(lot_id)) AS income_usd
    FROM lots_base GROUP BY project_id, field_id
), direct_costs AS (
    SELECT project_id, field_id, sum(v4_ssot.direct_cost_for_lot(lot_id)) AS direct_costs_usd
    FROM lots_base GROUP BY project_id, field_id
), rent_totals AS (
    SELECT project_id, field_id, sum(v4_ssot.rent_per_ha_for_lot(lot_id) * hectares) AS rent_total_usd
    FROM lots_base GROUP BY project_id, field_id
), admin_totals AS (
    SELECT project_id, field_id, sum(v4_ssot.admin_cost_per_ha_for_lot(lot_id) * hectares) AS admin_total_usd
    FROM lots_base GROUP BY project_id, field_id
), field_hectares AS (
    SELECT project_id, field_id, sum(hectares) AS hectares
    FROM lots_base GROUP BY project_id, field_id
), project_hectares AS (
    SELECT project_id, sum(hectares) AS hectares
    FROM lots_base GROUP BY project_id
), supply_costs_field AS (
    SELECT project_id, field_id,
        sum(semillas_usd) AS semillas_usd,
        sum(total_insumos_usd - semillas_usd - fertilizantes_usd) AS agroquimicos_usd,
        sum(fertilizantes_usd) AS fertilizantes_usd,
        sum(total_insumos_usd) AS total_supply_usd
    FROM v4_calc.field_crop_supply_costs_by_lot GROUP BY project_id, field_id
), supply_costs_project AS (
    SELECT project_id,
        sum(semillas_usd) AS semillas_usd,
        sum(total_insumos_usd - semillas_usd - fertilizantes_usd) AS agroquimicos_usd,
        sum(fertilizantes_usd) AS fertilizantes_usd,
        sum(total_insumos_usd) AS total_supply_usd
    FROM v4_calc.field_crop_supply_costs_by_lot GROUP BY project_id
), labor_costs AS (
    SELECT project_id, field_id,
        sum(v4_ssot.labor_cost_for_lot(lot_id)) AS labor_total_usd,
        sum(v4_ssot.labor_cost_pre_harvest_for_lot(lot_id)) AS labor_pre_harvest_usd
    FROM lots_base GROUP BY project_id, field_id
), project_invested AS (
    SELECT DISTINCT project_id,
        v4_ssot.supply_movements_invested_total_for_project(project_id) AS supply_invested_usd,
        v4_ssot.seeds_invested_for_project_mb(project_id) AS seeds_invested_usd,
        v4_ssot.agrochemicals_invested_for_project_mb(project_id) AS agrochemicals_invested_usd
    FROM lots_base
), project_fertilizers_invested AS (
    SELECT project_id, fertilizantes_invertidos_usd
    FROM v4_calc.dashboard_fertilizers_invested_by_project
)
SELECT lb.project_id, lb.customer_id, lb.campaign_id, lb.field_id,
    COALESCE(it.income_usd, 0::numeric) AS income_usd,
    COALESCE(it.income_usd, 0::numeric) - COALESCE(dc.direct_costs_usd, 0::numeric) - COALESCE(rt.rent_total_usd, 0::numeric) - COALESCE(ad.admin_total_usd, 0::numeric) AS operating_result_usd,
    v4_ssot.renta_pct(COALESCE(it.income_usd, 0::numeric) - COALESCE(dc.direct_costs_usd, 0::numeric) - COALESCE(rt.rent_total_usd, 0::numeric) - COALESCE(ad.admin_total_usd, 0::numeric), COALESCE(dc.direct_costs_usd, 0::numeric) + COALESCE(rt.rent_total_usd, 0::numeric) + COALESCE(ad.admin_total_usd, 0::numeric)) AS operating_result_pct,
    COALESCE(dc.direct_costs_usd, 0::numeric) AS costos_directos_ejecutados_usd,
    COALESCE(pi.supply_invested_usd, 0::numeric) *
        CASE WHEN COALESCE(scp.total_supply_usd, 0::numeric) > 0::numeric
            THEN v4_core.safe_div(COALESCE(scf.total_supply_usd, 0::numeric), scp.total_supply_usd)
            ELSE v4_core.safe_div(COALESCE(fh.hectares, 0::numeric), COALESCE(ph.hectares, 0::numeric))
        END + COALESCE(lc.labor_pre_harvest_usd, 0::numeric) AS costos_directos_invertidos_usd,
    COALESCE(pi.supply_invested_usd, 0::numeric) *
        CASE WHEN COALESCE(scp.total_supply_usd, 0::numeric) > 0::numeric
            THEN v4_core.safe_div(COALESCE(scf.total_supply_usd, 0::numeric), scp.total_supply_usd)
            ELSE v4_core.safe_div(COALESCE(fh.hectares, 0::numeric), COALESCE(ph.hectares, 0::numeric))
        END + COALESCE(lc.labor_pre_harvest_usd, 0::numeric) - COALESCE(dc.direct_costs_usd, 0::numeric) AS costos_directos_stock_usd,
    COALESCE(scf.semillas_usd, 0::numeric) AS semillas_ejecutados_usd,
    COALESCE(pi.seeds_invested_usd, 0::numeric) *
        CASE WHEN COALESCE(scp.semillas_usd, 0::numeric) > 0::numeric
            THEN v4_core.safe_div(COALESCE(scf.semillas_usd, 0::numeric), scp.semillas_usd)
            ELSE v4_core.safe_div(COALESCE(fh.hectares, 0::numeric), COALESCE(ph.hectares, 0::numeric))
        END AS semillas_invertidos_usd,
    COALESCE(pi.seeds_invested_usd, 0::numeric) *
        CASE WHEN COALESCE(scp.semillas_usd, 0::numeric) > 0::numeric
            THEN v4_core.safe_div(COALESCE(scf.semillas_usd, 0::numeric), scp.semillas_usd)
            ELSE v4_core.safe_div(COALESCE(fh.hectares, 0::numeric), COALESCE(ph.hectares, 0::numeric))
        END - COALESCE(scf.semillas_usd, 0::numeric) AS semillas_stock_usd,
    COALESCE(scf.agroquimicos_usd, 0::numeric) AS agroquimicos_ejecutados_usd,
    COALESCE(pi.agrochemicals_invested_usd, 0::numeric) *
        CASE WHEN COALESCE(scp.agroquimicos_usd, 0::numeric) > 0::numeric
            THEN v4_core.safe_div(COALESCE(scf.agroquimicos_usd, 0::numeric), scp.agroquimicos_usd)
            ELSE v4_core.safe_div(COALESCE(fh.hectares, 0::numeric), COALESCE(ph.hectares, 0::numeric))
        END AS agroquimicos_invertidos_usd,
    COALESCE(pi.agrochemicals_invested_usd, 0::numeric) *
        CASE WHEN COALESCE(scp.agroquimicos_usd, 0::numeric) > 0::numeric
            THEN v4_core.safe_div(COALESCE(scf.agroquimicos_usd, 0::numeric), scp.agroquimicos_usd)
            ELSE v4_core.safe_div(COALESCE(fh.hectares, 0::numeric), COALESCE(ph.hectares, 0::numeric))
        END - COALESCE(scf.agroquimicos_usd, 0::numeric) AS agroquimicos_stock_usd,
    COALESCE(scf.fertilizantes_usd, 0::numeric) AS fertilizantes_ejecutados_usd,
    COALESCE(pfi.fertilizantes_invertidos_usd, 0::numeric) *
        CASE WHEN COALESCE(scp.fertilizantes_usd, 0::numeric) > 0::numeric
            THEN v4_core.safe_div(COALESCE(scf.fertilizantes_usd, 0::numeric), scp.fertilizantes_usd)
            ELSE v4_core.safe_div(COALESCE(fh.hectares, 0::numeric), COALESCE(ph.hectares, 0::numeric))
        END AS fertilizantes_invertidos_usd,
    COALESCE(pfi.fertilizantes_invertidos_usd, 0::numeric) *
        CASE WHEN COALESCE(scp.fertilizantes_usd, 0::numeric) > 0::numeric
            THEN v4_core.safe_div(COALESCE(scf.fertilizantes_usd, 0::numeric), scp.fertilizantes_usd)
            ELSE v4_core.safe_div(COALESCE(fh.hectares, 0::numeric), COALESCE(ph.hectares, 0::numeric))
        END - COALESCE(scf.fertilizantes_usd, 0::numeric) AS fertilizantes_stock_usd,
    COALESCE(lc.labor_total_usd, 0::numeric) AS labores_ejecutados_usd,
    COALESCE(lc.labor_pre_harvest_usd, 0::numeric) AS labores_invertidos_usd,
    COALESCE(lc.labor_pre_harvest_usd, 0::numeric) - COALESCE(lc.labor_total_usd, 0::numeric) AS labores_stock_usd,
    COALESCE(rt.rent_total_usd, 0::numeric) AS arriendo_ejecutados_usd,
    COALESCE(rt.rent_total_usd, 0::numeric) AS arriendo_invertidos_usd,
    0::numeric AS arriendo_stock_usd,
    COALESCE(ad.admin_total_usd, 0::numeric) AS estructura_ejecutados_usd,
    COALESCE(ad.admin_total_usd, 0::numeric) AS estructura_invertidos_usd,
    0::numeric AS estructura_stock_usd,
    COALESCE(scf.semillas_usd, 0::numeric) AS semilla_cost,
    COALESCE(scf.agroquimicos_usd, 0::numeric) AS insumos_cost,
    COALESCE(lc.labor_total_usd, 0::numeric) AS labores_cost,
    COALESCE(scf.fertilizantes_usd, 0::numeric) AS fertilizantes_cost
FROM (SELECT DISTINCT project_id, customer_id, campaign_id, field_id FROM lots_base) lb
LEFT JOIN income_totals it ON it.project_id = lb.project_id AND it.field_id = lb.field_id
LEFT JOIN direct_costs dc ON dc.project_id = lb.project_id AND dc.field_id = lb.field_id
LEFT JOIN rent_totals rt ON rt.project_id = lb.project_id AND rt.field_id = lb.field_id
LEFT JOIN admin_totals ad ON ad.project_id = lb.project_id AND ad.field_id = lb.field_id
LEFT JOIN field_hectares fh ON fh.project_id = lb.project_id AND fh.field_id = lb.field_id
LEFT JOIN project_hectares ph ON ph.project_id = lb.project_id
LEFT JOIN supply_costs_field scf ON scf.project_id = lb.project_id AND scf.field_id = lb.field_id
LEFT JOIN supply_costs_project scp ON scp.project_id = lb.project_id
LEFT JOIN labor_costs lc ON lc.project_id = lb.project_id AND lc.field_id = lb.field_id
LEFT JOIN project_invested pi ON pi.project_id = lb.project_id
LEFT JOIN project_fertilizers_invested pfi ON pfi.project_id = lb.project_id;

COMMIT;
