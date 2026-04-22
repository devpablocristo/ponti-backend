# 1. Resumen ejecutivo
- Estado general: el proyecto compila, `go test ./...` pasa y `go vet ./...` no reporta errores; eso prueba build básica, no solidez.
- Nivel de salud técnica: medio-bajo.
- Calidad real de la arquitectura: hay capas `handler/usecases/repository`, pero la aplicación no es hexagonal estricta; los bordes están perforados por wiring manual, singletons globales, lógica cross-módulo y transporte filtrándose a negocio.
- Riesgos principales: secretos reales expuestos, pipeline DEV con auth débil y servicio público, errores duros de integridad en `stock`, ciclo de vida inconsistente de `project`, respuestas dummy/no-op silenciosas, contratos HTTP que prometen datos que el repositorio no entrega.
- Juicio general honesto: transmite operatividad, no solidez. La base funciona, pero está llena de decisiones frágiles, inconsistencias entre módulos y deuda que ya produce bugs reales.

# 2. Cobertura de auditoría
- Archivos brutos detectados en el workspace: `7886`.
- Archivos `.git/**` detectados y excluidos del universo auditable: `7350`.
- Archivos de caché binaria excluidos: `1` (`scripts/__pycache__/compare_endpoints.cpython-312.pyc`).
- Universo auditable del proyecto: `535` archivos de código/configuración técnica.
- Total de archivos revisados: `535/535` del universo auditable.
- Archivos versionados revisados: `530`.
- Archivos ignorados locales revisados también: `.env`, `.claude/settings.local.json`, `scripts/db/db_force_reset_gcp.env`, `scripts/db/db_gcp_reset_and_load_local.env`, `scripts/db/db_staging_to_local.env`.
- Carpetas revisadas: raíz/config, `.github/workflows`, `docs`, `cmd`, `internal` completo, `migrations_v4`, `scripts`, `wire`.
- Listado de archivos revisados: completo en `# 3`.
- Archivos excluidos:
  - `.git/**`: metadatos del VCS; no son código ni configuración operativa del proyecto.
  - `scripts/__pycache__/compare_endpoints.cpython-312.pyc`: binario generado; no es auditable humanamente.
- Advertencia explícita: se revisó el `100%` del universo auditable disponible en el repo/workspace. No revisé código fuente de dependencias externas no vendorizadas (`github.com/devpablocristo/core/...`, `firebase`, etc.) porque no está dentro del repositorio; sí revisé todas sus invocaciones locales.

