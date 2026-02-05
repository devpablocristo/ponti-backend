# Unificación DEV/STG: verificación final

**Estado:** ✅ Completado (2026-01-30)

## Resultado

| Recurso | Estado |
|---------|--------|
| Instancia única | `new-ponti-db-dev` (new-ponti-dev) |
| DB dev | `new_ponti_db_dev` |
| DB stg | `new_ponti_db_staging` |
| Instancia vieja | `new-ponti-db-stg` eliminada |
| ponti-backend STG | Usa instancia dev, DB staging |
| ponti-auth STG | Usa instancia dev, DB staging |

## Pendiente (opcional)

- [ ] `github-actions@new-ponti-stg` con `roles/cloudsql.admin` en new-ponti-dev (para refresh-golden-snapshot)
- [ ] Rotar password `app_stg` si fue expuesta

## Comandos útiles

```bash
# Verificar STG
curl -s $(gcloud run services describe ponti-backend --project=new-ponti-stg --region=us-central1 --format='value(status.url)')/ping

# Config Cloud Run STG
gcloud run services describe ponti-backend --project=new-ponti-stg --region=us-central1 --format="value(spec.template.metadata.annotations['run.googleapis.com/cloudsql-instances'])"
```
