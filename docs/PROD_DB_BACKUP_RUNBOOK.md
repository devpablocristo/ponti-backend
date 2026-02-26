# PROD DB Backup Runbook

## Alcance
Backup y restauracion de la base de datos de produccion `new_ponti_db_prod`.

## Estado actual (aplicado)
- Frecuencia: diaria.
- Scheduler: `weekly-prod-db-export-only` (en `new-ponti-dev`, `us-central1`).
- Cron: `0 4 * * *` (04:00 UTC todos los dias).
- Job ejecutado: `prod-db-weekly-export` (Cloud Run Job).
- Instancia SQL: `new-ponti-db-dev` (instancia compartida).
- Base exportada: `new_ponti_db_prod`.
- Bucket destino: `gs://ponti-prod-only-backups-1087442197188/sql-exports/prod/`.
- Retencion automatica: 84 dias (12 semanas) sobre prefijo `sql-exports/prod/`.
- Alerta por fallo: policy `Ponti - Falla backup diario DB PROD`.
- Canal alerta: `Ponti PROD DB Backup Alerts` (email `devpablocristo@gmail.com`).

## Como ver los backups
```bash
gcloud storage ls gs://ponti-prod-only-backups-1087442197188/sql-exports/prod/ | sort
```

## Restore de emergencia (paso a paso)
1. Elegir backup:
```bash
gcloud storage ls gs://ponti-prod-only-backups-1087442197188/sql-exports/prod/ | sort
```
2. Crear DB temporal para validar restore:
```bash
gcloud sql databases create restore_drill_prod_YYYYMMDD \
  --instance=new-ponti-db-dev \
  --project=new-ponti-dev
```
3. Importar backup a DB temporal:
```bash
gcloud sql import sql new-ponti-db-dev \
  gs://ponti-prod-only-backups-1087442197188/sql-exports/prod/prod-YYYYMMDD-HHMMSS.sql.gz \
  --database=restore_drill_prod_YYYYMMDD \
  --project=new-ponti-dev \
  --quiet
```
4. Verificar estado de operacion:
```bash
gcloud sql operations list \
  --instance=new-ponti-db-dev \
  --project=new-ponti-dev \
  --limit=10 \
  --format='table(name,operationType,status,startTime,endTime,error.errors[0].message)'
```
5. Si hay incidente real, repetir import en la DB objetivo acordada por el equipo de release.

## Simulacro ejecutado
- Fecha: 2026-02-25.
- DB de prueba: `restore_drill_prod_20260225`.
- Resultado: import `DONE`.

## Backup de configuracion critica (Cloud Run, Secrets, recursos)
Se guarda snapshot en:
`gs://ponti-prod-only-backups-1087442197188/config-snapshots/<timestamp>/`

Incluye:
- servicios y jobs de Cloud Run (`dev/stg/prod`),
- secretos y versiones habilitadas (`dev/stg/prod`),
- instancias y bases SQL (`dev/stg/prod`),
- buckets y resumen de inventario.

## Validaciones operativas rapidas
```bash
# Scheduler diario habilitado
gcloud scheduler jobs describe weekly-prod-db-export-only \
  --project=new-ponti-dev --location=us-central1 \
  --format='value(schedule,state,timeZone)'

# Ultimas ejecuciones del job
gcloud run jobs executions list \
  --project=new-ponti-dev --region=us-central1 \
  --job=prod-db-weekly-export

# Alerta activa
gcloud monitoring policies list \
  --project=new-ponti-dev \
  --filter='displayName="Ponti - Falla backup diario DB PROD"'
```