# 3. Inventario auditado
- Raíz/config `[revisado]`: `.air.toml`, `.claude/settings.json`, `.claude/settings.local.json`, `.codex`, `.cursor/rules/{core-rules.mdc,crud-canonical.mdc,ideal-first.mdc,no-git.mdc}`, `.cursorignore`, `.dockerignore`, `.env`, `.env.example`, `.gitignore`, `Dockerfile`, `Dockerfile.dev`, `Makefile`, `README.md`, `docker-compose.yml`, `go.mod`, `go.sum`, `ponti-backend.code-workspace`.
- `.github/workflows` `[revisado]`: `apply-migrations-prod.yml`, `audit-service-alignment.yml`, `ci-pr.yml`, `deploy-dev.yml`, `deploy-prod.yml`, `deploy-staging.yml`, `reset-db-dev-from-staging.yml`, `reset-db-staging-from-prod-sanitized.yml`.
- `docs` `[revisado]`: `README.md`, `ARCHITECTURE.md`, `PROD_DB_BACKUP_RUNBOOK.md`, `CONFIGURAR_VARIABLES_GITHUB.md`, `GITHUB_SECRETS.md`, `GCP_DB_CREDS.md`, `INVESTIGACION_ECOSISTEMA_IA_HANDOFF_CLAUDE.md`, `**Descripción General Del Sistema De IA .md`. Observación: índice roto y drift documental.
- `cmd/api` `[revisado]`: `http_server.go`, `main.go`.
- `cmd/config` `[revisado]`: `ai.go`, `api.go`, `auth.go`, `db.go`, `http_server.go`, `loadconfig.go`, `migrations.go`, `review.go`, `service.go`, `words_suggester.go`.
- `cmd/migrate` `[revisado]`: `main.go`, `migrate_gorm.go`, `migrate_sql.go`, `migrate_sql_test.go`.
- `internal/admin` `[revisado]`: `{handler.go,repository.go,idp/firebase_admin.go,idp/idp.go,idp/noop.go}`.
- `internal/ai` `[revisado]`: `{client.go,handler.go,usecases/usecases.go}`.
- `internal/business-parameters` `[revisado]`: `{handler.go,repository.go,usecases.go,handler/dto/business_parameter.go,repository/models/business_parameter.go,usecases/domain/business_parameter.go}`.
- `internal/businessinsights` `[revisado]`: `{handler.go,repository.go,service.go,service_test.go,repository/models/candidate_model.go,repository/models/read_model.go}`.
- `internal/campaign` `[revisado]`: `{handler.go,repository.go,usecases.go,handler/dto/campaign.go,repository/models/campaign.go,usecases/domain/campaign.go}`.
- `internal/category` `[revisado]`: `{handler.go,repository.go,usecases.go,handler/dto/requests.go,handler/dto/responses.go,repository/models/category.go,usecases/domain/category.go}`.
- `internal/class-type` `[revisado]`: `{handler.go,repository.go,usecases.go,handler/dto/requests.go,handler/dto/responses.go,repository/models/classtype.go,usecases/domain/classtype.go}`.
- `internal/commercialization` `[revisado]`: `{handler.go,repository.go,usecases.go,handler/dto/commercialization.go,handler/dto/post_request.go,repository/models/commercialization.go,usecases/domain/domain.go}`.
- `internal/crop` `[revisado]`: `{handler.go,repository.go,usecases.go,handler/dto/requests.go,handler/dto/responses.go,repository/models/create_crop.go,repository/models/crop.go,usecases/domain/crop.go}`.
- `internal/customer` `[revisado]`: `{handler.go,repository.go,usecases.go,repository_harddelete_test.go,handler/dto/requests.go,handler/dto/responses.go,repository/models/create_customer.go,repository/models/customer.go,usecases/domain/customer.go}`.
- `internal/dashboard` `[revisado]`: `{handler.go,repository.go,usecases.go,handler/dto/dashboard.go,handler/dto/dashboard_test.go,repository/models/dashboard.go,repository/models/mappers.go,repository/repository_test.go,usecases/domain/dashboard.go}`.
- `internal/data-integrity` `[revisado]`: `{handler.go,usecases.go,usecases_mock_test.go,usecases_test.go,handler/dto/integrity_check.go,usecases/domain/types.go}`.
- `internal/dollar` `[revisado]`: `{handler.go,repository.go,usecases.go,handler/dto/base.go,handler/dto/put_request.go,repository/models/dollar.go,usecases/domain/dollar.go}`.
- `internal/field` `[revisado]`: `{handler.go,repository.go,usecases.go,handler/dto/requests.go,handler/dto/responses.go,repository/models/create_field.go,repository/models/field.go,usecases/domain/field.go}`.
- `internal/investor` `[revisado]`: `{handler.go,repository.go,usecases.go,handler/dto/requests.go,handler/dto/responses.go,repository/models/create_investor.go,repository/models/investor.go,usecases/domain/investor.go}`.
- `internal/invoice` `[revisado]`: `{handler.go,repository.go,usecase.go,repository_test.go,handler/dto/invoice.go,handler/dto/list_invoice.go,handler/dto/post_invoice.go,repository/models/invoice.go,usecases/domain/invoice.go}`.
- `internal/labor` `[revisado]`: `{excel-service.go,handler.go,repository.go,usecases.go,handler_update_labor_test.go,repository_test.go,excel/config.go,excel/excel-dto.go,excel/excel-table-dto.go,handler/dto/base.go,handler/dto/base_test.go,handler/dto/create_labors.go,handler/dto/helpers.go,handler/dto/labor_metrics.go,handler/dto/list_labor.go,handler/dto/list_labor_categories.go,handler/dto/list_labor_group.go,handler/dto/list_labor_group_test.go,handler/dto/list_labor_workorder.go,repository/models/base.go,repository/models/labor_category.go,repository/models/labor_item.go,usecases/domain/domain.go,usecases/domain/labor_category.go,usecases/domain/labor_list_workorder.go,usecases/domain/labor_metrics.go,utils/month_mapper.go,utils/month_mapper_test.go}`.
- `internal/lease-type` `[revisado]`: `{handler.go,repository.go,usecases.go,handler/dto/requests.go,handler/dto/responses.go,repository/models/leasetype.go,usecases/domain/leasetype.go}`.
- `internal/lot` `[revisado]`: `{excel-service.go,handler.go,repository.go,usecases.go,repository_upsert_test.go,usecases_test.go,validations.go,validations_test.go,excel/config.go,excel/excel-dto.go,handler/dto/create_lot.go,handler/dto/list_lot_table.go,handler/dto/lot-table.go,handler/dto/lot.go,handler/dto/lot_metrics.go,handler/dto/lot_update.go,repository/models/lot-table.go,repository/models/lot.go,repository/models/lot_dates_operations.go,usecases/domain/lot-table.go,usecases/domain/lot.go,usecases/domain/lot_filters.go,usecases/mocks/usecases_mock.go}`.
- `internal/manager` `[revisado]`: `{handler.go,repository.go,usecases.go,handler/dto/requests.go,handler/dto/responses.go,repository/models/create_manager.go,repository/models/manager.go,usecases/domain/manager.go}`.
- `internal/platform` `[revisado]`: `config/godotenv/godotenv.go`, `files/excel/excelize/{bootstrap.go,config.go,service.go}`, `http/middlewares/gin/{error_handling.go,local_dev_authz.go,middleares.go,request_and_response_logger.go,require_credentials.go,require_identity_platform_authz.go,require_jwt.go,require_user_id_header.go,types.go}`, `http/servers/gin/{bootstrap.go,config.go,server.go}`, `persistence/gorm/{bootstrap.go,config.go,repository.go}`, `words-suggesters/trigram-search/{bootstrap.go,config.go,gorm_adapter.go,suggester.go,types.go}`.
- `internal/project` `[revisado]`: `{handler.go,repository.go,usecases.go,words-suggester.go,handler_integration_test.go,handler/dto/create_project.go,handler/dto/list_projects.go,handler/dto/project.go,handler/dto/project_unmarshal_test.go,repository/models/project.go,usecases/domain/project.go}`.
- `internal/provider` `[revisado]`: `{handler.go,repository.go,usecases.go,handler/dto/get_providers.go,repository/models/base.go,usecases/domain/provider.go}`.
- `internal/report` `[revisado]`: `{handler.go,helpers.go,repository.go,usecases.go,handler/dto/field-crop.go,handler/dto/investors-contributors.go,handler/dto/summary-results.go,repository/models/field-crop.go,repository/models/helpers.go,repository/models/investor-contribution.go,repository/models/investor_contribution_test.go,repository/models/summary-results.go,usecases/domain/field-crop.go,usecases/domain/investor-contribution.go,usecases/domain/summary-results.go,usecases/mappers/summary_mappers.go,usecases/validators.go}`.
- `internal/reviewproxy` `[revisado]`: `{client.go,client_test.go}`.
- `internal/shared` `[revisado]`: `db/{report_schema.go,tx_context.go}`, `domain/base.go`, `filters/workspace.go`, `handlers/{auth.go,bind.go,errors.go,pagination.go,params_compat.go,query.go,responses.go,workspace_filters.go}`, `models/base.go`, `repository/{errors.go,validation.go}`, `types/{errors_compat.go,page_info.go,requests.go}`, `utils/{files.go,jwt_tools.go,strings.go}`.
- `internal/stock` `[revisado]`: `{excel-service.go,handler.go,repository.go,repository_tx.go,usecases.go,excel/config.go,excel/excel-dto.go,handler/dto/stock.go,handler/dto/stock_test.go,handler/dto/update_close_date.go,handler/dto/update_real_stock.go,repository/models/base.go,repository/models/update_close_date.go,repository/models/update_real_stock.go,repository_test.go,usecases/domain/stock.go,usecases/domain/stock_test.go}`.
- `internal/supply` `[revisado]`: `{excel-service.go,handler.go,repository.go,repository_movement.go,repository_tx.go,usecases.go,usecases_movement.go,usecases_movement_import.go,usecases_delete_test.go,usecases_movement_test.go,repository_delete_test.go,repository_movement_delete_test.go,handler_update_supply_test.go,mocks/mock_repository.go,mocks/mock_stock_usecases.go,excel/config.go,excel/excel_dto.go,excel/excel_dto_table.go,excel/helpers.go,handler/dto/create/create_request.go,handler/dto/create/create_request_test.go,handler/dto/create/create_supply_movement_request.go,handler/dto/create/create_supply_movement_response.go,handler/dto/create/pending_request.go,handler/dto/get/get_providers_response.go,handler/dto/get/get_supply_movement_response.go,handler/dto/list/list_response.go,handler/dto/list/list_response_test.go,handler/dto/update/update_supply_movement_request.go,repository/models/supply.go,repository/models/supply_movement.go,usecases/domain/entry_type.go,usecases/domain/supply.go,usecases/domain/supply_filters.go,usecases/domain/supply_movement.go}`.
- `internal/work-order` `[revisado]`: `{excel-service.go,handler.go,repository.go,usecases.go,handler_test.go,usecases_test.go,excel/config.go,excel/excel-dto.go,handler/dto/investor_payment_status.go,handler/dto/list_work_order.go,handler/dto/work_order.go,handler/dto/work_order_metrics.go,repository/models/list_work_order.go,repository/models/work_order.go,repository/models/work_order_investor_split.go,usecases/domain/work_order.go,usecases/domain/work_order_list.go,usecases/domain/work_order_metrics.go}`.
- `internal/work-order-draft` `[revisado]`: `{handler.go,pdf-service.go,repository.go,usecases.go,usecases_test.go,handler/dto/work_order_draft.go,repository/models/work_order_draft.go,usecases/domain/work_order_draft.go}`. Observación: es de las áreas más consistentes del repo.
- `migrations_v4` `[revisado]`: `000001_extensions.{up,down}.sql`, `000010_core_tables.{up,down}.sql`, `000015_core_functions_triggers.{up,down}.sql`, `000020_projects_tables.{up,down}.sql`, `000030_fields_lots_tables.{up,down}.sql`, `000040_crops_tables.{up,down}.sql`, `000050_workorders_labors_tables.{up,down}.sql`, `000060_supplies_inventory_tables.{up,down}.sql`, `000070_investors_commercialization_tables.{up,down}.sql`, `000080_constraints_fks_indexes.{up,down}.sql`, `000090_v4_schemas.{up,down}.sql`, `000100_v4_core_functions.{up,down}.sql`, `000110_v4_ssot_functions.{up,down}.sql`, `000120_v4_calc_views.{up,down}.sql`, `000130_v4_report_views.{up,down}.sql`, `000140_compat_seeded_area_aliases.{up,down}.sql`, `000150_dashboard_field_filters.{up,down}.sql`, `000160_lot_dates_unique_lot_sequence.{up,down}.sql`, `000170_fix_dashboard_costs_hectares.{up,down}.sql`, `000180_authn_authz_mvp.{up,down}.sql`, `000190_workorder_investor_splits.{up,down}.sql`, `000191_investor_real_contributions_use_workorder_splits.{up,down}.sql`, `000192_labor_list_use_workorder_splits.{up,down}.sql`, `000193_auth_fk_indexes.{up,down}.sql`, `000194_workorder_list_add_labor_row.{up,down}.sql`, `000195_fix_dashboard_field_views.{up,down}.sql`, `000196_allow_archived_labors_supplies_in_views.{up,down}.sql`, `000197_supply_partial_price_flag.{up,down}.sql`, `000198_labor_partial_price_flag.{up,down}.sql`, `000199_supply_return_movement_type.{up,down}.sql`, `000200_supply_return_reporting_views.{up,down}.sql`, `000201_auth_uuid_migration.{up,down}.sql`, `000202_workorder_split_payment_status.{up,down}.sql`, `000203_stock_real_count_flag.{up,down}.sql`, `000204_invoice_per_investor.{up,down}.sql`, `000205_work_order_drafts.{up,down}.sql`, `000206_reconcile_work_order_drafts_schema.{up,down}.sql`, `000207_work_order_drafts_digital_flag.{up,down}.sql`, `000208_workorder_list_include_digital_drafts.{up,down}.sql`, `000209_business_insight_candidates.{up,down}.sql`, `000210_business_insight_reads.{up,down}.sql`, `000211_supply_pending_flag.{up,down}.sql`, `000212_supply_name_snapshot_in_order_items.{up,down}.sql`, `000213_dashboard_balance_extra_stocks.{up,down}.sql`.
- `scripts` `[revisado]`: `check_schema_guardrails.sh`, `docker_cleanup.sh`, `down_ponti_local.sh`, `entrypoint.sh`, `run_ponti_local.sh`, `smoke_release.sh`.
- `scripts/db` `[revisado]`: `db_adopt_baseline.sh`, `db_force_reset_gcp.env`, `db_force_reset_gcp.env.example`, `db_force_reset_gcp.sh`, `db_gcp_reset_and_load_local.env`, `db_gcp_reset_and_load_local.env.example`, `db_gcp_reset_and_load_local.sh`, `db_migrate_up.sh`, `db_reset.sh`, `db_schema_diff.sh`, `db_schema_snapshot.sh`, `db_staging_to_local.env`, `db_staging_to_local.env.example`, `db_staging_to_local.sh`, `db_validate.sh`, `db_validate.sql`, `hardening_post_restore.sql`, `repair_stocks_investor_granularity.sql`, `sanitize_prod_to_staging.sql`, `schema.expected.sql`, `schema.snapshot.sql`. Observación: tooling poderoso y destructivo.
- `wire` `[revisado]`: `{admin_providers.go,ai_providers.go,bootstrap_providers.go,business_parameters_providers.go,campaign_providers.go,category_providers.go,class_type_providers.go,commercialization_providers.go,config_providers.go,crop_providers.go,customer_providers.go,dashboard_providers.go,data_integrity_providers.go,dollar_providers.go,field_providers.go,investor_providers.go,invoice_providers.go,labor_providers.go,lease_type_providers.go,lot_providers.go,manager_providers.go,middleware_providers.go,project_providers.go,provider_providers.go,report_providers.go,stock_providers.go,supply_providers.go,wire.go,wire_gen.go,work_order_draft_providers.go,work_order_providers.go}`.

