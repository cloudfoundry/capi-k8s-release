#!/usr/bin/env bash

set -ex

SCRIPT_DIR=$(dirname $0)
REPO_BASE_DIR="${SCRIPT_DIR}/.."
IMAGE_DESTINATION="${IMAGE_DESTINATION:-gcr.io/cf-capi-arya/dev-ccng}"

${SCRIPT_DIR}/build-into-values.sh "${REPO_BASE_DIR}/values.yml" "${IMAGE_DESTINATION}"

${SCRIPT_DIR}/rollout.sh "$@"


