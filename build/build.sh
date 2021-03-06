#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
KBLD_CONFIG_DIR="$(mktemp -d)"
KBLD_LOCK_FILE="${SCRIPT_DIR}/kbld.lock.yml"

function cleanup() {
  echo "Cleaning up..."
  rm -rf "${KBLD_CONFIG_DIR}"
}

trap cleanup EXIT

if [ "$#" -lt 1 ]; then
  echo "Error: You must provide at least one component directory to build"
  exit 1
fi

pushd "${SCRIPT_DIR}" > /dev/null
  components=( "$@" )
  # ensure input components are actually valid
  for component in "${components[@]}"; do
    if [ ! -d "${SCRIPT_DIR}/${component}" ]; then
      echo "Error: The following component directory does not exist: ${component}"
      echo "Usage: build.sh [component ...]"
      exit 1
    fi
  done

  vendir sync

  # build each input component
  for component in "${components[@]}"; do
    mkdir "${KBLD_CONFIG_DIR}/${component}"

    # use component specific kbld generation if available, else fall back to default generation script
    if [[ -f "${SCRIPT_DIR}/${component}/generate-kbld-config.sh" ]]; then
      "${SCRIPT_DIR}/${component}/generate-kbld-config.sh" "${KBLD_CONFIG_DIR}/${component}/kbld.yml"
    else
      "${SCRIPT_DIR}/default-generate-kbld-config.sh" "${KBLD_CONFIG_DIR}/${component}/kbld.yml" "${SCRIPT_DIR}/${component}/kbld.yml"
    fi

  done

  kbld -f "${KBLD_CONFIG_DIR}" -f <(ytt -f "${SCRIPT_DIR}/../config/" -f "${SCRIPT_DIR}/sample-capi-install-values.yml") --lock-output "${KBLD_LOCK_FILE}"
popd > /dev/null