# 4. Hallazgos críticos
1. `Credenciales reales expuestas en archivos versionados y material local sensible`. Ubicación: [docs/GCP_DB_CREDS.md](/home/pablo/Projects/Pablo/ponti/ponti-backend/docs/GCP_DB_CREDS.md:18), [scripts/db/db_force_reset_gcp.env.example](/home/pablo/Projects/Pablo/ponti/ponti-backend/scripts/db/db_force_reset_gcp.env.example:7), [.env](/home/pablo/Projects/Pablo/ponti/ponti-backend/.env:7), [.claude/settings.local.json](/home/pablo/Projects/Pablo/ponti/ponti-backend/.claude/settings.local.json:22). Evidencia: el repo contiene `SRC_PASS='Soalen*25.'`, host público, API keys, y un PAT de GitHub (`GO_MODULES_TOKEN`). El example de `db_force_reset_gcp` dice “editar y poner SRC_PASS real”, pero ya trae una password real. Explicación técnica: no es secret management; es secret sprawl. Impacto: compromiso directo de DB/infra, fuga de acceso a módulos privados y propagación accidental vía logs, backups o herramientas. Severidad: crítica.

2. `El workflow DEV despliega un backend público con auth local insegura`. Ubicación: [deploy-dev.yml](/home/pablo/Projects/Pablo/ponti/ponti-backend/.github/workflows/deploy-dev.yml:93), [local_dev_authz.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/platform/http/middlewares/gin/local_dev_authz.go:17), [local_dev_authz.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/platform/http/middlewares/gin/local_dev_authz.go:43). Evidencia: el deploy setea `AUTH_ENABLED=false` y `--allow-unauthenticated`; el middleware “local dev” no valida firmas JWT, acepta `X-USER-ID`/`sub`, usa `local-dev-user` si no hay identidad y asigna rol `admin`. Explicación: el perímetro queda reducido, como máximo, a una API key compartida cuyo enforcement real no pude auditar porque vive en dependencia externa. Impacto: si esa key se expone, el servicio queda con privilegios de admin sobre Internet. Severidad: crítica.




























