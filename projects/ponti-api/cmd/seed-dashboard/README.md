# 🌱 SEED DASHBOARD

Herramienta para poblar la base de datos con datos mínimos necesarios para que el dashboard funcione correctamente.

## 🚀 Características

- **Sistema Genérico**: Carga automáticamente todos los scripts SQL del directorio `sql/` en orden alfabético
- **Orden Controlado**: Usa prefijos numéricos (01_, 02_, 03_, etc.) para controlar el orden de ejecución
- **Flexible**: Agrega nuevos cultivos simplemente creando nuevos scripts SQL
- **Reseteo Completo**: Opción para limpiar todos los datos antes de sembrar

## 📁 Estructura de Archivos

```
cmd/seed-dashboard/
├── main.go                    # Aplicación principal (genérica)
├── Makefile                   # Comandos simplificados
├── README.md                  # Este archivo
└── sql/                       # Directorio con scripts SQL
    ├── 01_dashboard_minimal.sql      # Datos básicos (Soja)
    ├── 02_add_maiz_crop.sql          # Agregar Maíz
    ├── 03_add_trigo_crop.sql         # Agregar Trigo
    └── 04_add_girasol_crop.sql      # Agregar Girasol
```

## 🛠️ Uso

### Comando Básico
```bash
# Crear datos del dashboard (ejecuta todos los scripts SQL en orden)
go run cmd/seed-dashboard/main.go

# Limpiar y crear datos del dashboard
go run cmd/seed-dashboard/main.go -reset

# Mostrar ayuda
go run cmd/seed-dashboard/main.go -help
```

### Con Makefile
```bash
# Cargar todos los scripts SQL del dashboard
make all

# Limpiar archivos compilados
make clean

# Mostrar ayuda
make help
```

## 📊 Datos Creados

### Script 01: Datos Básicos
- **1 cultivo**: Soja
- **1 cliente**: Cliente A
- **1 campaña**: 2024-2025
- **1 proyecto**: Proyecto Soja
- **1 campo**: Campo Norte
- **1 lote**: Lote A1 con Soja (10.5 hectáreas)

### Script 02: Agregar Maíz
- **1 cultivo adicional**: Maíz
- **1 campo adicional**: Campo Sur
- **1 lote adicional**: Lote B1 con Maíz (8.5 hectáreas)

### Script 03: Agregar Trigo
- **1 cultivo adicional**: Trigo
- **1 campo adicional**: Campo Este
- **1 lote adicional**: Lote C1 con Trigo (12.0 hectáreas)

### Script 04: Agregar Girasol
- **1 cultivo adicional**: Girasol
- **1 campo adicional**: Campo Oeste
- **1 lote adicional**: Lote D1 con Girasol (15.0 hectáreas)

## 🎯 Resultado Final

El dashboard mostrará todos los cultivos configurados:
- **Soja**: 10.5 hectáreas
- **Maíz**: 8.5 hectáreas
- **Trigo**: 12.0 hectáreas
- **Girasol**: 15.0 hectáreas
- **Total**: 46.0 hectáreas

## 🔧 Cómo Agregar Nuevos Cultivos

1. **Crear nuevo script SQL** en el directorio `sql/` con prefijo numérico (ej: `05_add_cebada_crop.sql`)
2. **El sistema lo detectará automáticamente** y lo ejecutará en el orden correcto
3. **No es necesario modificar el código Go** - solo agregar el archivo SQL

### Ejemplo de Nuevo Script
```sql
-- 05_add_cebada_crop.sql
INSERT INTO crops (name, created_at, updated_at) VALUES 
('Cebada', NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

INSERT INTO fields (name, project_id, lease_type_id, created_at, updated_at) VALUES 
('Campo Central', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO lots (name, field_id, hectares, previous_crop_id, current_crop_id, season, created_at, updated_at) VALUES 
('Lote E1', 5, 9.0, 5, 5, '2024-2025', NOW(), NOW())
ON CONFLICT DO NOTHING;
```

## 🎯 Endpoint de Prueba

Después de ejecutar el seed, prueba el dashboard:
```bash
curl -H "X-USER-ID: 1" -H "X-API-KEY: abc123secreta" \
     http://localhost:8080/api/v1/dashboard
```

## 📝 Notas Importantes

- **Orden de Ejecución**: Los scripts se ejecutan en orden alfabético, por eso usamos prefijos numéricos
- **Dependencias**: Cada script asume que los datos básicos ya existen
- **Reseteo**: El flag `-reset` limpia todas las tablas antes de ejecutar los scripts
- **Flexibilidad**: El sistema es completamente genérico y no necesita modificaciones para nuevos cultivos

## 🚀 Ventajas del Sistema Genérico

1. **Mantenimiento Simple**: Solo agregar archivos SQL, no modificar código Go
2. **Escalabilidad**: Fácil agregar nuevos cultivos sin tocar la lógica
3. **Consistencia**: Todos los scripts siguen el mismo patrón
4. **Orden Controlado**: Los prefijos numéricos garantizan el orden correcto
5. **Reutilización**: El mismo sistema puede usarse para otros tipos de datos