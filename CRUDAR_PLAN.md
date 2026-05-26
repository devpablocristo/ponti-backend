# Homogeneización Integral del Frontend Ponti — CRUDAR universal

---

## ESTADO ACTUALIZADO — 2026-05-19 (rama `new-cns3`)

Después de los merges con `develop`, recuperación de commits y PRs #92/#99: el plan original avanzó significativamente. Update conservando estructura original; ver § "Fase X — Estado real" al final de cada fase.

### Resumen ejecutivo del avance

| Fase | Original | Estado real | Pendiente |
|---|---|---|---|
| **A** — Extraer primitivas | 9 componentes + 3 hooks | ✅ **>90% completa** | `PageHeader`, `RowActions`, `DependencyError` |
| **B** — Reemplazar duplicación inline | 70+ reemplazos | ⚠️ **Parcial** | 4 `LoaderCircle` residuales en lots/stock/reports |
| **C** — Listas activas con kebab | 6 listas | ⚠️ **Parcial** | Sin kebab universal; `RowActions` no existe |
| **D** — Drawers faltantes | 5 entidades | ✅ **3 + 1 extra completos** (Investors, Managers, Campaigns, Actors) | CustomerFormDrawer, ProjectsList |
| **E** — Bulk actions | 18 listas | ⚠️ **Primitiva existe** (`BulkActionBar`, `useBulkSelection`, `makeSelectColumn`) | Aplicación a 18 listas pendiente |
| **F** — UX polish | sidebar, breadcrumbs, etc | ❌ **Pendiente** | Todo |
| **G** — Refactor 15 hooks → `useEntityCrud` | Opt-in | ⚠️ **Parcial** | Investors/Managers/Campaigns/Fields ya migrados; resto pendiente |

### Inventario verificado (2026-05-19)

**Primitivas Fase A — EXISTEN:**
- ✅ `components/feedback/{LoadingOverlay,ErrorBanner,SuccessBanner,EmptyState,WarningBanner,InlineSpinner,DismissButton}.tsx`
- ✅ `components/crud/{EntityFormDrawer,ConfirmModal,BulkActionBar,ArchivedDrawer,BulkSelectionPanel,makeSelectColumn}.tsx`
- ✅ `components/ArchivedListPage/ArchivedListPage.tsx` + `hooks/useArchiveActions/`
- ✅ `hooks/{useConfirmDialog,useBulkSelection,useEntityCrud,useEntityFormDrawer}/`

**Primitivas Fase A — FALTAN:**
- ❌ `components/layout/PageHeader.tsx` (carpeta vacía)
- ❌ `components/crud/RowActions.tsx` (kebab menu)
- ❌ `components/crud/DependencyError.tsx` (mensaje de bloqueo por dependencias)

**Drawers Fase D — EXISTEN:**
- ✅ `InvestorFormDrawer.tsx`
- ✅ `ManagerFormDrawer.tsx`
- ✅ `CampaignFormDrawer.tsx`
- ✅ `ActorFormDrawer.tsx` (entidad nueva no prevista en plan original)

**Drawers Fase D — FALTAN:**
- ❌ `CustomerFormDrawer.tsx` (Customer sigue editándose vía `CustomerEditor.tsx` modal)
- ❌ `ProjectsList.tsx` / drawer propio (Projects solo tiene pantalla `/archived`)

### Estado CRUDAR por entidad (FE) — confirmado por audit

| Entidad | Create | Read (list) | Update | Archive | Restore | Hard Delete | Drawer | Lista activa | Notas |
|---|---|---|---|---|---|---|---|---|---|
| **customers** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ (usa Editor) | ✅ | `useCustomers` extendido en esta sesión |
| **projects** | ⚠️ (solo via hook) | ❌ | ❌ | ✅ | ✅ | ✅ | ❌ | ❌ | Solo `/archived` |
| **lots** | ⚠️ | ❌ database/ | ⚠️ | ✅ | ✅ | ✅ | ❌ | `/admin/lots` separado | Sin List en `/database/lots/` |
| **work-orders** | ❌ | ❌ | ❌ | ✅ | ✅ | ✅ | ❌ | ❌ database/ | Solo `/archived` |
| **labors** | ⚠️ | ⚠️ (en `/tasks`) | ⚠️ | ❌ visible | ❌ visible | ❌ visible | ❌ | `/database/tasks/` | Ambigüedad labors vs tasks |
| **supplies** | ⚠️ (via project) | ❌ database/ | ⚠️ | ✅ | ✅ | ✅ | ❌ | `/admin/products` | Sin List en `/database/supplies/` |
| **fields** | ❌ | ❌ | ❌ | ✅ | ✅ | ✅ | ❌ | ❌ | `useEntityCrud` genérico sin create/update |
| **investors** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | **CRUDAR COMPLETO** |
| **managers** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | **CRUDAR COMPLETO** |
| **campaigns** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | **CRUDAR COMPLETO** |
| **actors** | ✅ | ✅ | ⚠️ | ✅ | ❌ | ❌ | ✅ | ✅ + duplicates | Entidad nueva |

### Estado CRUDAR (BE) — confirmado por audit

✅ **Todas las 10 entidades** (customers, projects, lots, work-orders, labors, supplies, fields, investors, managers, campaigns) tienen CRUDAR HTTP completo: `POST /entity`, `GET /entity`, `GET /entity/:id`, `PUT /entity/:id`, `POST /entity/:id/archive`, `POST /entity/:id/restore`, `GET /entity/archived`, `DELETE /entity/:id/hard`.

✅ Multi-tenant uniforme vía `authz.MaybeTenantScope` en todas. Soft delete con `gorm.DeletedAt` + base model con `CreatedBy/UpdatedBy/DeletedBy`.

⚠️ Asimetría en `labor`: dual routing `/projects/:id/labors` (scope proyecto) vs `/labors` (global). Archive/restore solo bajo global.

### Trabajo prioritario derivado del estado real

Reordenado vs roadmap original. El "hito mínimo shippeable" del plan (A+B+F, ~6 días) ya está mayormente completo en Fase A. Lo restante prioritario:

1. **Cerrar Fase A**: implementar `RowActions` (kebab), `PageHeader`, `DependencyError`. ~1 día.
2. **Terminar Fase B**: migrar 4 `LoaderCircle` residuales (`pages/admin/lots/components/EditableTonsCell.tsx`, `pages/admin/reports/SummaryResultsReport.tsx`, `pages/admin/stock/Stock.tsx`, `pages/admin/stock/CreateStockItem.tsx`) a `LoadingOverlay` o `InlineSpinner`. ~0.5 día.
3. **Fase D parcial**: `CustomerFormDrawer` + lista activa propia de Customers (separada del giant form). `ProjectsList` propia. ~2 días.
4. **Fase C**: aplicar kebab `RowActions` a las listas existentes una vez exista la primitiva. ~3 días.
5. **Fase E**: aplicar bulk actions (primitiva ya lista). ~2 días.
6. **Fase F + G**: UX polish + completar migración de hooks. ~3 días.