///////////////////////////////////////////
///////////////////////////////////////////

Tareas revisadas:

3. `La integridad de stock está rota por locking/timestamps inconsistentes y mapeo semántico incorrecto`. Ubicación: [internal/stock/repository.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/stock/repository.go:141), [internal/stock/repository.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/stock/repository.go:170), [internal/stock/repository/models/base.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/stock/repository/models/base.go:47), [internal/stock/handler.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/stock/handler.go:201). Evidencia: `UpdateRealStockUnits` filtra por `updated_at` viejo y vuelve a persistir ese mismo timestamp; `UpdateUnitsConsumed` tampoco lo avanza; `UnitsTransferred` se llena desde `UnitsConsumed`. Explicación: el optimistic locking no versiona realmente el registro y la entidad devuelta al dominio mezcla consumido con transferido. Impacto: conflictos perdidos/falsos, inventario y balance económico incorrectos. Severidad: crítica.

stocks 
lotes
work-order
project
supply

Después de esos, en segunda prioridad:
field
customer
crop
category
class-type
lease-type
manager
investor
business-parameters

///////////////////////////////////////////
///////////////////////////////////////////






















4. `El ciclo de vida de project es incoherente y puede fallar o dejar datos activos colgando`. Ubicación: [internal/project/repository.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/project/repository.go:571), [internal/project/repository.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/project/repository.go:793), [internal/customer/repository.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/customer/repository.go:285), [000080_constraints_fks_indexes.up.sql](/home/pablo/Projects/Pablo/ponti/ponti-backend/migrations_v4/000080_constraints_fks_indexes.up.sql:196), [internal/labor/repository.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/labor/repository.go:143). Evidencia: `ArchiveProject`, `RestoreProject` y `DeleteProject` no tocan `labors`; la FK `fk_labors_project` es `ON DELETE RESTRICT`; el borrado en cascada del módulo `customer` sí elimina `labors`; el listado de labores filtra solo por `project_id`. Explicación: hay dos políticas divergentes para la misma relación. Impacto: hard delete de proyecto potencialmente roto con datos normales y soft delete que deja labores activas asociadas a proyectos archivados. Severidad: crítica.

