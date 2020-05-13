#!/usr/bin/env bash

set -ex

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_BASE_DIR="${SCRIPT_DIR}/.."
: "${@?No values file supplied.}"

${SCRIPT_DIR}/build-into-values.sh "${REPO_BASE_DIR}/values/images.yml" "${IMAGE_DESTINATION}"

${SCRIPT_DIR}/rollout.sh "$@"