**Total restante estimado: ~11 días** (vs 16 días originales — diferencia = trabajo ya hecho en Fases A, D parcial, G parcial).

---

## Contexto

El proyecto Ponti acaba de completar la estandarización backend de Archive/Restore/HardDelete para 9 dominios archivables y agregó las páginas FE de "Archivados" para todos. Eso está committeado y funciona.

**Pero la UI activa quedó desigual.** Una auditoría comparativa lado a lado de 8 páginas de lista activa, 8 drawers/forms, y todos los patrones de feedback al usuario reveló algo importante:

> **La fragmentación NO es por diseños distintos**. Los patrones ya convergen — el problema es que están **duplicados inline en cada página** en vez de extraídos a un componente único. La solución NO es rediseñar nada nuevo: es **extraer lo que ya está bien hecho repetido y aplicarlo en todos lados**.

Hallazgos concretos de la auditoría:

| Patrón | Instancias | ¿Coinciden? |
|---|---|---|
| `<LoaderCircle className="w-10 h-10 text-blue-600 animate-spin">` con overlay `bg-white bg-opacity-70 backdrop-blur-sm` | 6 páginas + 5 drawers = 11 | **Idénticos byte-a-byte** |
| Drawer con estructura `flex flex-col h-full` + header h2 + form `space-y-4 flex-1` + footer `flex justify-end gap-2 mt-auto pt-6` | 5 (LotDrawer, CreateItem, CreateOrder, UpdateOrder, CreateStockItem) | **Idéntica estructura** |
| Banner error inline `bg-red-50 text-red-800` con SVG icon + dismiss | 20+ | **Idéntica estructura** |
| Banner success inline `bg-green-50 text-green-800` con SVG icon + dismiss | 16+ | **Idéntica estructura** |
| `<BaseModal>` para confirmaciones | 13 archivos | **Mismo componente, mismo uso** |
| `<FilterBar>` + `useWorkspaceFilters` | 7 de 8 páginas | **Mismo componente** |
| `DataTable` de `@devpablocristo/modules-ui-data-display` | 8 de 8 páginas con tabla | **Mismo componente** |
| `<InputField>`, `<SelectField>`, `<Search>` | universal | **Ya unificados** |

**Lo que SÍ está fragmentado** (3 puntos puntuales, no diseños divergentes):

1. **3 usos de `window.confirm`** sobreviven en [Items.tsx](ui/src/pages/admin/database/products/Items.tsx) (×2) y [Products.tsx](ui/src/pages/admin/products/Products.tsx) — todos los demás ya usan `BaseModal`.
2. **`/lib/toast.ts` con sonner ya instalado y wrappeado** pero NUNCA se importa. 21 páginas reimplementan el banner verde inline en vez de usar la librería.
3. **No existe componente `<EmptyState>`** — 8 páginas pasan un `message="No hay X disponibles"` plano al DataTable.

**Lo que falta como primitiva** (existe el patrón mental pero ningún componente):

4. **No existe kebab menu de row actions** — todos usan callbacks `onEdit`/`onDelete` con iconos sueltos o handlers a nivel de tabla. Bueno como base, pero falta uniformar el dropdown ⋮.
5. **No existe drawer genérico extraído** — los 5 drawers existentes son copias casi byte-a-byte de la misma estructura, pero cada uno la implementa en su propio archivo.
6. **No existe vista activa propia** para Investor / Manager / Campaign — esas 3 entidades sólo tienen "archivados". Customer tampoco tiene Create/Edit propio (sólo embebido en form gigante de Project).

---

## Estrategia: extraer dominante → aplicar universal → completar gaps

**Principio rector:** no se inventa nada nuevo cuando ya hay un patrón ganador implícito. Se extrae el patrón dominante a una primitiva, se reemplazan todas las copias inline por la primitiva, y sólo entonces se llenan los gaps faltantes (con la misma primitiva, no con variantes nuevas).

Tres tipos de cambio, en este orden:

1. **EXTRAER**: lo que ya existe duplicado idéntico → 1 componente reusable.
2. **MIGRAR**: los 3 stragglers (window.confirm, banners-en-vez-de-toast, etc) al patrón dominante.
3. **COMPLETAR**: lo que falta para que las 9 entidades tengan CRUDAR completo, usando las mismas primitivas extraídas — no creando primitivas paralelas.

**Convención CRUDAR (locked):**
- `archive` = soft delete (reversible)
- `restore` = restaurar archivado
- `delete` = hard delete (irreversible, BE bloquea con 409 si hay dependientes)
- Prohibido: `deactivate`, `disable`, `remove`, `hidden`, `trash`, `destroy`, `erase`

**Decisiones acordadas con el usuario:**
1. Forms nuevos (Customer/Investor/Manager/Campaign) van como **drawer lateral** (mismo `<Drawer>` de material-tailwind ya en uso 5×).
2. Acciones por fila van como **kebab menu (⋮)** universal (estándar Linear/GitHub).
3. **Bulk actions IN SCOPE**: multi-select + barra de acciones masivas.
4. **OUT OF SCOPE**: Optimistic UI, Cmd+K, toggle archivados en lista activa.
5. **No introducir librerías nuevas**: Tailwind + lucide + sonner + axios + DataTable + FilterBar existentes alcanzan.
6. **No refactor del giant Customers.tsx (1142 LOC)** — deuda técnica explícita, fuera de scope.

---

## Inventario: lo que YA EXISTE y se reusa tal cual

