# Ponti Backend (API)

TLDR:
1. Configurá `.env`.
2. `docker compose up -d` (levanta DB y corre migraciones).
3. `go run ./cmd/api` (levanta la API local).

## Local

### Requisitos
- Docker + Docker Compose
- Go (para ejecutar la API)

### Configuración
- Usamos **un solo** archivo `.env` para local.
- No hay configuración por ambiente dentro del código.
- Ejemplo base en `.env.example`.

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

## Headers requeridos
```
X-API-KEY: abc123secreta
X-USER-ID: 123
```

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
