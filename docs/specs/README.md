# `docs/specs/` — Specs definitivos

Acá viven los **specs definitivos** del backend. Cada archivo es la fuente de verdad de una pieza
de trabajo, lista para implementar.

- `scripts/` — specs de scripts/tooling (ej. `scripts/reset-local-db-from-prod.md`).
- `features/` — specs de features recuperadas de `develop-problematico` (ver abajo).

**Convención de naming:** un archivo suelto por spec, sin número, sin sufijo: `docs/specs/<área>/<slug>.md`.

---

## Recuperación controlada de `develop-problematico`

En `develop-problematico` se hizo mucho trabajo que dio problemas y se revirtió. Ese trabajo está
documentado en `docs/features-develop-problematico/` (el **backlog**) y se recupera **de a una
feature por vez, a nivel spec, antes de implementar**.

### Cómo correrlo

En el chat de Claude Code, escribí:

```
/recuperar-feature <id-o-slug>
```

Ejemplos:

```
/recuperar-feature 001
/recuperar-feature crudar-archive-surface
```

### Qué hace (por cada corrida)

1. Encuentra la feature en `docs/features-develop-problematico/`.
2. **Re-baselinea contra el `develop` real** (solo lectura) para sacar el *diff de verdad* (no el que
   asume el spec viejo, que está desactualizado).
3. Escribe **un único** spec definitivo en `docs/specs/features/<slug>.md`
   (Propósito · Estado vs develop · Alcance/archivos · Migraciones · Dependencias · Plan de
   implementación · Validación · Riesgos).
4. **Borra** (`git rm`) la carpeta de esa feature del backlog → así `docs/features-develop-problematico/`
   muestra solo lo que falta tratar y `docs/specs/features/` lo ya tratado.

### Qué NO hace

- **No** escribe código de la app, **no** crea ramas, **no** corre migraciones, **no** mergea.
- La **implementación** es una etapa posterior, en la rama de trabajo sincronizada con `develop`.

> Definición del comando: [`.claude/commands/recuperar-feature.md`](../../.claude/commands/recuperar-feature.md).