# 5. Hallazgos importantes
1. `Update de supply movement tiene nil dereference y validaciones incompletas`. Ubicación: [update_supply_movement_request.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/supply/handler/dto/update/update_supply_movement_request.go:71), [update_supply_movement_request.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/supply/handler/dto/update/update_supply_movement_request.go:146), [server.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/platform/http/servers/gin/server.go:38), [error_handling.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/platform/http/middlewares/gin/error_handling.go:10). Evidencia: dereferencia `*provider.ID` y `*provider.Name` sin nil-check; además el stack global no incluye `gin.Recovery()`. Explicación: una request mal formada puede disparar panic en el path de update. Impacto: 500 no controlado, ruido operativo y superficie de DoS por request. Severidad: alta.

2. `STAGING y PROD están configurados para desplegarse públicamente, contradiciendo la documentación`. Ubicación: [deploy-staging.yml](/home/pablo/Projects/Pablo/ponti/ponti-backend/.github/workflows/deploy-staging.yml:88), [deploy-prod.yml](/home/pablo/Projects/Pablo/ponti/ponti-backend/.github/workflows/deploy-prod.yml:120), [docs/GITHUB_SECRETS.md](/home/pablo/Projects/Pablo/ponti/ponti-backend/docs/GITHUB_SECRETS.md:55). Evidencia: ambos workflows usan `--allow-unauthenticated`; la doc dice que PROD tiene `--no-allow-unauthenticated`. Explicación: la seguridad perimetral documentada no coincide con el pipeline real. Impacto: superficie pública innecesaria y falsa sensación de protección IAM. Severidad: alta.

3. `El proxy AI confía en headers de transporte para identidad/ámbito y además se degrada con respuestas dummy`. Ubicación: [internal/ai/handler.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/ai/handler.go:156), [internal/ai/usecases/usecases.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/ai/usecases/usecases.go:37), [cmd/api/http_server.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/cmd/api/http_server.go:110). Evidencia: el handler toma `X-USER-ID` y `X-PROJECT-ID` del request en lugar del contexto autenticado; si el AI service no está configurado devuelve `200` con payload dummy; business insights también se convierte en no-op si falta `REVIEW_URL`. Explicación: el contrato visible no distingue entre servicio real y degradado, y el borde HTTP no ata identidad al middleware de auth. Impacto: bypass conceptual de contexto, respuestas engañosas y features que “funcionan” sin funcionar. Severidad: alta.

4. `Los reportes multi-proyecto colapsan a un solo proyecto y pierden cancelación del request`. Ubicación: [internal/report/repository.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/report/repository.go:668). Evidencia: `getRelatedProjectIDs` usa `context.Background()` y `getProjectInfo` toma `projectIDs[0]`. Explicación: cuando un filtro resuelve más de un proyecto, la metadata del reporte se calcula sobre el primero y no sobre el conjunto real. Impacto: reportes inconsistentes y sin propagación de contexto/cancelación. Severidad: alta.

