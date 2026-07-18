#!/usr/bin/env bash
set -Eeuo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

completed=()

print_summary() {
  local status="$1"
  echo
  echo "Ponti Axis smoke summary: ${status}"
  if [[ ${#completed[@]} -gt 0 ]]; then
    for item in "${completed[@]}"; do
      echo "  OK ${item}"
    done
  fi
}

on_error() {
  local exit_code="$?"
  print_summary "FAILED"
  echo "Stopped on first failing smoke. Exit code: ${exit_code}" >&2
  exit "${exit_code}"
}

trap on_error ERR

run_step() {
  local name="$1"
  local script="$2"
  echo
  echo "==> ${name}"
  "${SCRIPT_DIR}/${script}"
  completed+=("${name}")
}

run_step "Ponti onboarding" "onboard-ponti.sh"
run_step "Read-only capabilities" "smoke-ponti-axis-readonly.sh"
run_step "Draft action governance" "smoke-ponti-axis-draft-actions.sh"
run_step "Draft previews" "smoke-ponti-axis-draft-previews.sh"
run_step "Nexus-approved draft execution" "smoke-ponti-axis-nexus-approved-draft.sh"
run_step "Chat through Axis" "smoke-ponti-axis-chat.sh"

trap - ERR
print_summary "OK"
