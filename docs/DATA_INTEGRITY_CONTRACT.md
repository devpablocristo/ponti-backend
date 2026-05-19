# Data Integrity: contrato de controles de calculo

`/data-integrity/costs-check` es una herramienta de auditoria numerica. Su contrato no es "todo rojo o verde", sino distinguir controles que comparan la misma magnitud de controles que hoy revelan definiciones de negocio no alineadas.

## Estados

- `OK`: el valor del sistema y sus recalculos comparables coinciden dentro de tolerancia.
- `ERROR`: el control fuerte compara la misma magnitud y la diferencia excede la tolerancia.
- `WARNING`: el control detecta que las fuentes todavia no son semanticamente equivalentes; muestra diferencias, pero no debe contarse como formula rota hasta alinear definiciones.
- `SKIPPED`: faltan datos para ejecutar el control.

## Tipos de control

- `STRONG`: control numerico valido entre fuentes comparables.
- `WEAK`: sanity check interno; sirve para detectar inconsistencias dentro de la misma pantalla, pero no valida una segunda formula independiente.
- `FORMULA_ALIGNMENT`: control que hoy compara fuentes con definiciones distintas. Debe guiar correcciones de formulas antes de usarse como gate numerico.

## Controles que requieren alineacion

- Control 8, Administracion y Estructura: `lot_list` expone `admin_cost_per_ha`, mientras Aportes/field crop usan administracion prorrateada por superficie total del proyecto.
- Control 9, Arriendo Capitalizable: `lot_list` usa arriendo total por lote, mientras Aportes/field crop usan arriendo capitalizable/fijo.
- Controles 11, 12 y 13, Resultado Operativo: heredan la diferencia anterior porque el resultado operativo depende de arriendo y administracion.
- Control 17, Renta / Total activo: mezcla total de costos dashboard, ordenes, arriendo y estructura. Debe alinearse con la misma definicion de activo/renta antes de tratar diferencias como error.

## Controles debiles actuales

- Control 5: suma componentes del mismo dashboard contra el total del dashboard.
- Control 14: stock se recalcula desde componentes del mismo dashboard; todavia no usa un camino independiente desde stock/movimientos.

## Regla operativa

Para validar cambios de calculo, usar como gate solo controles `STRONG` con `status=ERROR`. Los controles `WARNING` son backlog de alineacion de formulas y deben resolverse antes de convertir data integrity en golden master completo.