5. `El módulo field promete lots en la API pero el repositorio no los carga ni los mapea`. Ubicación: [internal/field/handler/dto/responses.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/field/handler/dto/responses.go:23), [internal/field/repository.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/field/repository.go:57), [internal/field/repository/models/field.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/field/repository/models/field.go:55). Evidencia: el DTO expone `lots`; `ListFields`/`GetField` no hacen preload; `ToDomain()` no mapea `Lots` ni inversores. Explicación: el contrato público y la implementación no coinciden. Impacto: respuestas incompletas que parecen válidas. Severidad: alta.

6. `El dashboard altera semántica financiera y destruye precisión`. Ubicación: [mappers.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/dashboard/repository/models/mappers.go:167), [mappers.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/dashboard/repository/models/mappers.go:301), [dashboard.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/dashboard/handler/dto/dashboard.go:493), [dashboard.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/dashboard/handler/dto/dashboard.go:680). Evidencia: sin inversores devuelve `ProgressPct=100`; `ResultUSD` se llena con `IncomeUSD`; todos los decimales se redondean a enteros; si `RentUSD > RentExecutedUSD` intercambia ambos valores. Explicación: no es solo presentación; hay mutación de significado. Impacto: métricas financieras incorrectas al usuario. Severidad: alta.

7. `El endpoint de duplicación de work order es un stub que responde éxito sin hacer nada`. Ubicación: [internal/work-order/handler.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/work-order/handler.go:113), [internal/work-order/usecases.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/work-order/usecases.go:68), [internal/work-order/handler_test.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/work-order/handler_test.go:32). Evidencia: el handler devuelve `201` fijo y el use case retorna `("", nil)`; no hay test real del endpoint. Explicación: es funcionalidad simulada, no implementada. Impacto: falso positivo funcional y posible corrupción de UX/automatismos cliente. Severidad: alta.

8. `Data-integrity mezcla dependencia muerta con múltiples nil dereferences potenciales`. Ubicación: [internal/data-integrity/usecases.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/data-integrity/usecases.go:49), [internal/data-integrity/usecases.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/data-integrity/usecases.go:315), [internal/data-integrity/usecases.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/data-integrity/usecases.go:351). Evidencia: `stockRepo` se inyecta pero no se usa; varios controles asumen `sd.dashboardData.ManagementBalance.Summary` no nil. Explicación: la orquestación está acoplada y no protege estructuras opcionales. Impacto: panic risk y ruido arquitectónico. Severidad: alta.