| Recurso | Ruta | Decisión |
|---|---|---|
| `<Drawer>` (material-tailwind) | dependencia | **Reusar en 100% de forms** (ya 5×) |
| `<FilterBar>` (@devpablocristo/modules-ui-filters) | dependencia | **Reusar en 100% de listas** (ya 7/8) |
| `DataTable` (@devpablocristo/modules-ui-data-display) | dependencia | **Reusar en 100% de tablas** (ya 8/8) |
| `<BaseModal>` | [components/Modal/BaseModal.tsx](ui/src/components/Modal/BaseModal.tsx) | **Reusar para todas las confirmaciones** (ya 13×) |
| `<Button>` | [components/Button/Button.tsx](ui/src/components/Button/Button.tsx) | Variantes completas, **reusar siempre** |
| `<InputField>`, `<SelectField>`, `<Search>`, `<TextAreaField>` | [components/Input/](ui/src/components/Input/) | **Reusar siempre** |
| `apiClient` (axios + auth) | [api/client.ts](ui/src/api/client.ts) | **Único cliente HTTP** |
| `toast` (wrapper sonner) | [lib/toast.ts](ui/src/lib/toast.ts) | **Activar uso universal** (hoy 0 importaciones) |
| `useWorkspaceFilters` | hooks | **Reusar en todas las listas** |
| `<ArchivedListPage<T>>` | [components/ArchivedListPage/](ui/src/components/ArchivedListPage/ArchivedListPage.tsx) | **Reusar tal cual** (ya 9×) |
| `useArchiveActions<T>` | [hooks/useArchiveActions/](ui/src/hooks/useArchiveActions/index.ts) | **Reusar tal cual** (ya 9×) |
| `getArchiveCopy/getRestoreCopy/getHardDeleteCopy` | [components/Modal/copy.ts](ui/src/components/Modal/copy.ts) | **Extender con create/update/dependency copy** |
| `<IndicatorCard>` | components | **Reusar para KPIs en headers** |
| Hooks `useX` por entidad (15) | hooks/useX/ | Mantener forma actual; sólo agregar `createX`/`updateX` donde falten |

---

## Diseño objetivo: el "Sistema CRUD UI" universal

### 1. Una plantilla por tipo de página

**A. Lista activa** — todas las páginas que listan registros activos con tabla:

```
┌─────────────────────────────────────────────┐
│  PageHeader: title + [+ Nuevo X] (Button)   │  ← componente
├─────────────────────────────────────────────┤
│  FilterBar (existente)                      │  ← reusar
├─────────────────────────────────────────────┤
│  IndicatorCards (opcional, ya existe)       │  ← reusar
├─────────────────────────────────────────────┤
│  BulkActionBar (sticky cuando hay selección)│  ← NUEVO
├─────────────────────────────────────────────┤
│  DataTable                                  │  ← reusar
│   ├─ checkbox | data... | ⋮ (RowActions)   │  ← RowActions NUEVO
│   └─ EmptyState si data=[] | LoadingOverlay│  ← NUEVOS extracciones
└─────────────────────────────────────────────┘
```

**B. Lista archivada** — `<ArchivedListPage<T>>` ya existe, no se toca. Sólo se le pueden agregar bulk actions cuando se agregue selección global.

**C. Drawer de Create/Edit** — `<EntityFormDrawer>` extraído del patrón ya 5× repetido:

```
<Drawer open onClose maxWidth="max-w-xl">
  <div className="flex flex-col h-full p-6 relative">
    {processing && <LoadingOverlay />}        ← extraída
    <header>
      <h2>Title (Create | Edit)</h2>
      <button onClick={onClose}>×</button>
    </header>
    <form className="space-y-4 flex-1 overflow-auto">
      {error && <ErrorBanner ... />}           ← extraída
      {success && <SuccessBanner ... />}       ← extraída
      {children /* campos del form */}
    </form>
    <footer className="flex justify-end gap-2 mt-auto pt-6">
      <Button variant="secondary" onClick={onClose}>Cancelar</Button>
      <Button variant="primary" onClick={handleSubmit} disabled={processing}>Guardar</Button>
    </footer>
  </div>
</Drawer>
```

**D. Confirm modal** — `<BaseModal>` ya está bien. Sólo se agrega un wrapper `<ConfirmModal>` con severity (info/warning/danger) que setea `primaryButtonColor` automáticamente, y un slot opcional para `<DependencyError>`.

### 2. Primitivas a EXTRAER (no diseñar — copiar lo que ya está repetido)

| Primitiva nueva | Origen (lo que se copia) | Reemplaza |
|---|---|---|
| `<LoadingOverlay>` | El JSX exacto que aparece 11× con `<LoaderCircle className="w-10 h-10 text-blue-600 animate-spin">` + backdrop blur | 11 copias inline + 12 variantes de tamaño |
| `<ErrorBanner>` | El JSX exacto que aparece 20× con `bg-red-50 text-red-800` + SVG + dismiss | 20+ copias inline |
| `<SuccessBanner>` | El JSX exacto que aparece 16× con `bg-green-50` + SVG + dismiss | 16+ copias inline |
| `<EmptyState icon title description cta?>` | NO existe — diseñar 1 vez basado en el `message="No hay X"` actual del DataTable | 8+ "No hay X" planos |
| `<EntityFormDrawer>` | Estructura común de los 5 drawers existentes | 5 drawers reescriben mismo wrapper |
| `<RowActions actions={Action[]}>` | NO existe como kebab — basar en `<Button>` + `<BaseModal>`-like dropdown headless | Iconos sueltos + handlers ad-hoc |
| `<BulkActionBar selectedCount actions onClear>` | NO existe — diseñar 1 vez (sticky, animación slide-up) | — |
| `<ConfirmModal severity dependentsBlock?>` | Wrapper sobre `BaseModal` con `primaryButtonColor` derivado de severity | Cada uso de `BaseModal` reasigna color manualmente |
| `<PageHeader title actions?>` | Markup repetido en cada página (h1 + botones flex right) | Headers ad-hoc |

### 3. Hooks a EXTRAER

| Hook | Origen | Reemplaza |
|---|---|---|
| `useConfirmDialog()` | NO existe — provee `await confirm({...})` que monta `<ConfirmModal>` y resuelve | `window.confirm` (3 usos) + reimplementaciones de modal state (5+ páginas) |
| `useBulkSelection<T>(items)` | NO existe | — (nueva capacidad) |
| `useEntityCrud<T>` (opt-in) | Patrón duplicado en 15 hooks (useCustomers, useSupplies, etc) | Eventualmente las 15 implementaciones de useReducer + actions |

### 4. Convenciones globales

**Iconografía** (lucide-react, locked):
- Edit: `Pencil` — Archive: `Archive` — Restore: `RotateCcw` — Delete: `Trash2` — Kebab: `MoreVertical` — Loading: `Loader2` o `LoaderCircle` (consistente con uso actual)

**Colores semánticos** (Tailwind, ya en uso):
- Primary action: `blue-600` — Archive: `slate-700` — Restore: `green-700` — Delete: `red-700` (todos ya usados así)

**Naming en hooks de entidad** (regla):
```ts
useX = () => ({
  // datos
  data, total, processing, error,
  archivedData,
  // operaciones
  list, get, create, update, archive, restore, hardDelete, listArchived,
})
```
Hooks existentes mantienen sus nombres actuales hasta que se migren — la regla aplica a hooks nuevos y a renombres opt-in.

