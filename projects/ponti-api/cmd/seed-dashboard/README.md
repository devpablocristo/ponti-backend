# 🌱 Seed Dashboard - Datos para el Dashboard

## 📋 Descripción

Este módulo contiene scripts SQL para poblar la base de datos con datos reales y ricos para el dashboard de Ponti. Los seeds están diseñados para funcionar con la estructura actual de las tablas y proporcionar métricas realistas.

## 🚀 Uso Rápido

### Opción 1: Usar Makefile (Recomendado)
```bash
cd cmd/seed-dashboard
make all
```

### Opción 2: Ejecutar manualmente
```bash
# 1. Copiar scripts al contenedor
docker cp sql/00_base_data.sql ponti-api-ponti-db-1:/tmp/
docker cp sql/99_complete_dashboard_data.sql ponti-api-ponti-db-1:/tmp/

# 2. Ejecutar en orden
docker compose exec ponti-db psql -U admin -d ponti_api_db -f /tmp/00_base_data.sql
docker compose exec ponti-db psql -U admin -d ponti_api_db -f /tmp/99_complete_dashboard_data.sql
```

## 📊 Datos que se cargan

### 🔧 Datos Base (00_base_data.sql)
- **6 Types**: Semillas, Fertilizantes, Agroquímicos, Combustibles, Labores, Maquinaria
- **7 Categories**: Semillas, Fertilizantes, Herbicidas, Fungicidas, Insecticidas, Diesel, Nafta
- **6 Labor Types**: Siembra, Mantenimiento, Cosecha, Post-cosecha, Transporte, Almacenamiento
- **8 Labor Categories**: Siembra directa, Fertilización, Pulverización, Riego, Cosecha, Secado, Transporte
- **6 Providers**: AgroSeed S.A., Fertilizantes del Sur, AgroQuímica Plus, etc.
- **6 Users**: Admin, Manager, Operadores, Contador, Técnico

### 📈 Datos del Dashboard (99_complete_dashboard_data.sql)
- **20 Supplies**: Semillas, fertilizantes, agroquímicos, combustibles con precios reales
- **20 Labors**: Diferentes tipos de labores con contratistas reales
- **20 Workorders**: Órdenes de trabajo con fechas reales (Enero a Julio 2024)
- **5 Investors**: Fondos de inversión con porcentajes de contribución
- **Stock**: Inventario valorado en ~$700,000
- **8 Invoices**: Facturas por $173,000 (pagadas, pendientes, borrador)
- **Lots actualizados**: Con fechas de siembra y toneladas cosechadas

## 🎯 Métricas Esperadas del Dashboard

Después de cargar los seeds, el dashboard debería mostrar:

- **Siembra**: 46.0 ha / 46.0 ha (100%)
- **Cosecha**: 19.0 toneladas totales
- **Costos**: ~$1,200 / $25,000 (4.8%)
- **Ingresos**: $80,000 (facturas pagadas)
- **Stock**: ~$700,000
- **Contribuciones**: 100% (5 inversores)
- **Crops**: 4 cultivos con datos reales

## 🔍 Verificación

Para verificar que los seeds funcionaron correctamente:

```bash
# Copiar script de prueba
docker cp sql/test_seeds.sql ponti-api-ponti-db-1:/tmp/

# Ejecutar verificación
docker compose exec ponti-db psql -U admin -d ponti_api_db -f /tmp/test_seeds.sql
```

## 🧹 Limpieza

```bash
make clean
```

## 📝 Notas Importantes

1. **Orden de ejecución**: Siempre ejecutar `00_base_data.sql` antes que `99_complete_dashboard_data.sql`
2. **Estructura de tablas**: Los seeds están diseñados para la estructura actual de las tablas
3. **Datos reales**: Todos los precios, fechas y cantidades son realistas para un proyecto agrícola
4. **Sin conflictos**: Usa `ON CONFLICT DO NOTHING` para evitar errores de duplicados

## 🐛 Solución de Problemas

### Error: "column does not exist"
- Verificar que la estructura de las tablas coincida con la esperada
- Ejecutar `\d nombre_tabla` en psql para ver la estructura real

### Error: "foreign key constraint"
- Verificar que los datos base se cargaron correctamente
- Ejecutar `00_base_data.sql` primero

### Dashboard no muestra datos
- Verificar que la función `get_dashboard_payload` existe
- Ejecutar `SELECT get_dashboard_payload(1,1,1,1);` para probar

## 📞 Soporte

Si encuentras problemas, verifica:
1. Estructura de las tablas en la base de datos
2. Orden de ejecución de los scripts
3. Logs de errores en psql