9. `ListCustomers(status=all)` rompe la paginación`. Ubicación: [internal/customer/handler.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/customer/handler.go:126). Evidencia: pagina activos y archivados por separado, luego concatena ambas páginas y suma totales. Explicación: no existe un orden global paginado. Impacto: páginas inconsistentes, duplicadas o faltantes según volumen. Severidad: alta.

10. `Commercialization normaliza en exceso errores y trata lista vacía como NotFound`. Ubicación: [internal/commercialization/usecases.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/commercialization/usecases.go:39), [internal/commercialization/repository.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/commercialization/repository.go:56). Evidencia: update/create aplanan toda causa en `Internal`; `ListByProject` devuelve `NotFound` si no hay filas. Explicación: eso destruye semántica de dominio y complica clientes. Impacto: manejo de errores pobre y UX/API inconsistente. Severidad: alta.

# 6. Hallazgos menores
1. `Investor` expone `percentage` en request pero la persistencia lo descarta. Ubicación: [requests.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/investor/handler/dto/requests.go:8), [investor.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/investor/repository/models/investor.go:9). Impacto: contrato mentiroso.

2. `Manager.Type` existe en dominio pero no en modelo ni mapping. Ubicación: [manager.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/manager/usecases/domain/manager.go:10), [manager.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/manager/repository/models/manager.go:10). Impacto: campo huérfano.

3. `GetApiVersion()` devuelve el puerto y además no se usa. Ubicación: [internal/platform/http/servers/gin/config.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/platform/http/servers/gin/config.go:23). Impacto: bug latente + API muerta.

4. `lot` documenta optimistic locking que no implementa y bloquea updates a cero/vacío. Ubicación: [internal/lot/repository.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/lot/repository.go:124), [internal/lot/handler/dto/lot_update.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/lot/handler/dto/lot_update.go:14). Impacto: semántica parcial de PATCH y conflicto falso en comentario.

5. `labor` conserva legacy explícito y fallback silencioso de IVA al `10.5%`. Ubicación: [internal/labor/repository.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/labor/repository.go:478), [internal/labor/repository.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/labor/repository.go:494). Impacto: comportamiento implícito difícil de razonar.

6. `customer` ignora el error al borrar `lot_dates` en hard delete. Ubicación: [internal/customer/repository.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/customer/repository.go:310). Impacto: fallas parciales silenciosas.

7. `docs/README` referencia múltiples documentos inexistentes y `docs/ARCHITECTURE` describe endpoints AI que el código no expone. Ubicación: [docs/README.md](/home/pablo/Projects/Pablo/ponti/ponti-backend/docs/README.md:7), [docs/ARCHITECTURE.md](/home/pablo/Projects/Pablo/ponti/ponti-backend/docs/ARCHITECTURE.md:44), [internal/ai/handler.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/ai/handler.go:55). Impacto: documentación no confiable.

# 7. Violaciones de arquitectura
- Ubicación: [cmd/api/http_server.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/cmd/api/http_server.go:32). Principio violado: composición centralizada/DI coherente. Por qué rompe: `businessinsights` se arma manualmente fuera de `wire`. Consecuencia: wiring parcial, features opcionales silenciosas y más caminos de inicialización.
- Ubicación: [internal/platform/http/servers/gin/server.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/platform/http/servers/gin/server.go:14). Principio violado: infraestructura sin estado global. Por qué rompe: `instance`, `once` e `initError` convierten al server en singleton de proceso. Consecuencia: pruebas, reinicialización y aislamiento de config débiles.
- Ubicación: [internal/ai/handler.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/ai/handler.go:156), [require_identity_platform_authz.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/platform/http/middlewares/gin/require_identity_platform_authz.go:133). Principio violado: el caso de uso debe depender de identidad/tenant ya resueltos, no del transporte. Por qué rompe: el handler relee headers de usuario/proyecto en vez del contexto autenticado. Consecuencia: borde HTTP contaminando negocio.
- Ubicación: [internal/project/repository.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/project/repository.go:600), [internal/customer/repository.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/customer/repository.go:257), [internal/field/repository/models/field.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/field/repository/models/field.go:25). Principio violado: bounded contexts con bajo acoplamiento. Por qué rompe: repositorios conocen tablas/modelos de otros módulos y ejecutan cascadas manuales cross-aggregate. Consecuencia: lógica duplicada y divergente.
- Ubicación: [internal/platform/http/middlewares/gin/middleares.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/platform/http/middlewares/gin/middleares.go:38). Principio violado: contratos de middleware significativos. Por qué rompe: existe `GetProtected()` en todas las interfaces, pero la implementación devuelve slice vacío y no hay usos. Consecuencia: API interna inflada y diseño incompleto.

# 8. Código muerto, huérfano, legacy o deprecated
- `internal/work-order/usecases.DuplicateWorkOrder` y `Handler.DuplicateWorkOrder`; evidencia: [usecases.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/work-order/usecases.go:68), [handler.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/work-order/handler.go:113); categoría: stub/huérfano; riesgo: éxito falso.
- `internal/labor/repository.ListGroupLaborOld`; evidencia: [repository.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/labor/repository.go:494); categoría: legacy explícito; riesgo: duplicación conceptual.
- `internal/manager/usecases/domain.Manager.Type`; evidencia: [manager.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/manager/usecases/domain/manager.go:13); categoría: campo huérfano; riesgo: contrato ambiguo.
- `internal/data-integrity.UseCases.stockRepo`; evidencia: [usecases.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/data-integrity/usecases.go:54); categoría: dependencia muerta; riesgo: complejidad sin valor.
- `internal/platform/http/middlewares/gin/{RequireUserIDHeader,RequireCredentials,RequireJWT}`; evidencia: [require_user_id_header.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/platform/http/middlewares/gin/require_user_id_header.go:15), [require_credentials.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/platform/http/middlewares/gin/require_credentials.go:14), [require_jwt.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/platform/http/middlewares/gin/require_jwt.go:17), búsquedas sin referencias locales; categoría: middleware no cableado; riesgo: superficie muerta con bugs latentes.
- `internal/platform/http/servers/gin.Config.GetApiVersion`; evidencia: [config.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/platform/http/servers/gin/config.go:23), sin usos locales; categoría: getter muerto y roto; riesgo: bug futuro.
- `internal/platform/persistence/gorm.connectWithConnectorIAMAuthN`; evidencia: [repository.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/platform/persistence/gorm/repository.go:117); categoría: código inactivo (`//nolint:unused`); riesgo: si se activa, aborta proceso con `log.Fatalf`.

# 9. Riesgos de seguridad
- `Secret sprawl`: credenciales reales en [docs/GCP_DB_CREDS.md](/home/pablo/Projects/Pablo/ponti/ponti-backend/docs/GCP_DB_CREDS.md:18), [scripts/db/db_force_reset_gcp.env.example](/home/pablo/Projects/Pablo/ponti/ponti-backend/scripts/db/db_force_reset_gcp.env.example:7) y secretos locales en [.env](/home/pablo/Projects/Pablo/ponti/ponti-backend/.env:7). Impacto probable: compromiso de DB, tokens privados y operativa.
- `DEV público + auth laxa`: [deploy-dev.yml](/home/pablo/Projects/Pablo/ponti/ponti-backend/.github/workflows/deploy-dev.yml:96) + [local_dev_authz.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/platform/http/middlewares/gin/local_dev_authz.go:17). Impacto probable: acceso remoto con una sola shared API key.
- `Logs excesivos`: [request_and_response_logger.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/platform/http/middlewares/gin/request_and_response_logger.go:36), [middleares.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/platform/http/middlewares/gin/middleares.go:25), [require_identity_platform_authz.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/platform/http/middlewares/gin/require_identity_platform_authz.go:348). Impacto probable: fuga de headers sensibles, user IDs, tenant IDs y rutas.
- `Identidad desacoplada del borde real`: [internal/ai/handler.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/ai/handler.go:156). Impacto probable: request con headers manipulados, incluso con auth válida, puede desalinear contexto vs token.
- `Validación auth no totalmente auditable`: [wire/middleware_providers.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/wire/middleware_providers.go:22) prepara `Issuer`/`Audience`, pero [require_identity_platform_authz.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/platform/http/middlewares/gin/require_identity_platform_authz.go:63) solo pasa `JWKSURL` al verificador externo. No puedo confirmar desde el repo si la librería externa valida `iss/aud`.

