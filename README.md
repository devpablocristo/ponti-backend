# Ponti Backend (API)

TLDR:
1. Configurá `.env`.
2. `docker compose up -d` (levanta DB y corre migraciones).
3. `go run ./cmd/api` (levanta la API local).
4. Deploys: ver `docs/DEPLOY.md`.
5. Índice de docs: `docs/README.md`.

## Local

### Requisitos
- Docker + Docker Compose
- Go (para ejecutar la API)
- Si trabajás con tooling Node alrededor del stack Ponti, usar `20.19.0` para alinear con frontend.

### Configuración
- Usamos **un solo** archivo `.env` para local.
- No hay configuración por ambiente dentro del código.
- Ejemplo base en `.env.example`.
- Si el cache de Go está vacío o usás Docker dev, configurá `GO_MODULES_TOKEN` para bajar los módulos privados de `github.com/devpablocristo/core/*`.
- En CI/deploy el token ya no viaja por `ARG`: el build prod consume un secret BuildKit (`go_modules_token`).
- `.dockerignore` excluye `.env`, artefactos y el árbol `pkg/` legacy del contexto de build prod.

```bash
cp .env.example .env
```

### Levantar servicios
```bash
docker compose up -d
```

Esto levanta:
- `ponti-db` (PostgreSQL)
- `migrate` (aplica migraciones automáticamente)
- `pgadmin` (opcional)

### Ejecutar la API
```bash
go run ./cmd/api
```

### Sincronizar DB local con dev remoto
```bash
SRC_FROM_CLOUD_RUN=1 SRC_FORCE_CLOUD_RUN=1 SRC_SERVICE_NAME=ponti-backend \
SRC_PROJECT_ID=<gcp_project_id> SRC_REGION=<gcp_region> make staging-db-2-local-db
```

## Headers requeridos
```
X-API-KEY: <tu_api_key>
X-USER-ID: 123
```

## AI (`InsightService` + `CopilotAgent`)
Flujo seguro y cerrado:
```
FE (UI)
 → BFF (ponti-frontend/api, valida JWT)
 → Backend Go (proxy seguro)
 → Ponti AI (`InsightService` + `CopilotAgent`, READ-ONLY)
```

Notas:
- El FE nunca ve claves.
- El Backend Go usa `X-SERVICE-KEY` para hablar con Ponti AI.
- Ponti AI solo lee dominio (SELECT) y solo escribe en tablas `ai_*`.

Doc de features reales:
- `docs/FEATURE-MAP.md`

## Modelo de datos (resumen)
- Project es el núcleo: Customer, Campaign, Managers, Investors y Fields.
- Fields → Lots → Crops (cultivo actual y previo).

```
Customer (1)──────(∞) Project (∞)──────(1) Campaign
                        │
                        │
           ┌────────────┴────────────┐
           │           │             │
        (∞)Manager   (∞)Investor   (∞)Field
           │           │             │
   [project_managers] [project_investors]
                                     │
                                   (∞)
                                   Lot
                                     │
                            ┌────────┴─────────┐
                            │                  │
                      (1)CurrentCrop     (1)PreviousCrop
                            │                  │
                         Crop               Crop
```

## Docs útiles
- `docs/DEPLOY.md`
- `docs/CONFIGURAR_VARIABLES_GITHUB.md`
- `docs/DIAGNOSTICO_CLOUD_RUN.md`
- `docs/ENDPOINT_NORMALIZATION.md`
