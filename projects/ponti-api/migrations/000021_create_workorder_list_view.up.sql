CREATE OR REPLACE VIEW workorder_list_view AS
SELECT
  w.number,
  p.name              AS project_name,
  f.name              AS field_name,
  l.name              AS lot_name,
  w.date,
  c.name              AS crop_name,
  lb.name             AS labor_name,
  t.name              AS type_name,
  w.contractor,
  w.effective_area    AS surface_ha,
  s.name              AS supply_name,
  wi.total_used       AS consumption,
  cat.name            AS category_name,
  wi.final_dose       AS dose,
  (wi.final_dose * s.price)                     AS cost_per_ha,
  s.price                                       AS unit_price,
  ((wi.final_dose * s.price) * w.effective_area) AS total_cost
FROM workorders w
JOIN projects p       ON p.id   = w.project_id
JOIN fields f         ON f.id   = w.field_id
JOIN lots l           ON l.id   = w.lot_id
JOIN crops c          ON c.id   = w.crop_id
JOIN labors lb        ON lb.id  = w.labor_id
-- ahora unimos por workorder_id en lugar de number
JOIN workorder_items wi ON wi.workorder_id = w.id
JOIN supplies s        ON s.id        = wi.supply_id
JOIN types t           ON t.id        = s.type_id
JOIN categories cat    ON cat.id      = s.category_id;
