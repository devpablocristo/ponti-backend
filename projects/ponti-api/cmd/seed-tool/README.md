# 🌱 SEED TOOL

Herramienta independiente para poblar la base de datos con datos de prueba usando archivos SQL puros, completamente desacoplada de la API principal.

## 🎯 **Propósito**

Este tool te permite:
- **Ejecutar seeds de forma controlada** - solo lo que necesites
- **Limpiar datos** antes de sembrar
- **Ejecutar seeds en orden correcto** respetando dependencias
- **Debuggear problemas** ejecutando seeds individuales
- **Evitar conflictos** con la API principal
- **Simplicidad máxima** - solo archivos SQL puros
- **Sin dependencias complejas** - solo PostgreSQL driver
- **Fácil mantenimiento** - editas SQL directamente

## 🚀 **Uso Básico**

### **Ejecutar todos los seeds:**
```bash
go run cmd/seed-tool/main.go -all
```

### **Limpiar y ejecutar todos los seeds:**
```bash
go run cmd/seed-tool/main.go -reset -all
```

### **Ejecutar seed específico:**
```bash
go run cmd/seed-tool/main.go -seed users
go run cmd/seed-tool/main.go -seed customers
go run cmd/seed-tool/main.go -seed crops
```

### **Ver ayuda:**
```bash
go run cmd/seed-tool/main.go -help
```

## 📋 **Orden de Ejecución Automático**

Cuando usas `-all`, los archivos SQL se ejecutan en orden alfabético:

1. **00_reset.sql** - Limpieza de base de datos (si se especifica reset)
2. **01_users.sql** - Usuarios básicos
3. **02_types.sql** - Tipos del sistema
4. **03_units.sql** - Unidades de medida
5. **04_customers.sql** - Clientes
6. **05_campaigns.sql** - Campañas
7. **06_crops.sql** - Cultivos
8. **07_projects.sql** - Proyectos
9. **08_lots.sql** - Lotes (incluye campos)

**Nota:** Los archivos se ejecutan en orden alfabético, por eso usamos numeración con prefijos.

## 🎛️ **Opciones Disponibles**

| Flag | Descripción | Ejemplo |
|------|-------------|---------|
| `-all` | Ejecutar todos los seeds en orden | `-all` |
| `-reset` | Limpiar todos los datos antes de sembrar | `-reset -all` |
| `-seed <nombre>` | Ejecutar seed específico | `-seed users` |

**Seeds disponibles:**
- `users` - Usuarios básicos
- `types` - Tipos del sistema
- `units` - Unidades de medida
- `customers` - Clientes
- `campaigns` - Campañas
- `crops` - Cultivos
- `projects` - Proyectos
- `lots` - Lotes (incluye campos)

## 🔧 **Configuración**

### **Variables de Entorno:**
```bash
export DB_HOST=localhost
export DB_USER=admin
export DB_PASSWORD=admin
export DB_NAME=ponti_api_db
export DB_PORT=5432
export DB_SSL_MODE=disable
```

### **Valores por Defecto:**
- `DB_HOST`: localhost
- `DB_USER`: admin
- `DB_PASSWORD`: admin
- `DB_NAME`: ponti_api_db
- `DB_PORT`: 5432
- `DB_SSL_MODE`: disable

## 📊 **Datos Generados**

### **Cantidades por defecto:**
- **2 usuarios** (seed@local, seed123@local)
- **10 clientes** (Customer 1-10)
- **10 campañas** (Campaign 1-10)
- **4 tipos** (Semilla, Agroquímicos, Fertilizantes, Labores)
- **3 unidades** (Lts, Kg, Ha)
- **10 cultivos** (Crop 1-10)
- **10 proyectos** (Project 1-10)
- **10 campos** (Field 1-10)
- **10 lotes** (Lot 1-10)

### **Datos especiales:**
- **Lot 1**: 50 toneladas, 5 hectáreas
- **Lotes 2-10**: 10, 15, 20... hectáreas respectivamente
- **Proyectos**: Asociados automáticamente a clientes y campañas existentes
- **Campos**: Creados automáticamente con los proyectos

## 🧹 **Limpieza de Datos**

### **Tablas limpiadas (en orden):**
1. `workorder_items`
2. `workorders`
3. `lot_dates`
4. `lots`
5. `fields`
6. `projects`
7. `supplies`
8. `labors`
9. `categories`
10. `types`
11. `units`
12. `class_types`
13. `crop_commercializations`
14. `crops`
15. `lease_types`
16. `managers`
17. `investors`
18. `campaigns`
19. `customers`
20. `users`

### **Secuencias reseteadas:**
Todas las secuencias de ID se resetean a 1 después de la limpieza.

## 🚨 **Casos de Uso Comunes**

### **1. Desarrollo inicial:**
```bash
go run cmd/seed-tool/main.go -reset -all
```

### **2. Testing específico:**
```bash
go run cmd/seed-tool/main.go -seed users -seed customers -seed crops -seed projects
```

### **3. Solo datos básicos:**
```bash
go run cmd/seed-tool/main.go -seed users -seed types -seed units -seed customers
```

### **5. Limpiar solo:**
```bash
go run cmd/seed-tool/main.go -reset
```

## 🔍 **Troubleshooting**

### **Error de conexión a BD:**
- Verifica que PostgreSQL esté corriendo
- Confirma las variables de entorno
- Usa `docker compose ps` para ver el estado de los contenedores

### **Error de dependencias:**
- Ejecuta `go mod tidy` en el directorio del seed-tool
- Verifica que los modelos estén disponibles

### **Error de permisos:**
- Asegúrate de que el usuario tenga permisos de escritura en la BD
- Verifica que las tablas existan

## 📁 **Estructura del Proyecto**

```
cmd/seed-tool/
├── main.go           # Punto de entrada principal
├── sql/              # Archivos SQL de seeds
│   ├── 00_reset.sql  # Limpieza de base de datos
│   ├── 01_users.sql  # Usuarios básicos
│   ├── 02_types.sql  # Tipos del sistema
│   ├── 03_units.sql  # Unidades de medida
│   ├── 04_customers.sql # Clientes
│   ├── 05_campaigns.sql # Campañas
│   ├── 06_crops.sql  # Cultivos
│   ├── 07_projects.sql # Proyectos
│   └── 08_lots.sql   # Lotes (incluye campos)
├── go.mod            # Dependencias del tool
├── Makefile          # Comandos make útiles
└── README.md         # Esta documentación
```

## 🎉 **Ventajas de esta Arquitectura**

1. **✅ Separación de responsabilidades** - Seeds separados de la API
2. **✅ Control granular** - Ejecutar solo lo que necesites
3. **✅ Orden automático** - Archivos SQL ejecutados en orden alfabético
4. **✅ Fácil debugging** - Seeds individuales para testing
5. **✅ Limpieza controlada** - Reset completo cuando sea necesario
6. **✅ Independencia** - No interfiere con la API principal
7. **✅ Simplicidad máxima** - Solo archivos SQL puros
8. **✅ Sin dependencias complejas** - Solo PostgreSQL driver
9. **✅ Fácil mantenimiento** - Editas SQL directamente
10. **✅ Reutilizable** - Puedes usar en CI/CD, testing, desarrollo

## 🚀 **Próximos Pasos**

1. **Ejecuta el tool** para verificar que funciona
2. **Personaliza los datos** editando los archivos SQL directamente
3. **Añade nuevos seeds** creando nuevos archivos SQL numerados
4. **Integra en tu workflow** de desarrollo/testing
5. **Usa el Makefile** para comandos rápidos y simples 