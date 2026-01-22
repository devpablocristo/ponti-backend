#!/bin/bash
# Script para completar la configuración de producción
# Ejecutar cuando Cloud SQL esté en estado RUNNABLE

set -e

PROJECT_ID="new-ponti-prod"
INSTANCE_NAME="ponti-prod-db"
DB_NAME="ponti_api_db"
DB_USER="ponti-prod-user"
DB_PASSWORD="APzpgD9NrQG5n5VF1iKEdVRFY"  # Contraseña generada
REGION="us-central1"
SERVICE_NAME="ponti-backend-prod"

echo "=== Completando configuración de producción ==="
echo ""

# Verificar que Cloud SQL esté listo
echo "Verificando estado de Cloud SQL..."
STATE=$(gcloud sql instances describe "$INSTANCE_NAME" --format="value(state)" 2>/dev/null || echo "NOT_FOUND")

if [ "$STATE" != "RUNNABLE" ]; then
  echo "ERROR: Cloud SQL no está listo. Estado actual: $STATE"
  echo "Espera a que el estado sea RUNNABLE y vuelve a ejecutar este script."
  exit 1
fi

echo "✓ Cloud SQL está listo"
echo ""

# Crear base de datos
echo "Creando base de datos..."
gcloud sql databases create "$DB_NAME" \
  --instance="$INSTANCE_NAME" \
  --project="$PROJECT_ID" || echo "Base de datos puede que ya exista"

echo "✓ Base de datos creada"
echo ""

# Crear usuario
echo "Creando usuario de base de datos..."
gcloud sql users create "$DB_USER" \
  --instance="$INSTANCE_NAME" \
  --password="$DB_PASSWORD" \
  --project="$PROJECT_ID" || echo "Usuario puede que ya exista"

echo "✓ Usuario creado"
echo ""

# Conectar Cloud Run a Cloud SQL
echo "Conectando Cloud Run a Cloud SQL..."
gcloud run services update "$SERVICE_NAME" \
  --project="$PROJECT_ID" \
  --region="$REGION" \
  --add-cloudsql-instances="$PROJECT_ID:$REGION:$INSTANCE_NAME"

echo "✓ Conexión Cloud Run → Cloud SQL configurada"
echo ""

# Configurar variables de entorno (ajustar valores según necesidad)
echo "Configurando variables de entorno en Cloud Run..."
echo "NOTA: Ajusta X_API_KEY y otros valores según tu configuración"

gcloud run services update "$SERVICE_NAME" \
  --project="$PROJECT_ID" \
  --region="$REGION" \
  --update-env-vars="GO_ENVIRONMENT=production,DEPLOY_ENV=prod,DEPLOY_PLATFORM=gcp,APP_NAME=ponti-api,APP_VERSION=1.0,APP_MAX_RETRIES=5,X_API_KEY=<CAMBIAR_API_KEY>,API_VERSION=v1,HTTP_SERVER_NAME=http-server,HTTP_SERVER_HOST=0.0.0.0,DB_TYPE=postgres,DB_USER=$DB_USER,DB_PASSWORD=$DB_PASSWORD,DB_HOST=/cloudsql/$PROJECT_ID:$REGION:$INSTANCE_NAME,DB_NAME=$DB_NAME,DB_SSL_MODE=require,DB_PORT=5432,MIGRATIONS_DIR=file://migrations,WORDS_SUGGESTER_LIMIT=100,WORDS_SUGGESTER_THRESHOLD=0.3,REPORT_SCHEMA=v4_report"

echo ""
echo "=== Configuración completada ==="
echo ""
echo "Resumen:"
echo "- Proyecto: $PROJECT_ID"
echo "- Cloud SQL: $INSTANCE_NAME (IP privada)"
echo "- Base de datos: $DB_NAME"
echo "- Usuario: $DB_USER"
echo "- Cloud Run: $SERVICE_NAME"
echo ""
echo "IMPORTANTE: Actualiza X_API_KEY en las variables de entorno con un valor real"
echo ""
echo "Para verificar:"
echo "  gcloud run services describe $SERVICE_NAME --project=$PROJECT_ID --region=$REGION"
echo "  gcloud sql instances describe $INSTANCE_NAME --project=$PROJECT_ID"
