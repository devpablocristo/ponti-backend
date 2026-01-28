-- =============================================================================
-- MIGRACIÓN 000332: Rollback SSOT siembra + per-ha inversor
-- =============================================================================
--
-- Propósito:
-- 1) Restaurar seeded_area_for_lot basado en sowing_date
-- 2) Restaurar per-ha inversor basado en total_seeded_area_ha
-- Nota: Comentarios en español, código en inglés
--

BEGIN;

-- 1) Rollback siembra basada en fecha del lote
CREATE OR REPLACE FUNCTION v3_lot_ssot.seeded_area_for_lot(p_lot_id bigint) 
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT v3_core_ssot.seeded_area(l.sowing_date, l.hectares::numeric)
  FROM public.lots l
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;

-- 2) Rollback per-ha inversor basado en total_seeded_area_ha
CREATE OR REPLACE VIEW v4_report.investor_contribution_data AS
WITH investor_base AS (
  SELECT
    pi.project_id,
    pi.investor_id,
    i.name AS investor_name,
    pi.percentage AS share_pct_agreed
  FROM public.project_investors pi
  JOIN public.investors i ON i.id = pi.investor_id AND i.deleted_at IS NULL
  WHERE pi.deleted_at IS NULL
),
project_seed_data AS (
  SELECT
    project_id,
    MAX(total_seeded_area_ha)::numeric AS total_seeded_area_ha
  FROM v4_report.investor_contribution_categories
  GROUP BY project_id
),
investor_harvest_real AS (
  SELECT
    w.project_id,
    w.investor_id,
    COALESCE(SUM(lab.price * w.effective_area), 0)::numeric AS harvest_real_usd
  FROM public.workorders w
  JOIN public.labors lab ON w.labor_id = lab.id AND lab.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = lab.category_id
  WHERE w.deleted_at IS NULL
    AND cat.type_id = 4
    AND cat.name = 'Cosecha'
  GROUP BY w.project_id, w.investor_id
),
harvest_totals AS (
  SELECT
    psd.project_id,
    COALESCE(SUM(hr.harvest_real_usd), 0)::numeric AS total_harvest_usd,
    CASE
      WHEN COALESCE(psd.total_seeded_area_ha, 0) > 0
      THEN COALESCE(SUM(hr.harvest_real_usd), 0) / psd.total_seeded_area_ha
      ELSE 0
    END::numeric AS total_harvest_usd_ha
  FROM project_seed_data psd
  LEFT JOIN investor_harvest_real hr ON hr.project_id = psd.project_id
  GROUP BY psd.project_id, psd.total_seeded_area_ha
)
SELECT
  pb.project_id,
  pb.project_name,
  pb.customer_id,
  pb.customer_name,
  pb.campaign_id,
  pb.campaign_name,
  -- Headers de inversores (desde SSOT)
  (
    SELECT jsonb_agg(
      jsonb_build_object(
        'investor_id', irc.investor_id,
        'investor_name', irc.investor_name,
        'share_pct', ROUND(irc.share_pct_agreed::numeric, 2)
      )
      ORDER BY irc.investor_id
    )
    FROM v4_calc.investor_real_contributions irc
    WHERE irc.project_id = pb.project_id
  ) AS investor_headers,
  -- Datos generales del proyecto
  jsonb_build_object(
    'surface_total_ha', ROUND(pb.surface_total_ha::numeric, 2),
    'lease_fixed_total_usd', ROUND(pb.lease_fixed_total_usd::numeric, 2),
    'lease_is_fixed', pb.lease_is_fixed,
    'lease_per_ha_usd', ROUND(pb.lease_per_ha_usd::numeric, 2),
    'admin_total_usd', ROUND(pb.admin_total_usd::numeric, 2),
    'admin_per_ha_usd', ROUND(pb.admin_per_ha_usd::numeric, 2)
  ) AS general_project_data,
  -- Categorías de aportes (usa datos desde SSOT)
  (
    SELECT jsonb_agg(
      jsonb_build_object(
        'key', cat.key,
        'sort_index', cat.sort_index,
        'type', cat.type,
        'label', cat.label,
        'total_usd', ROUND(cat.total_usd::numeric, 2),
        'total_usd_ha', ROUND(cat.total_usd_ha::numeric, 2),
        'investors', cat.investors,
        'requires_manual_attribution', cat.requires_manual_attribution,
        'attribution_note', cat.attribution_note
      )
      ORDER BY cat.sort_index
    )
    FROM (
      -- Agroquímicos
      SELECT
        'agrochemicals'::text AS key,
        1 AS sort_index,
        'pre_harvest'::text AS type,
        'Agroquímicos'::text AS label,
        cc.agrochemicals_total_usd AS total_usd,
        CASE
          WHEN cc.total_seeded_area_ha > 0
          THEN cc.agrochemicals_total_usd / cc.total_seeded_area_ha
          ELSE 0
        END AS total_usd_ha,
        (
          SELECT jsonb_agg(
            jsonb_build_object(
              'investor_id', irc.investor_id,
              'investor_name', irc.investor_name,
              'amount_usd', ROUND(irc.agrochemicals_real_usd::numeric, 2),
              'share_pct', ROUND(
                CASE
                  WHEN cc.agrochemicals_total_usd > 0
                  THEN (irc.agrochemicals_real_usd / cc.agrochemicals_total_usd * 100)
                  ELSE 0
                END::numeric, 2
              )
            )
            ORDER BY irc.investor_id
          )
          FROM v4_calc.investor_real_contributions irc
          WHERE irc.project_id = pb.project_id
        ) AS investors,
        false AS requires_manual_attribution,
        NULL AS attribution_note
      FROM v4_report.investor_contribution_categories cc
      WHERE cc.project_id = pb.project_id

      UNION ALL

      -- Fertilizantes
      SELECT
        'fertilizers'::text,
        2,
        'pre_harvest'::text,
        'Fertilizantes'::text,
        cc.fertilizers_total_usd,
        CASE
          WHEN cc.total_seeded_area_ha > 0
          THEN cc.fertilizers_total_usd / cc.total_seeded_area_ha
          ELSE 0
        END,
        (
          SELECT jsonb_agg(
            jsonb_build_object(
              'investor_id', irc.investor_id,
              'investor_name', irc.investor_name,
              'amount_usd', ROUND(irc.fertilizers_real_usd::numeric, 2),
              'share_pct', ROUND(
                CASE
                  WHEN cc.fertilizers_total_usd > 0
                  THEN (irc.fertilizers_real_usd / cc.fertilizers_total_usd * 100)
                  ELSE 0
                END::numeric, 2
              )
            )
            ORDER BY irc.investor_id
          )
          FROM v4_calc.investor_real_contributions irc
          WHERE irc.project_id = pb.project_id
        ),
        false,
        NULL
      FROM v4_report.investor_contribution_categories cc
      WHERE cc.project_id = pb.project_id

      UNION ALL

      -- Semillas
      SELECT
        'seeds'::text,
        3,
        'pre_harvest'::text,
        'Semilla'::text,
        cc.seeds_total_usd,
        CASE
          WHEN cc.total_seeded_area_ha > 0
          THEN cc.seeds_total_usd / cc.total_seeded_area_ha
          ELSE 0
        END,
        (
          SELECT jsonb_agg(
            jsonb_build_object(
              'investor_id', irc.investor_id,
              'investor_name', irc.investor_name,
              'amount_usd', ROUND(irc.seeds_real_usd::numeric, 2),
              'share_pct', ROUND(
                CASE
                  WHEN cc.seeds_total_usd > 0
                  THEN (irc.seeds_real_usd / cc.seeds_total_usd * 100)
                  ELSE 0
                END::numeric, 2
              )
            )
            ORDER BY irc.investor_id
          )
          FROM v4_calc.investor_real_contributions irc
          WHERE irc.project_id = pb.project_id
        ),
        false,
        NULL
      FROM v4_report.investor_contribution_categories cc
      WHERE cc.project_id = pb.project_id

      UNION ALL

      -- Labores Generales
      SELECT
        'general_labors'::text,
        4,
        'pre_harvest'::text,
        'Labores Generales'::text,
        cc.general_labors_total_usd,
        CASE
          WHEN cc.total_seeded_area_ha > 0
          THEN cc.general_labors_total_usd / cc.total_seeded_area_ha
          ELSE 0
        END,
        (
          SELECT jsonb_agg(
            jsonb_build_object(
              'investor_id', irc.investor_id,
              'investor_name', irc.investor_name,
              'amount_usd', ROUND(irc.general_labors_real_usd::numeric, 2),
              'share_pct', ROUND(
                CASE
                  WHEN cc.general_labors_total_usd > 0
                  THEN (irc.general_labors_real_usd / cc.general_labors_total_usd * 100)
                  ELSE 0
                END::numeric, 2
              )
            )
            ORDER BY irc.investor_id
          )
          FROM v4_calc.investor_real_contributions irc
          WHERE irc.project_id = pb.project_id
        ),
        false,
        NULL
      FROM v4_report.investor_contribution_categories cc
      WHERE cc.project_id = pb.project_id

      UNION ALL

      -- Siembra
      SELECT
        'sowing'::text,
        5,
        'pre_harvest'::text,
        'Siembra'::text,
        cc.sowing_total_usd,
        CASE
          WHEN cc.total_seeded_area_ha > 0
          THEN cc.sowing_total_usd / cc.total_seeded_area_ha
          ELSE 0
        END,
        (
          SELECT jsonb_agg(
            jsonb_build_object(
              'investor_id', irc.investor_id,
              'investor_name', irc.investor_name,
              'amount_usd', ROUND(irc.sowing_real_usd::numeric, 2),
              'share_pct', ROUND(
                CASE
                  WHEN cc.sowing_total_usd > 0
                  THEN (irc.sowing_real_usd / cc.sowing_total_usd * 100)
                  ELSE 0
                END::numeric, 2
              )
            )
            ORDER BY irc.investor_id
          )
          FROM v4_calc.investor_real_contributions irc
          WHERE irc.project_id = pb.project_id
        ),
        false,
        NULL
      FROM v4_report.investor_contribution_categories cc
      WHERE cc.project_id = pb.project_id

      UNION ALL

      -- Riego
      SELECT
        'irrigation'::text,
        6,
        'pre_harvest'::text,
        'Riego'::text,
        cc.irrigation_total_usd,
        CASE
          WHEN cc.total_seeded_area_ha > 0
          THEN cc.irrigation_total_usd / cc.total_seeded_area_ha
          ELSE 0
        END,
        (
          SELECT jsonb_agg(
            jsonb_build_object(
              'investor_id', irc.investor_id,
              'investor_name', irc.investor_name,
              'amount_usd', ROUND(irc.irrigation_real_usd::numeric, 2),
              'share_pct', ROUND(
                CASE
                  WHEN cc.irrigation_total_usd > 0
                  THEN (irc.irrigation_real_usd / cc.irrigation_total_usd * 100)
                  ELSE 0
                END::numeric, 2
              )
            )
            ORDER BY irc.investor_id
          )
          FROM v4_calc.investor_real_contributions irc
          WHERE irc.project_id = pb.project_id
        ),
        false,
        NULL
      FROM v4_report.investor_contribution_categories cc
      WHERE cc.project_id = pb.project_id

      UNION ALL

      -- Arriendo Capitalizable
      SELECT
        'capitalizable_lease'::text,
        7,
        'pre_harvest'::text,
        'Arriendo Capitalizable'::text,
        cc.rent_capitalizable_total_usd,
        CASE
          WHEN cc.total_seeded_area_ha > 0
          THEN cc.rent_capitalizable_total_usd / cc.total_seeded_area_ha
          ELSE 0
        END,
        (
          SELECT jsonb_agg(
            jsonb_build_object(
              'investor_id', irc.investor_id,
              'investor_name', irc.investor_name,
              'amount_usd', ROUND(irc.rent_real_usd::numeric, 2),
              'share_pct', ROUND(
                CASE
                  WHEN cc.rent_capitalizable_total_usd > 0
                  THEN (irc.rent_real_usd / cc.rent_capitalizable_total_usd * 100)
                  ELSE 0
                END::numeric, 2
              )
            )
            ORDER BY irc.investor_id
          )
          FROM v4_calc.investor_real_contributions irc
          WHERE irc.project_id = pb.project_id
        ),
        true,
        'Requiere atribución manual por inversor'
      FROM v4_report.investor_contribution_categories cc
      WHERE cc.project_id = pb.project_id

      UNION ALL

      -- Administración y Estructura
      SELECT
        'administration_structure'::text,
        8,
        'pre_harvest'::text,
        'Administración y Estructura'::text,
        cc.administration_total_usd,
        CASE
          WHEN cc.total_seeded_area_ha > 0
          THEN cc.administration_total_usd / cc.total_seeded_area_ha
          ELSE 0
        END,
        (
          SELECT jsonb_agg(
            jsonb_build_object(
              'investor_id', irc.investor_id,
              'investor_name', irc.investor_name,
              'amount_usd', ROUND(irc.administration_real_usd::numeric, 2),
              'share_pct', ROUND(
                CASE
                  WHEN cc.administration_total_usd > 0
                  THEN (irc.administration_real_usd / cc.administration_total_usd * 100)
                  ELSE 0
                END::numeric, 2
              )
            )
            ORDER BY irc.investor_id
          )
          FROM v4_calc.investor_real_contributions irc
          WHERE irc.project_id = pb.project_id
        ),
        true,
        'Requiere atribución manual por inversor'
      FROM v4_report.investor_contribution_categories cc
      WHERE cc.project_id = pb.project_id
    ) AS cat
  ) AS contribution_categories,
  -- Comparación de aportes acordado vs real (usa SSOT)
  (
    SELECT jsonb_agg(
      jsonb_build_object(
        'investor_id', irc.investor_id,
        'investor_name', irc.investor_name,
        'agreed_share_pct', irc.share_pct_agreed,
        'agreed_usd', ROUND((irc.project_total_contributions_usd * irc.share_pct_agreed / 100)::numeric, 2),
        'actual_usd', ROUND(irc.total_real_contribution_usd::numeric, 2),
        'share_pct', ROUND(irc.contributions_progress_pct::numeric, 2),
        'adjustment_usd', ROUND((irc.total_real_contribution_usd - (irc.project_total_contributions_usd * irc.share_pct_agreed / 100))::numeric, 2)
      )
      ORDER BY irc.investor_id
    )
    FROM v4_calc.investor_real_contributions irc
    WHERE irc.project_id = pb.project_id
  ) AS investor_contribution_comparison,
  -- Liquidación de cosecha
  jsonb_build_object(
    'rows', jsonb_build_array(
      jsonb_build_object(
        'key', 'harvest',
        'type', 'harvest',
        'total_usd', ROUND(COALESCE(ht.total_harvest_usd, 0)::numeric, 2),
        'total_us_ha', ROUND(COALESCE(ht.total_harvest_usd_ha, 0)::numeric, 2),
        'investors', COALESCE((
          SELECT jsonb_agg(
            jsonb_build_object(
              'investor_id', ib.investor_id,
              'investor_name', ib.investor_name,
              'amount_usd', ROUND(COALESCE(hr.harvest_real_usd, 0)::numeric, 2),
              'share_pct', ROUND(
                CASE 
                  WHEN COALESCE(ht.total_harvest_usd, 0)::numeric > 0 
                  THEN (COALESCE(hr.harvest_real_usd, 0)::numeric / COALESCE(ht.total_harvest_usd, 0)::numeric * 100)::numeric
                  ELSE 0::numeric
                END, 2
              )
            )
            ORDER BY ib.investor_id
          )
          FROM investor_base ib
          LEFT JOIN investor_harvest_real hr
            ON hr.project_id = ib.project_id AND hr.investor_id = ib.investor_id
          WHERE ib.project_id = pb.project_id
        ), '[]'::jsonb)
      ),
      jsonb_build_object(
        'key', 'totals',
        'type', 'totals',
        'total_usd', ROUND(COALESCE(ht.total_harvest_usd, 0)::numeric, 2),
        'total_us_ha', ROUND(COALESCE(ht.total_harvest_usd_ha, 0)::numeric, 2),
        'investors', COALESCE((
          SELECT jsonb_agg(
            jsonb_build_object(
              'investor_id', ib.investor_id,
              'investor_name', ib.investor_name,
              'amount_usd', ROUND(COALESCE(hr.harvest_real_usd, 0)::numeric, 2),
              'share_pct', ROUND(
                CASE 
                  WHEN COALESCE(ht.total_harvest_usd, 0)::numeric > 0 
                  THEN (COALESCE(hr.harvest_real_usd, 0)::numeric / COALESCE(ht.total_harvest_usd, 0)::numeric * 100)::numeric
                  ELSE 0::numeric
                END, 2
              )
            )
            ORDER BY ib.investor_id
          )
          FROM investor_base ib
          LEFT JOIN investor_harvest_real hr
            ON hr.project_id = ib.project_id AND hr.investor_id = ib.investor_id
          WHERE ib.project_id = pb.project_id
        ), '[]'::jsonb)
      )
    ),
    'footer_payment_agreed', COALESCE((
      SELECT jsonb_agg(
        jsonb_build_object(
          'investor_id', ib.investor_id,
          'investor_name', ib.investor_name,
          'amount_usd', ROUND((COALESCE(ht.total_harvest_usd, 0)::numeric * ib.share_pct_agreed::numeric / 100)::numeric, 2),
          'share_pct', ROUND(ib.share_pct_agreed::numeric, 2)
        )
        ORDER BY ib.investor_id
      )
      FROM investor_base ib
      WHERE ib.project_id = pb.project_id
    ), '[]'::jsonb),
    'footer_payment_adjustment', COALESCE((
      SELECT jsonb_agg(
        jsonb_build_object(
          'investor_id', ib.investor_id,
          'investor_name', ib.investor_name,
          'amount_usd', ROUND((
            COALESCE(hr.harvest_real_usd, 0)::numeric -
            (COALESCE(ht.total_harvest_usd, 0)::numeric * ib.share_pct_agreed::numeric / 100)
          )::numeric, 2),
          'share_pct', ROUND(ib.share_pct_agreed::numeric, 2)
        )
        ORDER BY ib.investor_id
      )
      FROM investor_base ib
      LEFT JOIN investor_harvest_real hr
        ON hr.project_id = ib.project_id AND hr.investor_id = ib.investor_id
      WHERE ib.project_id = pb.project_id
    ), '[]'::jsonb)
  ) AS harvest_settlement
FROM v4_report.investor_project_base pb
JOIN v4_report.investor_contribution_categories cc ON cc.project_id = pb.project_id
LEFT JOIN harvest_totals ht ON ht.project_id = pb.project_id
ORDER BY pb.project_id;

COMMENT ON VIEW v4_report.investor_contribution_data IS
'Vista final para informe de Aportes por Inversor. SSOT: usa v4_calc.investor_real_contributions (000327).';

COMMIT;
