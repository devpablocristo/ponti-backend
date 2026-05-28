# Auditoria Arquitectonica: Proyectos y Archivado

Fecha: 2026-05-26

## Veredicto

Correcto con deuda.

El comportamiento de producto queda bien alineado: en la pantalla de Proyectos, las acciones operan sobre proyectos; archivar un cliente completo debe ser una accion explicita de Clientes. La correccion FE es chica y coherente con ese contrato.

En BE, el fix actual resuelve el bug real respetando el orden parent-before-child requerido por las invariantes de DB. No lo considero un parche roto, pero si una deuda: el restore de customer/project sigue usando una ruta manual porque `RunCascadeRestore` todavia no es una base segura para reemplazarla tal como esta.

## Hallazgos

- FE: `CustomersList(projectsOnly)` ahora fuerza modo proyecto y abre `Proyectos Archivados` aun cuando el filtro sea "Todos los proyectos". Esto evita el error de operar sobre clientes desde la pantalla de Proyectos.
- FE: la seleccion masiva en Proyectos arma copy de proyectos (`4 proyectos`) y no de clientes. Smoke local confirmado.
- BE: `RestoreCustomer` ahora restaura primero el customer, despues sus projects, despues fields/lots, y recien luego tablas operativas como workorders. Ese orden es arquitectonicamente correcto con triggers como `projects -> customers`, `fields -> projects`, `lots -> fields`, `workorders -> projects/fields/lots`.
- BE: el restore respeta `Cause`. Un proyecto archivado manualmente con origen `projects` no se restaura al restaurar su customer. Esto explica por que `Agro Lajitas` puede quedar con customer activo y proyectos archivados: esos proyectos tienen `archive_origin_entity = 'projects'`.
- Deuda BE: `RunCascadeRestore` camina `CascadeTables` y luego `ChildEntities` recursivos restaurando hijos antes que padres en algunos casos. No conviene migrar el restore de customer a ese helper hasta corregir y testear el orden generico.
- Deuda de datos local/prod-derived: el audit read-only `scripts/data-audit/archived_invariants.sql` reporta filas historicas inconsistentes, por ejemplo `IA-4 workorders_under_archived_parent = 9`, `IA-5 labors_under_archived_project = 33`, `IA-6 supplies_under_archived_project = 65`, `IA-7 movements_under_archived_parent = 10`, `IA-8 stocks_under_archived_parent = 8`, `IA-9 wo_items_under_archived_workorder = 50`, e `IA-14 untraceable_archives` en projects/fields/lots/workorders. Esto no lo introdujo el smoke; debe tratarse como cleanup de datos antes de endurecer mas invariantes.

## Validacion Ejecutada

- `go test ./internal/customer ./internal/project ./internal/shared/lifecycle`
- `yarn --cwd ui typecheck`
- Smoke API local con tenant `default`, customer `SOALEN SRL 26-27`:
  - proyectos activos antes: 4
  - `POST /customers/18/archive`: 200
  - proyectos archivados por customer: 4
  - `POST /customers/18/restore`: 200
  - proyectos activos despues: 4
- Smoke FE local en `localhost:5173/admin/master-data/projects/list`:
  - seleccion "Todos los proyectos" muestra filas de proyectos
  - seleccion masiva abre confirmacion con `4 proyectos`
  - `Archivados` abre `Proyectos Archivados`
  - no aparece `Clientes Archivados`
- Audit DB local:
  - queries directas de jerarquia detectan 0 para projects/fields/lots bajo padre archivado
  - detectan 9 workorders activos bajo project/field/lot archivado
  - `scripts/data-audit/archived_invariants.sql` confirma mas inconsistencias historicas relacionadas

## Recomendacion

- Mantener el fix actual para cerrar el bug de Proyectos/Clientes.
- No migrar `RestoreCustomer` a `RunCascadeRestore` todavia.
- Abrir una tarea separada para:
  - corregir `RunCascadeRestore` para restaurar padres antes que hijos cuando existan invariantes parent-child;
  - cubrirlo con tests que fallen si lots/workorders se restauran antes que fields/projects;
  - decidir si `RestoreProject` tambien debe usar ese helper cuando sea seguro.
- Abrir una tarea de data cleanup para resolver las violaciones IA-4 a IA-9 e IA-14 antes de depender de los triggers como garantia absoluta.
- Agregar un test FE de regresion para `CustomersList projectsOnly`: con `allSelection.project=true`, bulk archive usa entidad `PROJECT_ENTITY` y drawer abre `ArchivedProjects`.

## Tests Minimos Antes de Merge

- BE customer restore:
  - customer archivado con projects, fields, lots y workorders vuelve completo;
  - proyectos archivados manualmente por origen `projects` no se restauran con el customer;
  - customer activo con hijos archivados por el mismo cause repara el grafo;
  - project restore falla con `project parent customer is archived`.
- FE projects list:
  - `/admin/master-data/projects/list` siempre queda en modo proyecto;
  - "Archivados" muestra proyectos archivados;
  - seleccion masiva con "Todos los proyectos" confirma proyectos, no clientes.
