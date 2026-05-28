# Actor Master Data - Spec

## Alcance

`actors` es la identidad canonica para personas y sociedades usadas como:

- Cliente
- Inversor
- Proveedor
- Responsable
- Arrendatario
- Contratista
- Facturador

El frontend puede bloquear errores temprano, pero el backend y la base de datos son la fuente final de verdad.

## Nombre Unico

- No pueden existir dos actores activos con el mismo `normalized_name` dentro del mismo `tenant_id`.
- La unicidad no depende de `actor_kind`, roles, perfiles, email, telefono, aliases ni identificadores.
- El mismo nombre normalizado puede existir en otro tenant.
- Un actor archivado (`deleted_at IS NOT NULL`) no bloquea reutilizar el nombre.
- Un actor fusionado (`merged_into_actor_id IS NOT NULL`) no bloquea reutilizar el nombre.
- `CreateActor` y `UpdateActor` deben devolver `domainerr.KindConflict` con mensaje `actor already exists` si la base rechaza el nombre duplicado.

## Normalizacion

- El nombre persistido se guarda en `display_name`.
- La clave de unicidad se guarda en `normalized_name`.
- La normalizacion actual:
  - trim
  - lowercase
  - unaccent
  - espacios consecutivos a un solo espacio
- Si se cambia esta normalizacion, debe cambiarse en Go y en SQL al mismo tiempo.

## Migracion

- La base debe tener un indice unico parcial:
  `tenant_id, normalized_name WHERE deleted_at IS NULL AND merged_into_actor_id IS NULL`.
- El indice reemplaza el indice no unico legacy sobre esos mismos campos.
- Si hay datos duplicados activos en una base de desarrollo, la migracion debe consolidarlos antes de crear el indice:
  - conserva como canonico el actor activo de menor `id` dentro de cada `(tenant_id, normalized_name)`;
  - copia roles y metadatos no conflictivos al canonico;
  - mueve referencias operativas que no generen conflictos;
  - marca los duplicados como fusionados con `merged_into_actor_id` y `deleted_at`;
  - registra la fusion en `actor_merge_log`.

## Tests SDD

- Crear actor duplicado por nombre normalizado devuelve conflicto.
- Editar actor para tomar el nombre normalizado de otro actor activo devuelve conflicto.
- Editar actor manteniendo su propio nombre normalizado esta permitido.
- El mismo nombre normalizado en otro tenant esta permitido.
- El mismo nombre normalizado de un actor archivado o fusionado no bloquea crear uno nuevo.
