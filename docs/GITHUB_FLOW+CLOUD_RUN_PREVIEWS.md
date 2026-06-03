# GitHub Flow + Cloud Run Previews

## Flujo

```
feature/*
    ↓
PR  ──────────────▶  Preview Environment (FE + BE + mobile)  ◀── QA valida el FEATURE (gate principal, pre-merge)
    ↓
merge a main  ─────▶  se destruye el preview
    ↓
staging  ──────────────────────────────────────────────────  ◀── solo smoke/integración (NO se cazan bugs de feature)
    ↓
prod
```

## Preview Environments

- El preview **se crea al abrir el PR** y se actualiza en cada push.
- Es un entorno completo: **FE + BE + mobile**.
- **Se destruye al cerrar el PR** (sea por merge o por cierre sin merge).
- Sin intervención manual.

### Aislamiento de datos

- El preview tiene su **propia instancia de base de datos**, **separada de staging y de producción**.
- Así, ningún preview puede afectar (conexiones, contención ni datos) a staging/prod.
- Cada PR usa una base propia dentro de esa instancia, creada al abrir y borrada al cerrar.

> Nota: el "preview" de mobile no es un servicio Cloud Run, sino un **build/artifact**
> (Expo/EAS, APK o TestFlight) que apunta al BE del preview.

## QA — dónde valida

Se valida en los dos lados, pero **con propósitos distintos**:

1. **Preview (por PR, pre-merge) — gate principal.**
   QA valida el **feature** en la URL del preview (FE+BE aislados). **Solo lo aprobado se mergea.**
   Esto es lo que mantiene `main` y `staging` **limpios**.

2. **Staging (post-merge) — NO se cazan bugs de feature.**
   Es solo **smoke / integración del conjunto** + release candidate. Llega limpio porque cada PR
   ya pasó por su preview. Si queda OK → `approve-staging` (`QA_APPROVED`) → se promueve a prod.

> ⚠️ Clave: si QA recién validara features **en staging**, recrearíamos el problema de `develop`
> (5 PRs, 2 malos → staging sucio). Por eso el peso de la validación está **antes del merge**, en el preview.
> Lo que aparezca en staging debe ser de **integración** (raro y puntual), no de un feature suelto.

## Monorepo (mejora futura)

Este sistema funcionaría **mejor con un monorepo**: `be + fe + mobile` en 1 repo.

- 1 PR = 1 feature que cruza FE/BE/mobile → un solo preview, sin emparejar repos.
- Cambios atómicos, sin desfasaje de versiones entre repos.
- Cleanup único al cerrar el PR.
- Menos duplicación de CI/CD (hoy cada repo repite workflows).

## Versionado y rollback

- El **versionado por SHA ya se aplica**.
- Sirve para identificar **qué versión está corriendo** y **qué código la creó**.
- Permite **volver atrás** ante un problema, tanto en **código** como en **artifact** (imagen Docker inmutable taggeada por SHA).
- **Pendiente:** completar el **versionado legible por humanos** (SemVer, ej. `v1.1`).
  Ya existe parcialmente vía `release.yml` (tags `vX.Y.Z` + GitHub Release); falta usarlo de forma consistente.
