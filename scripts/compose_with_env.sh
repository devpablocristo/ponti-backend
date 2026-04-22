#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

# shellcheck disable=SC1091
source "${ROOT_DIR}/scripts/lib/backend_env.sh"
load_backend_env "${ROOT_DIR}"

exec docker compose -f "${ROOT_DIR}/docker-compose.yml" "$@"
