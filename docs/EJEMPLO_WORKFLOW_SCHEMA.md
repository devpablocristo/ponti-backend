# Ejemplo de Actualización del Workflow para Schema por Rama

## Cambios Necesarios en `.github/workflows/deploy-cloud-run.yml`

Agregar un step que determine `DB_SCHEMA` antes del deploy:

```yaml
- name: Set DB_SCHEMA for branch isolation
  id: set_schema
  run: |
    # Para PRs: usar pr_<number>
    if [ -n "${{ github.event.pull_request.number }}" ]; then
      DB_SCHEMA="pr_${{ github.event.pull_request.number }}"
      echo "schema=pr_${{ github.event.pull_request.number }}" >> "$GITHUB_OUTPUT"
      echo "DB_SCHEMA=pr_${{ github.event.pull_request.number }}" >> "$GITHUB_ENV"
    
    # Para develop/main: usar public (comportamiento legacy)
    elif [ "${{ github.ref_name }}" = "develop" ] || [ "${{ github.ref_name }}" = "main" ] || [ "${{ github.ref_name }}" = "staging" ]; then
      DB_SCHEMA="public"
      echo "schema=public" >> "$GITHUB_OUTPUT"
      echo "DB_SCHEMA=public" >> "$GITHUB_ENV"
    
    # Para feature branches sin PR: usar branch_<slug>_<sha>
    else
      BRANCH_SLUG=$(echo "${{ github.head_ref || github.ref_name }}" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9]/-/g' | cut -c1-30)
      SHORT_SHA=$(echo "${{ github.sha }}" | cut -c1-7)
      DB_SCHEMA="branch_${BRANCH_SLUG}_${SHORT_SHA}"
      echo "schema=branch_${BRANCH_SLUG}_${SHORT_SHA}" >> "$GITHUB_OUTPUT"
      echo "DB_SCHEMA=branch_${BRANCH_SLUG}_${SHORT_SHA}" >> "$GITHUB_ENV"
    fi
    
    echo "✅ DB_SCHEMA set to: $DB_SCHEMA"
```

Luego, en el step "Deploy to Cloud Run", agregar `DB_SCHEMA` a las variables de entorno:

```yaml
- name: Deploy to Cloud Run
  run: |
    image_uri="${{ vars.GCP_REGION }}-docker.pkg.dev/${{ env.GCP_PROJECT_ID }}/${{ vars.ARTIFACT_REGISTRY }}/${{ vars.IMAGE_NAME }}:${{ env.IMAGE_TAG }}"
    
    # Obtener DB_SCHEMA del step anterior
    DB_SCHEMA="${{ steps.set_schema.outputs.schema }}"
    
    # Prod no debe ser público, dev puede serlo
    if [ "${{ env.DEPLOY_ENV }}" = "prod" ]; then
      gcloud run deploy "${{ env.SERVICE_NAME }}" \
        --project="${{ env.GCP_PROJECT_ID }}" \
        --region="${{ vars.GCP_REGION }}" \
        --image="$image_uri" \
        --service-account="${{ env.CLOUD_RUN_SERVICE_ACCOUNT }}" \
        --update-env-vars="DB_SCHEMA=${DB_SCHEMA},..." \
        --no-allow-unauthenticated
    else
      gcloud run deploy "${{ env.SERVICE_NAME }}" \
        --project="${{ env.GCP_PROJECT_ID }}" \
        --region="${{ vars.GCP_REGION }}" \
        --image="$image_uri" \
        --service-account="${{ env.CLOUD_RUN_SERVICE_ACCOUNT }}" \
        --update-env-vars="DB_SCHEMA=${DB_SCHEMA},..." \
        --allow-unauthenticated
    fi
```

**Nota:** Reemplazar `...` con las demás variables de entorno existentes (puedes obtenerlas con `gcloud run services describe`).

---

## GitHub Action para Cleanup Automático

Crear `.github/workflows/cleanup-schema.yml`:

```yaml
name: Cleanup Schema on PR Close

on:
  pull_request:
    types: [closed]

permissions:
  contents: read
  id-token: write

jobs:
  cleanup:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set variables
        env:
          GCP_PROJECT_ID_DEV: ${{ vars.GCP_PROJECT_ID_DEV }}
          DB_HOST: ${{ secrets.DB_HOST_DEV }}
          DB_NAME: ${{ vars.DB_NAME_DEV }}
          DB_USER: ${{ secrets.DB_USER_DEV }}
          DB_PASSWORD: ${{ secrets.DB_PASSWORD_DEV }}
        run: |
          SCHEMA="pr_${{ github.event.pull_request.number }}"
          echo "SCHEMA=${SCHEMA}" >> "$GITHUB_ENV"
          echo "Cleaning up schema: ${SCHEMA}"

      - name: Auth GCP
        uses: google-github-actions/auth@v2
        with:
          workload_identity_provider: ${{ vars.WIF_PROVIDER_DEV }}
          service_account: ${{ vars.WIF_SERVICE_ACCOUNT_DEV }}

      - name: Setup gcloud
        uses: google-github-actions/setup-gcloud@v2
        with:
          project_id: ${{ vars.GCP_PROJECT_ID_DEV }}

      - name: Run cleanup script
        env:
          DB_HOST: ${{ secrets.DB_HOST_DEV }}
          DB_NAME: ${{ vars.DB_NAME_DEV }}
          DB_USER: ${{ secrets.DB_USER_DEV }}
          DB_PASSWORD: ${{ secrets.DB_PASSWORD_DEV }}
          DB_PORT: ${{ vars.DB_PORT_DEV }}
          DB_SSL_MODE: ${{ vars.DB_SSL_MODE_DEV }}
        run: |
          ./scripts/cleanup_schema.sh "${{ env.SCHEMA }}"
```

---

## Notas Importantes

1. **Para develop/main:** No es necesario cambiar nada, seguirán usando `public` por defecto.

2. **Para feature branches:** El schema se crea automáticamente en el primer deploy.

3. **Cleanup manual:** Si necesitas limpiar un schema manualmente:
   ```bash
   export DB_HOST=...
   export DB_NAME=...
   export DB_USER=...
   export DB_PASSWORD=...
   ./scripts/cleanup_schema.sh pr_123
   ```

4. **Verificación:** Después del deploy, puedes verificar que el schema se creó:
   ```sql
   \dn  -- Listar schemas
   \dt pr_123.*  -- Listar tablas en el schema
   ```
