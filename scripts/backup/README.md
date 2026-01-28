# 📸 Scripts de Snapshot del Schema

Scripts para crear backups del schema de la base de datos antes de hacer cambios.

---

## 🚀 Uso Rápido

### Snapshot Simple (Recomendado)

```bash
# Snapshot rápido con nombre automático
./scripts/backup/create_schema_snapshot_simple.sh

# Snapshot con nombre personalizado
./scripts/backup/create_schema_snapshot_simple.sh antes_fase_1
```

**Salida:** `snapshots/snapshot_YYYYMMDD_HHMMSS.sql`

---

### Snapshot Completo

```bash
# Snapshot completo con todos los detalles
./scripts/backup/create_schema_snapshot.sh antes_fase_1
```

**Salida:** `snapshots/antes_fase_1/` con múltiples archivos:
- `01_schema_completo.sql` - Schema completo
- `02_schemas_ssot.sql` - Solo schemas v4
- `03_vistas_v4.sql` - Solo vistas v4_report
- `04_funciones_ssot.sql` - Funciones v4
- `05_inventario_objetos.txt` - Lista de objetos
- `06_metadata.txt` - Metadata

---

## 📋 Qué incluye cada snapshot

### Snapshot Simple
- ✅ Schemas v4 (v4_core, v4_ssot, v4_calc, v4_report)
- ✅ Vistas v4_report
- ❌ Datos (solo estructura)
- ❌ Otros schemas

### Snapshot Completo
- ✅ Schema completo de la BD
- ✅ Schemas v4 separados
- ✅ Vistas v4_report separadas
- ✅ Funciones v4 separadas
- ✅ Inventario de objetos
- ✅ Metadata (versión, tamaño, etc.)
- ❌ Datos (solo estructura)

---

## 🔄 Restaurar Snapshot

### Desde snapshot simple

```bash
# Restaurar
docker compose -f docker-compose.yml exec -T ponti-db \
    psql -U admin -d ponti_api_db < snapshots/snapshot_YYYYMMDD_HHMMSS.sql
```

### Desde snapshot completo

```bash
# Restaurar solo schemas SSOT
docker compose -f docker-compose.yml exec -T ponti-db \
    psql -U admin -d ponti_api_db < snapshots/antes_fase_1/02_schemas_ssot.sql

# O restaurar todo
docker compose -f docker-compose.yml exec -T ponti-db \
    psql -U admin -d ponti_api_db < snapshots/antes_fase_1/01_schema_completo.sql
```

---

## ⚠️ Notas Importantes

1. **Solo estructura:** Los snapshots NO incluyen datos, solo definiciones de objetos
2. **Backup completo:** Para datos, usar `pg_dump` completo o backups de GCP
3. **Verificar antes de restaurar:** Asegurarse de que no haya conflictos
4. **Orden de restauración:** Restaurar schemas antes que vistas que los usan

---

## 📊 Cuándo crear snapshots

- ✅ Antes de Fase 1 (crear wrappers)
- ✅ Antes de Fase 2 (actualizar vistas)
- ✅ Antes de cualquier cambio importante en v4
- ✅ Antes de cualquier cambio importante

---

## 🗂️ Estructura de directorios

```
snapshots/
├── snapshot_20251222_065500.sql          # Snapshot simple
├── antes_fase_1/                          # Snapshot completo
│   ├── 01_schema_completo.sql
│   ├── 02_schemas_ssot.sql
│   ├── 03_vistas_v4.sql
│   ├── 04_funciones_ssot.sql
│   ├── 05_inventario_objetos.txt
│   ├── 06_metadata.txt
│   └── README.md
└── ...
```

---

**Última actualización:** 2025-12-22