**Feedback al usuario** (regla universal):
- Operación exitosa rápida (save, archive, restore) → `toast.success(...)` (sonner)
- Operación exitosa con detalle persistente (export, generate report) → `<SuccessBanner>` inline
- Error técnico (red, parsing) → `toast.error(...)` o `<ErrorBanner>` si es contextual al form
- Error de validación o dependencia → `<ErrorBanner>` o `<DependencyError>` inline en el form/modal
- Confirmación destructiva → `<ConfirmModal severity="danger">`
- **0 usos de `window.confirm` o `alert()` permitidos** post-migración.

---

## Plan por fases

### Fase A — Extraer primitivas (foundations)

**Objetivo:** crear las primitivas reusables copiando lo que ya está repetido. Ningún cambio visual a páginas existentes; sólo nuevos archivos. El visual hash de las páginas no debe cambiar tras esta fase.

**A.1 — Extracciones de patrones idénticos** (copiar JSX existente):

1. **`<LoadingOverlay>`** en `ui/src/components/feedback/LoadingOverlay.tsx`:
   ```tsx
   <LoadingOverlay show? size="md" />
   // Renderiza el div absoluto con backdrop blur y LoaderCircle del mismo color/tamaño que hoy
   ```

2. **`<ErrorBanner>`** en `ui/src/components/feedback/ErrorBanner.tsx`:
   ```tsx
   <ErrorBanner message dismissible? onDismiss? />
   ```

3. **`<SuccessBanner>`** en `ui/src/components/feedback/SuccessBanner.tsx`: idem.

4. **`<EmptyState>`** en `ui/src/components/feedback/EmptyState.tsx`:
   ```tsx
   <EmptyState icon? title description? cta? />
   ```
   Diseño: ícono lucide centrado + título + descripción opcional + botón opcional.

5. **`<PageHeader>`** en `ui/src/components/layout/PageHeader.tsx`:
   ```tsx
   <PageHeader title subtitle? actions? />
   ```

**A.2 — Componentes nuevos derivados de patrones existentes:**

6. **`<EntityFormDrawer>`** en `ui/src/components/crud/EntityFormDrawer.tsx`:
   - Reusa `<Drawer>` de material-tailwind tal cual los 5 drawers existentes.
   - Header + body scrollable + footer con Cancelar/Guardar.
   - Slots: `children` para el form, `actions?` para botones extra (ej: "Duplicar" en UpdateOrder).
   - Internamente usa `<LoadingOverlay>`, `<ErrorBanner>`, `<SuccessBanner>` extraídos.
   - Cierre con Esc, click outside (delegado a `<Drawer>`).
   - **Validación primaria con esta extracción:** refactorizar **1 drawer existente** (LotDrawer, el más simple) para usar `<EntityFormDrawer>` y verificar diff visual = 0. Si hay regression, ajustar el componente. Sólo entonces continuar.

7. **`<RowActions>`** en `ui/src/components/crud/RowActions.tsx`:
   - Botón ⋮ (`MoreVertical`) que abre dropdown con acciones provistas.
   - Acciones: `{ label, icon, onClick, variant?, disabled?, divider? }[]`.
   - Posicionamiento con `floating-ui` si está en el bundle, si no — cálculo manual con detector de borde de viewport.
   - Cierra con Esc, click outside.

8. **`<ConfirmModal>`** en `ui/src/components/crud/ConfirmModal.tsx`:
   - Wrapper sobre `BaseModal` con `severity: "info" | "warning" | "danger"`.
   - Mapea severity → `primaryButtonColor` (rojo / amarillo / azul).
   - Slot opcional `<DependencyError>` cuando viene 409.
   - Soporta `requireTypeToConfirm?: string` (input que pide escribir el nombre del item para confirmar; deshabilita botón hasta que coincide). Sólo se activa para hard-delete con dependientes muy críticos.

9. **`<DependencyError>`** en `ui/src/components/crud/DependencyError.tsx`:
   - Recibe `entityLabel`, `itemLabel`, `dependents: { type: string; count: number; status: "active" | "archived" }[]`.
   - Renderiza mensaje localizado: "No se puede eliminar el cliente «Pérez SA» porque tiene 3 proyectos activos y 1 archivado. Archivá o eliminá esos primero."

10. **`<BulkActionBar>`** en `ui/src/components/crud/BulkActionBar.tsx`:
    - Sticky bar arriba de la tabla cuando `selectedCount > 0`.
    - Acciones: `{ label, icon, onClick, variant? }[]`.
    - Botón "Limpiar selección".
    - Animación slide-down con CSS transition (sin librería).

**A.3 — Hooks nuevos:**

11. **`useConfirmDialog()`** en `ui/src/hooks/useConfirmDialog/index.ts`:
    ```tsx
    const confirm = useConfirmDialog();
    const ok = await confirm({
      title, message, severity: "danger",
      primaryLabel: "Eliminar",
      requireTypeToConfirm?: itemName,
    });
    if (!ok) return;
    ```
    Internamente: provider en root + portal para el modal.

12. **`useBulkSelection<T>(items)`** en `ui/src/hooks/useBulkSelection/index.ts`:
    - Devuelve `{ selectedIds, isSelected, toggle, toggleAll, clear, selectedItems }`.

13. **`useEntityCrud<T>` (opt-in)** en `ui/src/hooks/useEntityCrud/index.ts`:
    - Factory genérica con la forma de hook estándar.
    - **No obligatorio adoptarlo** — los hooks viejos siguen funcionando hasta que se migren uno a uno (Fase F, deuda técnica).

**A.4 — Extender [copy.ts](ui/src/components/Modal/copy.ts):**

```ts
export const getCreateSuccessCopy = (entityLabel: string) => `Se creó ${entityLabel} correctamente.`;
export const getUpdateSuccessCopy = (entityLabel: string) => `Se actualizó ${entityLabel}.`;
export const getDependencyErrorCopy = (entityLabel: string, itemLabel: string, deps: Dep[]) => {
  // arma el string a partir de deps
};
```

**Verificación Fase A:**
- `tsc --noEmit` pasa.
- LotDrawer migrado a `<EntityFormDrawer>` se ve idéntico al actual (visual diff 0).
- Ninguna otra página tocada — ningún regression en producción.

---

### Fase B — Reemplazar duplicación inline (homogeneización masiva)

**Objetivo:** sustituir TODAS las copias inline de los patrones extraídos en Fase A por las primitivas. Ningún cambio funcional, solo find-and-replace asistido. Cero invención.

