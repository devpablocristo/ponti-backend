#!/bin/bash
# Script para diagnosticar problemas de deploy en Cloud Run

set -e

PROJECT_ID="${1:-new-ponti-dev}"
SERVICE_NAME="${2:-ponti-backend-dev}"
REGION="${3:-us-central1}"

echo "🔍 Diagnosticando servicio: $SERVICE_NAME en proyecto $PROJECT_ID"
echo ""

echo "📋 1. Variables de entorno configuradas:"
echo "----------------------------------------"
gcloud run services describe "$SERVICE_NAME" \
  --project="$PROJECT_ID" \
  --region="$REGION" \
  --format="table(spec.template.spec.containers[0].env[].name,spec.template.spec.containers[0].env[].value)" || echo "Error obteniendo variables"

echo ""
echo "📊 2. Últimos logs del servicio:"
echo "----------------------------------------"
gcloud run services logs read "$SERVICE_NAME" \
  --project="$PROJECT_ID" \
  --region="$REGION" \
  --limit=50 || echo "Error obteniendo logs"

echo ""
echo "✅ Variables críticas que deberían estar configuradas:"
echo "  - GO_ENVIRONMENT=production"
echo "  - HTTP_SERVER_PORT=8080"
echo "  - DB_SCHEMA (configurado por workflow)"
echo "  - DB_HOST, DB_USER, DB_PASSWORD, DB_NAME, DB_PORT, DB_SSL_MODE"
echo "  - DEPLOY_ENV, DEPLOY_PLATFORM"
echo "  - X_API_KEY"
echo ""
