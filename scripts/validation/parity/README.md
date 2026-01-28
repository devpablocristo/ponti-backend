# Scripts de Validación de Paridad

## Uso

```bash
# Validar contrato (schema igual)
make validate-contract

# Validar paridad de datos
make validate-parity

# Validar todo
make validate-all

# Una vista específica
psql -v ON_ERROR_STOP=1 -d ponti -f scripts/validation/sanity_explain_v4_lot_metrics.sql
```

## Tipos de checks

### Verificación (datos)
- UNICIDAD_V4: Hay duplicados en v4
- ROWCOUNT: Diferente cantidad de filas
- NULL_MISMATCH: NULL inesperado
- NUMERIC_DIFF: Diferencia numérica > 0.01

## Interpretación
- **0 errores** = PASS ✅
- **> 0 errores** = FAIL ❌ (script hace RAISE EXCEPTION)

## Archivos

| Script | Propósito |
|--------|-----------|
| sanity_explain_v4_lot_metrics.sql | Verificar que la vista es ejecutable |
