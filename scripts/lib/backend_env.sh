#!/usr/bin/env bash

backend_root_dir() {
  local script_dir
  script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  cd "${script_dir}/../.." && pwd
}

resolve_backend_env_file() {
  local root_dir="$1"
  local env_file="$2"

  if [[ -z "${env_file}" ]]; then
    printf "%s/.env\n" "${root_dir}"
    return 0
  fi

  if [[ "${env_file}" = /* ]]; then
    printf "%s\n" "${env_file}"
    return 0
  fi

  printf "%s/%s\n" "${root_dir}" "${env_file}"
}

resolve_backend_profile_env_file() {
  local root_dir="$1"
  local profile="$2"

  case "${profile}" in
    "" )
      return 1
      ;;
    stg|main )
      printf "%s/.env.main.local\n" "${root_dir}"
      ;;
    dev|develop )
      printf "%s/.env.develop.local\n" "${root_dir}"
      ;;
    * )
      echo "ERROR: PONTI_LOCAL_ENV inválido: ${profile}. Usá stg|main o dev|develop." >&2
      return 1
      ;;
  esac
}

load_backend_env() {
  local root_dir="${1:-$(backend_root_dir)}"
  local base_env="${root_dir}/.env"
  local override_env_raw="${PONTI_ENV_FILE:-}"
  local profile_env_raw="${PONTI_LOCAL_ENV:-}"
  local override_env=""

  if [[ ! -f "${base_env}" ]]; then
    echo "ERROR: falta archivo base de entorno ${base_env}" >&2
    return 1
  fi

  if [[ -n "${override_env_raw}" ]]; then
    override_env="$(resolve_backend_env_file "${root_dir}" "${override_env_raw}")"
    if [[ ! -f "${override_env}" ]]; then
      echo "ERROR: falta archivo de override ${override_env}" >&2
      return 1
    fi
  elif [[ -n "${profile_env_raw}" ]]; then
    override_env="$(resolve_backend_profile_env_file "${root_dir}" "${profile_env_raw}")"
    if [[ ! -f "${override_env}" ]]; then
      echo "ERROR: falta archivo de override para perfil ${profile_env_raw}: ${override_env}" >&2
      return 1
    fi
  fi

  set -a
  # shellcheck disable=SC1090
  source "${base_env}"
  if [[ -n "${override_env}" ]]; then
    # shellcheck disable=SC1090
    source "${override_env}"
  fi
  set +a

  export PONTI_ROOT_DIR="${root_dir}"
  export PONTI_BASE_ENV_FILE="${base_env}"
  export PONTI_EFFECTIVE_ENV_FILE="${override_env:-${base_env}}"
}