**B.1 — Reemplazar 11 spinners inline por `<LoadingOverlay>`:**
- `grep -rn "LoaderCircle" ui/src` → 28+ ocurrencias.
- En CADA página/drawer reemplazar el bloque overlay por `<LoadingOverlay show={processing} />`.
- Verificación: visual diff 0 en cada página.

**B.2 — Reemplazar 20+ banners de error inline por `<ErrorBanner>`:**
- `grep -rn "bg-red-50" ui/src/pages` → ~20 ocurrencias.
- Reemplazar JSX inline por `<ErrorBanner message={errorMessage} onDismiss={...} />`.

**B.3 — Reemplazar 16+ banners de success inline por `<SuccessBanner>` o `toast.success`:**
- Reglas:
  - Success rápido y operacional (save, archive, restore) → `toast.success(...)` desde [lib/toast.ts](ui/src/lib/toast.ts) — primer caso real de uso de la librería.
  - Success persistente con detalle (export, batch result) → `<SuccessBanner>` inline.
- `grep -rn "bg-green-50" ui/src/pages` → identificar cuál es cada caso.

**B.4 — Reemplazar 3 `window.confirm` por `useConfirmDialog`:**
- [Items.tsx:265](ui/src/pages/admin/database/products/Items.tsx#L265) — "Hay cambios sin guardar. ¿Desea salir?"
- [Items.tsx:1037](ui/src/pages/admin/database/products/Items.tsx#L1037) — destructiva.
- [Products.tsx:361](ui/src/pages/admin/products/Products.tsx#L361) — "¿Estás seguro de eliminar este movimiento?"
- Migrar a `await confirm({ title, message, severity: "danger" })`.

**B.5 — Reemplazar empty states planos por `<EmptyState>`:**
- 8 instancias de `message="No hay X disponibles"` → `<EmptyState icon={Inbox} title="No hay X" description="..." />`.
- DataTable acepta render custom para empty (verificar API).

**B.6 — Migrar los 5 drawers existentes a `<EntityFormDrawer>`:**
Orden por simplicidad:
1. **LotDrawer** (validación de la primitiva, ya hecha en Fase A).
2. **CreateStockItem**.
3. **CreateItem**.
4. **CreateOrder**.
5. **UpdateOrder**.

Cada migración: el drawer pasa de implementar su propio `<Drawer>` + header + form + footer a usar `<EntityFormDrawer>` con `children`. Los campos del form quedan iguales. Visual diff 0.

**Verificación Fase B:**
- `grep -r "window.confirm" ui/src` → vacío.
- `grep -r "<LoaderCircle" ui/src/pages` → vacío (solo dentro de `<LoadingOverlay>`).
- Visual smoke test por página: las pantallas se ven igual o mejor.
- `tsc --noEmit` pasa.
- Tests existentes pasan.
- Bundle size verificado (no debería crecer significativamente; idealmente se reduce por dedup).

---

### Fase C — Listas activas con kebab + bulk + acceso CRUDAR

**Objetivo:** las 9 entidades archivables tienen lista activa con `<RowActions>` y soporte de bulk select. Cada entidad accede a sus 6 operaciones CRUDAR desde la misma UI.

**Patrón estándar para lista activa:**

```tsx
export default function XListPage() {
  const { data, processing, error, list, archive, hardDelete } = useX();
  const { selectedIds, toggle, toggleAll, clear } = useBulkSelection(data);
  const confirm = useConfirmDialog();
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [editing, setEditing] = useState<X | null>(null);

  const handleArchive = async (item: X) => {
    const ok = await confirm(getArchiveCopy("el X", item.name));
    if (!ok) return;
    try {
      await archive(item.id);
      toast.success(`Se archivó "${item.name}"`);
      await list();
    } catch (err) {
      toast.error(err.message);
    }
  };

  const columns = [
    { key: "select", header: <Checkbox onChange={toggleAll} />, render: (item) => 
      <Checkbox checked={isSelected(item.id)} onChange={() => toggle(item.id)} /> },
    { key: "name", header: "Nombre" },
    // ... otras columnas de la entidad
    { key: "actions", header: "", render: (item) => (
      <RowActions actions={[
        { label: "Editar", icon: Pencil, onClick: () => { setEditing(item); setDrawerOpen(true); } },
        { label: "Archivar", icon: Archive, onClick: () => handleArchive(item) },
        { divider: true },
        { label: "Eliminar", icon: Trash2, variant: "danger", onClick: () => handleHardDelete(item) },
      ]} />
    )},
  ];

  return (
    <>
      <PageHeader 
        title="Xs" 
        actions={<Button variant="primary" onClick={() => setDrawerOpen(true)}>+ Nuevo X</Button>} 
      />
      <FilterBar /> {/* existente */}
      {selectedIds.length > 0 && <BulkActionBar selectedCount={selectedIds.length} actions={...} onClear={clear} />}
      <LoadingOverlay show={processing} />
      {error && <ErrorBanner message={error} />}
      {data.length === 0 ? (
        <EmptyState icon={Inbox} title="No hay Xs" cta={...} />
      ) : (
        <DataTable data={data} columns={columns} />
      )}
      <XFormDrawer 
        open={drawerOpen} 
        onClose={() => { setDrawerOpen(false); setEditing(null); }} 
        item={editing} 
        onSaved={() => { list(); setDrawerOpen(false); }} 
      />
    </>
  );
}
```

**Migraciones a aplicar (orden por uso real):**

**C.1 — Lots** (`/admin/lots`): la más usada. Modificar [Lots.tsx](ui/src/pages/admin/lots/Lots.tsx) para agregar columna kebab con Edit/Archive/Delete + checkbox de selección. LotDrawer ya migrado en Fase B.

**C.2 — WorkOrders** (`/admin/work-orders`): mismo patrón. Mantener CreateOrder/UpdateOrder ya migrados.

**C.3 — Products/Supplies** (`/admin/products`): mismo patrón. CreateItem ya migrado.

**C.4 — Stock** (`/admin/stock`): no es archivable (transaccional) → solo Edit + Delete inline en kebab. Sin Archive/Restore. Bulk no aplica (es por celda).

**C.5 — Tasks/Labors** (`/admin/tasks`): Labors **no tiene** archive/restore en BE hoy. **Decisión**: no agregar kebab con archive hasta que BE lo implemente. Migrar lo demás (Edit, Delete inline) al kebab para coherencia visual.

**C.6 — Customers vista comercial** (`/admin/customers`): convertir [pages/admin/customers/Customers.tsx](ui/src/pages/admin/customers/Customers.tsx) (357 LOC) a usar el patrón. Ya tiene `onCopy` y `onDelete` callbacks — pasarlos al kebab. Agregar Archive/Restore.

**C.7 — Investor / Manager / Campaign**: hoy NO tienen lista activa propia. Se cubren en Fase D (junto con Create/Edit nuevo).

**C.8 — Customer / Project lista activa propia**: hoy se editan dentro del giant Customers.tsx form. Se cubren en Fase D.

**Verificación Fase C:**
- Por cada entidad migrada: archive desde kebab → toast → item desaparece de activos → aparece en `/archived`. Restore desde `/archived` → vuelve a activos.
- Visual diff: la página debe verse igual o mejor (kebab más limpio que iconos sueltos).
- Mobile (375px): kebab cabe en viewport, dropdown se abre arriba si está en bottom de pantalla.
- `tsc --noEmit` pasa.

---

### Fase D — Create/Edit drawers para entidades faltantes

**Objetivo:** Investor, Manager, Campaign, Customer pasan a tener Create/Edit drawer propio usando `<EntityFormDrawer>`. Cada uno con su lista activa propia.

**Verificación previa (bloqueante):**
- Endpoints BE de Create/Update existen para las 4 entidades (`POST /investors`, `PUT /investors/:id`, etc). Si alguno falta → escalar al usuario antes de empezar la sub-fase. **No inventar BE.**

**D.1 — Investors** (`/admin/database/investors`):
- Lista activa nueva: `InvestorsList.tsx` (patrón estándar Fase C).
- Drawer: `InvestorFormDrawer.tsx` usando `<EntityFormDrawer>`. Campos según modelo BE.
- Hook: extender [useInvestors](ui/src/hooks/useInvestors/index.ts) con `createInvestor`, `updateInvestor`.
- Router: registrar `/admin/database/investors`.
- Sidebar: agregar entry "Inversores" en grupo Database.

**D.2 — Managers** (`/admin/database/managers`):
- Validar BE primero (POST/PUT). Si OK, mismo patrón.
- Hoy `useManagers` solo provee dropdown options. Extender con CRUD completo.

**D.3 — Campaigns** (`/admin/database/campaigns`):
- Idem D.2.

**D.4 — Customers** (`/admin/database/customers/list`):
- Lista activa SOLO de customers (no proyectos).
- Drawer `CustomerFormDrawer.tsx` con campos sólo de Customer.
- El form gigante de Project sigue accesible en `/admin/database/customers` (renombrar entry sidebar a "Clientes y Proyectos (Form)" para diferenciar).
- Customer ahora se puede crear sin necesidad de iniciar un proyecto.

**D.5 — Projects** (`/admin/projects/list`):
- Lista activa simple de proyectos.
- Edit dispara navegación al form gigante existente (sin reescribir el giant).
- Archive/Delete desde kebab usando hooks existentes.

**D.6 — Fields**:
- **Decisión: NO crear drawer independiente.** Los fields tienen sentido sólo en contexto de un proyecto. Mantener creación dentro del form de Project.
- Vista archived independiente ya existe.

**Verificación Fase D:**
- Por entidad: crear → aparece en lista → editar desde kebab → cambio se persiste → archivar → desaparece → archived → restaurar → vuelve.
- Endpoints BE verificados pre-implementación.
- Sidebar muestra cada entidad con su entry coherente.

---

### Fase E — Bulk actions

**Objetivo:** las 9 listas activas + 9 listas archivadas soportan multi-select y operaciones masivas.

**E.1 — Activar BulkActionBar en cada lista:**
- Cada lista usa `useBulkSelection` (de Fase A).
- Header de tabla muestra checkbox "select all".
- Cada fila tiene checkbox.
- Cuando `selectedCount > 0`: aparece `<BulkActionBar>` con acciones según contexto:
  - Lista activa: "Archivar N", "Eliminar N" (con confirm que detalla cuántos).
  - Lista archivada: "Restaurar N", "Eliminar N definitivamente".

**E.2 — Implementación bulk:**
- FE-only: `Promise.allSettled(ids.map(id => archive(id)))`.
- Modal de resultado: muestra "Se archivaron 4 de 5. 1 falló: <DependencyError>".
- Limitar selección a 50 items para evitar floods (UX warning si selectAll >50).

**E.3 — Aplicar a las 18 listas:**
- 9 activas (incluyendo las nuevas de Fase D).
- 9 archivadas — modificar `<ArchivedListPage<T>>` para soportar bulk opcionalmente (prop `bulkActions?: BulkAction[]`). Backwards compatible: páginas que no pasan la prop no muestran checkboxes.

**Verificación Fase E:**
- Seleccionar 5 items → archivar → todos archivados.
- 5 archivados → hard-delete → 2 fallan por dependientes → modal con detalle.
- Cancelar selección con "Limpiar".
- Mobile: BulkActionBar legible.

---

### Fase F — UX polish

**F.1 — Sidebar collapsible "Archivados"**: agrupar las 9 entries en un sub-menú expandible (igual a "Informes" / "Base de Datos") en [Sidebar.tsx:563-573](ui/src/layout/Sidebar/Sidebar.tsx#L563). Persistir estado en localStorage.

**F.2 — URL stale a entidad archivada**: si el usuario navega a `/admin/database/customers/:id` y el customer está archivado → mostrar banner "Este cliente fue archivado el {fecha} por {user}" con botón "Restaurar". Implementar en el form gigante.

**F.3 — Breadcrumbs**: componente `<Breadcrumbs>` para navegación entre Cliente → Proyecto → Campo → Lote. Routes registradas explícitamente.

**F.4 — Convergencia naming en hooks (opt-in)**: empezar a renombrar `errorMessage`/`successMessage` → uso de `toast`; `processing`/`isSaving`/`isLoading` → `processing` único. No bloqueante.

**Verificación Fase F:**
- 0 usos de `window.confirm` (`grep -r "window.confirm" ui/src` vacío).
- 0 usos directos de `<LoaderCircle>` fuera de `<LoadingOverlay>`.
- Sidebar mobile/desktop: archivados colapsable, persiste estado.

---

### Fase G — Hooks refactor (opt-in, deuda técnica)

**Objetivo (no bloqueante):** migrar los 15 hooks con reducer pattern duplicado a `useEntityCrud<T>` extraído en Fase A. Uno a uno, **sin breaking changes** — la API pública del hook se mantiene durante la transición.

Orden sugerido: empezar por los nuevos creados en Fase D (useInvestors, useManagers, useCampaigns, useCustomers extendido), después los grandes (useSupplies, useLots).

---

## Archivos críticos

**Fase A (todo nuevo):**
- `ui/src/components/feedback/LoadingOverlay.tsx`
- `ui/src/components/feedback/ErrorBanner.tsx`
- `ui/src/components/feedback/SuccessBanner.tsx`
- `ui/src/components/feedback/EmptyState.tsx`
- `ui/src/components/layout/PageHeader.tsx`
- `ui/src/components/crud/EntityFormDrawer.tsx`
- `ui/src/components/crud/RowActions.tsx`
- `ui/src/components/crud/ConfirmModal.tsx`
- `ui/src/components/crud/DependencyError.tsx`
- `ui/src/components/crud/BulkActionBar.tsx`
- `ui/src/hooks/useConfirmDialog/index.ts`
- `ui/src/hooks/useBulkSelection/index.ts`
- `ui/src/hooks/useEntityCrud/index.ts` (opt-in)

**Fase A (modificar):**
- [components/Modal/copy.ts](ui/src/components/Modal/copy.ts) — agregar create/update/dependency copy.
- [pages/admin/lots/LotDrawer.tsx](ui/src/pages/admin/lots/LotDrawer.tsx) — caso piloto de migración a `<EntityFormDrawer>` (Fase A.2 verificación).

**Fase B (modificar masivo):**
- 11 archivos con `<LoaderCircle>` inline → `<LoadingOverlay>`.
- 20+ archivos con `bg-red-50` banners → `<ErrorBanner>`.
- 16+ archivos con `bg-green-50` banners → `<SuccessBanner>` o `toast.success`.
- 3 archivos con `window.confirm` → `useConfirmDialog`.
- 8+ archivos con empty state plano → `<EmptyState>`.
- 5 drawers (LotDrawer, CreateItem, CreateOrder, UpdateOrder, CreateStockItem) → `<EntityFormDrawer>`.

**Fase C (modificar listas activas):**
- [pages/admin/lots/Lots.tsx](ui/src/pages/admin/lots/Lots.tsx)
- [pages/admin/workorders/WorkOrders.tsx](ui/src/pages/admin/workorders/WorkOrders.tsx)
- [pages/admin/products/Products.tsx](ui/src/pages/admin/products/Products.tsx)
- [pages/admin/stock/Stock.tsx](ui/src/pages/admin/stock/Stock.tsx)
- [pages/admin/tasks/Tasks.tsx](ui/src/pages/admin/tasks/Tasks.tsx)
- [pages/admin/customers/Customers.tsx](ui/src/pages/admin/customers/Customers.tsx) (vista comercial, 357 LOC)

**Fase D (todo nuevo + extender hooks):**
- `ui/src/pages/admin/database/investors/InvestorsList.tsx`
- `ui/src/pages/admin/database/investors/InvestorFormDrawer.tsx`
- `ui/src/pages/admin/database/managers/ManagersList.tsx`
- `ui/src/pages/admin/database/managers/ManagerFormDrawer.tsx`
- `ui/src/pages/admin/database/campaigns/CampaignsList.tsx`
- `ui/src/pages/admin/database/campaigns/CampaignFormDrawer.tsx`
- `ui/src/pages/admin/database/customers/CustomerFormDrawer.tsx`
- `ui/src/pages/admin/customers/CustomersList.tsx` (lista activa propia, separada del giant)
- `ui/src/pages/admin/projects/ProjectsList.tsx`
- Extender hooks: useInvestors, useManagers, useCampaigns, useCustomers con `createX`/`updateX`.
- [router.tsx](ui/src/router.tsx) + [Sidebar.tsx](ui/src/layout/Sidebar/Sidebar.tsx).

**Fase E (modificar 18 listas):**
- 9 activas + 9 archivadas — agregar `useBulkSelection` y `<BulkActionBar>`.
- [components/ArchivedListPage/ArchivedListPage.tsx](ui/src/components/ArchivedListPage/ArchivedListPage.tsx) — extender con prop opcional `bulkActions`.

**Fuera de scope explícitamente:**
- [pages/admin/database/customers/Customers.tsx](ui/src/pages/admin/database/customers/Customers.tsx) (1142 LOC) — refactor mayor, deuda técnica conocida.
- Endpoints BE de bulk (se hace client-side con Promise.allSettled).
- Optimistic UI, Command palette (Cmd+K), toggle "Mostrar archivados" en lista activa.
- Reemplazar `<Drawer>` de material-tailwind por otra librería.
- Reemplazar `DataTable` o `<FilterBar>` externas.

---

## Roadmap priorizado

| Fase | Trabajo | Días | Riesgo | Hito |
|---|---|---|---|---|
| A | Extraer 9 primitivas + 3 hooks + extender copy.ts. Validar con LotDrawer migrado | 2 | Bajo (todo nuevo) | Foundations |
| B | Reemplazar duplicación inline (spinners, banners, window.confirm, empty state, 5 drawers) | 2.5 | Bajo (find-replace asistido, visual diff 0) | Homogeneización |
| C | Migrar 6 listas activas a kebab + accesos CRUDAR | 3 | Medio (toca páginas en uso) | Listas estandarizadas |
| D | Crear 5 listas+drawers nuevos (Investor/Manager/Campaign/Customer/Projects) | 3 | Medio (depende de BE endpoints) | CRUDAR completo en 9 entidades |
| E | Bulk actions en 18 listas | 2 | Bajo (FE-only, sin BE changes) | Acciones masivas |
| F | UX polish (sidebar collapsible, URL stale, breadcrumbs) | 1.5 | Bajo | Pulido final |
| G | Refactor 15 hooks a useEntityCrud (opt-in) | 2 | Bajo | Deuda técnica saldada |

**Total: 16 días.** Cada fase es mergeable y verificable por separado.

**Hito mínimo shippeable:** Fases A + B + F → ~6 días. Deja la UI 100% homogénea visualmente sin agregar capacidades nuevas. Fases C + D + E + G son extensiones que se pueden hacer en pulsos posteriores.

### Quick wins (≤1 día, paralelizables)

1. Sidebar archivados collapsible (F.1) — 1 hora.
2. Reemplazar 3 `window.confirm` (B.4) — 1 hora.
3. Extender [copy.ts](ui/src/components/Modal/copy.ts) (A.4) — 1 hora.
4. Activar `toast.success/error` reemplazando 16 banners de success por toast (B.3 parcial) — 2 horas.

### Medium

5. Crear 5 primitivas de feedback + LoadingOverlay (A.1) — 1 día.
6. Migrar LotDrawer a `<EntityFormDrawer>` (validación) — 0.5 días.
7. Migrar Lots/WorkOrders/Products a kebab menu (C.1-C.3) — 1 día.

### Heavy

8. Crear lista activa Investors + drawer + extender hook (D.1) — 1 día.
9. Bulk actions atravesando 18 listas (Fase E) — 2 días.

---

## Plan de testing

**Por cada fase:**
- `tsc --noEmit` exit 0.
- Visual smoke en navegador (golden path por entidad).
- Bundle size verificado (no debería crecer; idealmente bajar por dedup en Fase B).

**End-to-end (Playwright) — flujos representativos:**

```ts
test("Investor full CRUDAR cycle", async ({ page }) => {
  // ... login + workspace ...
  await page.goto("/admin/database/investors");
  await page.click("text=+ Nuevo Inversor");
  await page.fill("[name=name]", "Test Investor");
  await page.click("text=Guardar");
  await expect(page.locator("text=Test Investor")).toBeVisible();
  
  // Edit via kebab
  await page.click(`[data-testid=row-Test Investor] [data-testid=kebab]`);
  await page.click("text=Editar");
  await page.fill("[name=name]", "Test Investor 2");
  await page.click("text=Guardar");
  
  // Archive via kebab
  await page.click(`[data-testid=row-Test Investor 2] [data-testid=kebab]`);
  await page.click("text=Archivar");
  await page.click("text=Archivar"); // confirm modal
  await expect(page.locator("text=Test Investor 2")).not.toBeVisible();
  
  // Restore from archived
  await page.goto("/admin/database/investors/archived");
  // ... etc
});

test("Hard delete blocks on dependents", async ({ page }) => {
  // Crear customer + project. Intentar hard-delete customer.
  // Modal muestra DependencyError con detalle "1 proyecto activo".
});

test("Bulk archive 5 lots", async ({ page }) => {
  // selectAll → BulkActionBar visible → Archivar → confirm → todos archivados.
});
```

**Tests manuales obligatorios:**
- Mobile 375px: kebab dropdown cabe en viewport, drawer es full-screen.
- Keyboard: Esc cierra modales y drawers. Tab navega.
- Cliente archivado vía URL stale: banner Restore.
- 9 listas archivadas existentes: regression test, deben funcionar idéntico tras Fases A+B.

**Métricas de éxito final:**
- 9/9 entidades archivables tienen 6/6 operaciones CRUDAR accesibles desde UI.
- 0 usos de `window.confirm` en src.
- 0 usos directos de `<LoaderCircle>` fuera de `<LoadingOverlay>`.
- 0 banners `bg-red-50`/`bg-green-50` inline (todos via primitiva o toast).
- 0 `<Drawer>` directo en pages (todos via `<EntityFormDrawer>`).
- TypeScript `tsc --noEmit` clean.
- Bundle size delta ≤ 0% (dedup compensa primitivas nuevas).

---

## Riesgos y mitigaciones

| Riesgo | Probabilidad | Impacto | Mitigación |
|---|---|---|---|
| Migración de drawer rompe alguno de los 5 forms (LotDrawer, CreateItem, etc) | Media | Alto | Migrar uno por vez. LotDrawer primero como piloto en Fase A.2. Visual diff 0 obligatorio antes de continuar. |
| Endpoints BE de Create/Update faltantes para Manager/Campaign | Media | Bloqueante para Fase D | Verificar endpoints antes de comenzar D.2/D.3. Si falta, escalar al usuario. |
| `<RowActions>` dropdown se sale del viewport mobile | Media | Medio | Posicionamiento con detector de borde (cálculo manual) o `floating-ui` si está en bundle. Tests manuales 375px. |
| Bulk hard-delete con 50 items genera 50 requests al BE | Baja | Medio | Limitar selección a 50; rate-limit del lado FE con `p-limit` si fuera necesario; warning UX al usuario. |
| `useConfirmDialog` requiere provider en root pero no se monta | Baja | Alto | Agregar provider en [main.tsx](ui/src/main.tsx) en Fase A.3 desde el primer día. |
| Confusión sidebar entre "Clientes" (lista nueva) y "Crear Clientes y Sociedades" (form gigante) | Media | Medio | Renombrar entry vieja a "Form de Proyectos" para diferenciar. Documentar en sidebar tooltip. |
| Sidebar collapsible default-collapsed esconde archivados | Baja | Bajo | Default-expanded en primer load; persistir preferencia en localStorage. |
| Bundle size aumenta por primitivas nuevas | Baja | Bajo | La extracción reemplaza JSX duplicado por imports compartidos → debe bajar o quedar igual. Medir con `vite build` antes/después. |

---

## Restricciones de scope

- **No** romper rutas existentes ni funcionalidades en uso.
- **No** introducir librerías nuevas (Tailwind + lucide + sonner + axios + DataTable + FilterBar + material-tailwind alcanzan).
- **No** cambiar contratos backend (BE ya está estandarizado en plan anterior).
- **No** refactor del giant Customers.tsx form (1142 LOC) en este plan — deuda técnica explícita.
- **No** Optimistic UI, Cmd+K, ni toggle archivados en lista activa (decisión del usuario).
- **No** reescribir DataTable, FilterBar ni Drawer externos.
- **No** inventar endpoints BE: si Create/Update faltan, se reporta como bloqueante.
- **No** agregar Archive a entidades transaccionales (Invoice, Dollar, Commerce, Stock movimientos, Crop).

---

## Resumen ejecutivo

| # | Fase | Entrega | Días |
|---|---|---|---|
| A | Extraer 9 primitivas + 3 hooks + extender copy.ts | Foundations homogéneas | 2 |
| B | Reemplazar 70+ duplicaciones inline por primitivas | UI homogénea pixel-equivalente | 2.5 |
| C | Listas activas con kebab + accesos CRUDAR (6 entidades) | Acciones uniformes en lista | 3 |
| D | Create/Edit drawers para 5 entidades faltantes | CRUDAR completo en 9 entidades | 3 |
| E | Bulk actions en 18 listas | Operaciones masivas | 2 |
| F | UX polish (sidebar collapsible, URL stale, breadcrumbs) | Pulido | 1.5 |
| G | Refactor 15 hooks a useEntityCrud | Deuda técnica saldada | 2 |

**Total: 16 días.** Hito mínimo shippeable A+B+F: ~6 días.

**Criterio de éxito final:** desde cualquiera de las 9 entidades archivables, el usuario puede hacer Create / Read / Update / Archive / Restore / Delete con la misma UX, los mismos componentes y la misma terminología. La inconsistencia visual se reduce a cero. La duplicación de código en patrones (loading/error/success/drawer/confirm) se reduce a cero. Cualquier nueva entidad futura usa las mismas primitivas extraídas — no se vuelve a copiar JSX.
