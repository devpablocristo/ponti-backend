# Test del Endpoint Data Integrity - LEFT/RIGHT

## ✅ Estructura Implementada

Cada control ahora muestra **DOS cálculos independientes**:
- **LEFT** (Origen): Valor correcto desde la fuente de verdad
- **RIGHT** (Destino): Valor a validar

## Endpoint
```
GET /api/v1/data-integrity/costs-check?project_id=11
```

## Probar con curl

```bash
curl -X GET "http://localhost:8080/api/v1/data-integrity/costs-check?project_id=11" \
  -H "X-API-KEY: abc123secreta" \
  -H "X-USER-ID: 123" | jq
```

## Ver solo Control 1
```bash
curl -X GET "http://localhost:8080/api/v1/data-integrity/costs-check?project_id=11" \
  -H "X-API-KEY: abc123secreta" \
  -H "X-USER-ID: 123" -s | jq '.checks[0]'
```

## Ver solo Control 2 (Órdenes RAW vs Lotes)
```bash
curl -X GET "http://localhost:8080/api/v1/data-integrity/costs-check?project_id=11" \
  -H "X-API-KEY: abc123secreta" \
  -H "X-USER-ID: 123" -s | jq '.checks[1]'
```

## Respuesta Esperada (Control 2 ejemplo)

```json
{
  "control_number": 2,
  "source_module": "Órdenes de trabajo",
  "data_to_verify": "Costos directos ejecutados",
  "target_module": "Lotes",
  "control_rule": "∑(Costo_directo_ha_lote × Superficie_lote) = ∑(Órdenes.costo_total)",
  "left_calculation": "∑(workorders RAW)",
  "left_value": "18454.39",
  "left_source": "Tabla workorders RAW",
  "right_calculation": "∑(cost_usd_per_ha × sowed_area_ha)",
  "right_value": "18454.39",
  "right_source": "Vista v3_lot_list",
  "difference": "0.00",
  "status": "OK",
  "tolerance": "0.00"
}
```

## Estados posibles
- **OK**: Diferencia dentro de la tolerancia
- **ERROR**: Diferencia fuera de la tolerancia

## Tolerancias
- **Controles 1-4**: Tolerancia = 0 (deben ser exactos)
- **Controles 5-14**: Tolerancia = ±1 USD

## Ver todos los controles con ERROR
```bash
curl -X GET "http://localhost:8080/api/v1/data-integrity/costs-check?project_id=11" \
  -H "X-API-KEY: abc123secreta" \
  -H "X-USER-ID: 123" -s | jq '.checks[] | select(.status == "ERROR")'
```

## Ver resumen de estados
```bash
curl -X GET "http://localhost:8080/api/v1/data-integrity/costs-check?project_id=11" \
  -H "X-API-KEY: abc123secreta" \
  -H "X-USER-ID: 123" -s | jq '.checks[] | {control: .control_number, status: .status, diff: .difference}'
```

## Nota Importante
**LEFT es siempre el valor correcto** (origen/fuente de verdad)
**RIGHT se valida contra LEFT**

La diferencia se calcula como: `LEFT - RIGHT`

