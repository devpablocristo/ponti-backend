#!/usr/bin/env bash
# Exporta las conversaciones de ponti-ai a un archivo JSON-lines listo para
# archivar en GCS o reimportar a Companion (decisión: archivar 6 meses
# post-cutover; ver plan en .claude/plans/fe-https-...md).
#
# Variables de entorno:
#   PONTI_AI_DSN  (default: postgres://postgres:postgres@localhost:15435/ai
#                  para local con axis docker-compose ya levantado)
#   OUT_DIR       (default: ./var/exports)
#
# Output:
#   <OUT_DIR>/ai-conversations-YYYYMMDD-HHMM.jsonl  (una conversation por línea)
#   <OUT_DIR>/ai-conversations-YYYYMMDD-HHMM.summary.txt  (count + size + sha256)
#
# Para subir a GCS post-export:
#   gsutil cp <out_file> gs://ponti-archives/ai-conversations/
#
# Pre-requisitos:
#   - psql cliente instalado.
#   - PONTI_AI_DSN apuntando a la DB de ponti-ai (no a la de ponti-backend).
#
# IMPORTANTE: este script es read-only. No borra la tabla. La decisión de drop
# se ejecuta SOLO post-cutover y con confirmación humana — ver runbook en el
# plan file.
set -euo pipefail

DSN="${PONTI_AI_DSN:-postgres://postgres:postgres@localhost:5436/ai?sslmode=disable}"
OUT_DIR="${OUT_DIR:-./var/exports}"
mkdir -p "$OUT_DIR"

TS="$(date +%Y%m%d-%H%M)"
OUT="$OUT_DIR/ai-conversations-$TS.jsonl"
SUMMARY="$OUT_DIR/ai-conversations-$TS.summary.txt"

echo "Exportando ai_conversations desde $DSN..."

# Cada fila se serializa como un JSON object completo (jsonb_build_object). Eso
# nos da una shape uniforme y portable sin tener que post-procesar.
psql "$DSN" -At -c "
  SELECT jsonb_build_object(
    'id', id,
    'project_id', project_id,
    'user_id', user_id,
    'mode', mode,
    'title', title,
    'messages', messages,
    'tool_calls_count', tool_calls_count,
    'tokens_input', tokens_input,
    'tokens_output', tokens_output,
    'created_at', created_at,
    'updated_at', updated_at
  )::text
  FROM ai_conversations
  ORDER BY created_at ASC
" > "$OUT"

COUNT=$(wc -l < "$OUT")
SIZE=$(du -h "$OUT" | cut -f1)
SHA=$(sha256sum "$OUT" | cut -d' ' -f1)

{
  echo "Archivo: $OUT"
  echo "Conversaciones: $COUNT"
  echo "Tamaño: $SIZE"
  echo "SHA-256: $SHA"
  echo "Fecha export: $(date -Iseconds)"
} > "$SUMMARY"

echo
cat "$SUMMARY"
echo
echo "Done. Verificar con:"
echo "  head -n1 $OUT | jq ."
echo "  wc -l $OUT"
