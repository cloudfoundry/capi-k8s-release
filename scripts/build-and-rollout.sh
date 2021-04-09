#!/usr/bin/env bash

set -ex

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_BASE_DIR="${SCRIPT_DIR}/.."

# additional environment variable overrides are available in this other script
${SCRIPT_DIR}/build-into-values.sh "${REPO_BASE_DIR}/config/values/images.yml"

${SCRIPT_DIR}/rollout.sh "$@"
