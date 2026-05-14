# Versionado y trazabilidad de deploys

## Objetivo

Los workflows de deploy registran siempre el SHA real efectivamente desplegado. La fuente de verdad es `DEPLOY_SHA`, resuelto inmediatamente despues del checkout con:

```bash
git rev-parse HEAD
```

Esto evita que un `workflow_dispatch` con branch, tag o SHA custom registre el SHA del evento de GitHub en lugar del commit realmente desplegado.

## Que queda trazado

- Imagen Docker inmutable taggeada por SHA.
- Metadata expuesta por `/api/v1/version`.
- GitHub Deployments para `dev`, `staging`, `staging-approved` y `prod`.
- Version humana del servicio, por ejemplo `dev-<short_sha>`, `stg-<short_sha>` o `prod-<short_sha>`.
- Revision de Cloud Run y version de migracion cuando aplica.

## Verificar que ambiente esta desplegado

Consultar `/api/v1/version` del ambiente y mirar:

- `service.git_sha`: commit real desplegado.
- `service.version`: version operacional del deploy.
- `service.build_time`: timestamp UTC del deploy.

El SHA tambien debe coincidir con:

- tag de la imagen de Cloud Run;
- payload del GitHub Deployment;
- commit aprobado si el ambiente fue promovido.

## Flujo normal

1. DEV se despliega desde `develop`.
2. STAGING se despliega desde `main` o por `workflow_dispatch` con `ref` custom.
3. Si STAGING queda bien, ejecutar `approve-staging.yml`.
4. PROD se promueve con `deploy-prod.yml`, pasando el `staging_sha` aprobado.

Ejemplo de deploy manual a STAGING:

```bash
CUSTOM_SHA="$(git rev-parse HEAD)"
gh workflow run deploy-staging.yml --ref main -f ref="$CUSTOM_SHA"
```

## Rollback de STAGING

Si STAGING queda con bugs, usar `rollback-staging.yml`.

Sin `target_sha`, restaura el ultimo candidato aprobado anterior:

```bash
gh workflow run rollback-staging.yml \
  --ref main \
  -f reason="rollback por bug en staging"
```

Con un SHA especifico:

```bash
gh workflow run rollback-staging.yml \
  --ref main \
  -f target_sha="<sha_aprobado>" \
  -f reason="rollback a version estable"
```

El backend incluye guard de compatibilidad de migraciones. Si el rollback cambia la version de migracion registrada, el workflow falla salvo que se confirme explicitamente:

```bash
gh workflow run rollback-staging.yml \
  --ref main \
  -f target_sha="<sha_aprobado>" \
  -f confirm_migration_compatible=true \
  -f reason="rollback compatible con schema actual"
```

Usar esa confirmacion solo si el cambio de schema es compatible con la app anterior.

## Rollback de PROD

Para PROD usar `rollback-prod.yml`. Es el mismo concepto, pero impacta usuarios reales:

```bash
gh workflow run rollback-prod.yml \
  --ref main \
  -f reason="rollback prod por incidente"
```

Si se requiere un SHA puntual:

```bash
gh workflow run rollback-prod.yml \
  --ref main \
  -f target_sha="<sha_aprobado>" \
  -f reason="rollback prod a version estable"
```

## Ventajas operativas

- Rollback confiable: se restaura una imagen/commit conocido, no un tag mutable.
- Auditoria clara: GitHub Deployments muestra que se desplego, aprobo, promovio o rollbackeo.
- Debugging mas rapido: logs, `/version`, imagen y codigo fuente apuntan al mismo commit.
- Promocion segura: PROD valida que el SHA promovido coincide con lo desplegado en STAGING.
- Deploy manual seguro: un `workflow_dispatch` con ref custom registra el commit checkouteado, no `github.sha`.

## Limitaciones

- No corrige registros historicos previos al fix.
- No reemplaza la validacion funcional; solo asegura trazabilidad correcta.
- En backend, un rollback de app no revierte migraciones de base de datos.