# 10. Problemas de performance y escalabilidad
- `Suggester híbrido duplica filas y sobrecuenta`: [suggester.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/platform/words-suggesters/trigram-search/suggester.go:110) usa `UNION ALL`; [suggester.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/platform/words-suggesters/trigram-search/suggester.go:170) suma `pref + fuzzy`. Impacto esperado: paging incorrecto y trabajo extra.
- `Data-integrity materializa demasiados datos por request`: [usecases.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/data-integrity/usecases.go:89) pide hasta `10000` lotes y además dashboard, field metrics, summary results e investor report antes de correr controles. Impacto esperado: consumo alto de memoria/latencia en proyectos grandes.
- `Project/customer delete/archive hacen muchas operaciones seriales cross-tabla`: [internal/project/repository.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/project/repository.go:600), [internal/customer/repository.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/customer/repository.go:257). Impacto esperado: transacciones largas, mayor lock contention y mantenimiento costoso.
- `ReportRepository` pierde cancelación con `context.Background()`: [internal/report/repository.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/report/repository.go:679). Impacto esperado: queries que no se cancelan con el request original.
- `No encontré un N+1 clásico inequívoco leyendo el repo`; sí encontré mucha lógica de agregación y cascadas manuales que escala mal por cantidad de tablas/rows.

# 11. Calidad de testing
- Qué existe: `33` archivos de test; cobertura relativamente mejor en `supply`, `labor`, `lot`, `stock`, `work-order-draft`, `dashboard`, `data-integrity`, `project` handler, `migrate`, `reviewproxy`.
- Qué falta: no hay tests dedicados en `admin`, `ai`, `business-parameters`, `campaign`, `category`, `class-type`, `commercialization`, `crop`, `dollar`, `field`, `investor`, `lease-type`, `manager`, `platform`, `provider`, `shared`.
- Zonas no cubiertas: delete/archive/restore real de `project`, middlewares de auth/logging, scripts destructivos, workflows, migraciones SQL individuales, proxy AI, business-parameters, provider, field.
- Debilidades de diseño: varios tests son unitarios con stubs y no ejercen el flujo real; ejemplo claro: `work-order` tiene stub para `DuplicateWorkOrder` pero no test del endpoint defectuoso.
- Riesgos por falta de cobertura: bugs de lifecycle, seguridad perimetral y contratos HTTP inconsistentes pueden pasar porque el suite verifica compilación/comportamiento básico, no integridad sistémica.
- CI actual: [ci-pr.yml](/home/pablo/Projects/Pablo/ponti/ponti-backend/.github/workflows/ci-pr.yml:17) corre lint/build/test; `govulncheck` existe pero es `continue-on-error`, así que no bloquea PRs por hallazgos de seguridad.
- Resultado empírico de esta auditoría: `go test ./...` y `go vet ./...` pasan, aun con bugs funcionales y de seguridad relevantes.

# 12. Prioridad recomendada de corrección
1. Cortar secret sprawl y revocar credenciales expuestas.
2. Eliminar `AUTH_ENABLED=false` + `--allow-unauthenticated` del deploy DEV; al menos alinear auth real y perímetro.
3. Corregir la integridad de `stock` (`updated_at`, optimistic locking, `UnitsTransferred`).
4. Arreglar lifecycle de `project` con `labors` y unificarlo con `customer`.
5. Corregir `supply` update para evitar panics y agregar recovery real.
6. Corregir `report` multi-proyecto y propagación de contexto.
7. Reparar `field` contract mismatch y `dashboard` mappers.
8. Eliminar stubs/no-op silenciosos (`work-order duplicate`, AI dummy, businessinsights no-op si aplica).
9. Limpiar código muerto/legacy y reducir superficies no cableadas.
10. Reforzar testing sobre lifecycle, auth middlewares, scripts y flujos de integración.

# 13. Conclusión final
La arquitectura está aplicada solo de forma parcial. El proyecto usa nombres de capas compatibles con hexagonal, pero en la práctica no mantiene fronteras estrictas: el borde HTTP reinyecta headers como identidad, el wiring está partido, hay singletons globales, y la persistencia conoce demasiados detalles de otros módulos.

El sistema hoy transmite más fragilidad que solidez. Lo más preocupante no es que falle al compilar; eso no pasa. Lo preocupante es que compila y testea en verde mientras conserva secretos reales expuestos, flujos críticos stubbeados, lifecycle inconsistente de entidades core y errores silenciosos que degradan contratos sin avisar. Esa combinación es la firma de un backend operativo, pero peligrosamente confiado.