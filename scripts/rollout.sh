#!/usr/bin/env bash

set -ex

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
CAPI_K8S_DIR="${SCRIPT_DIR}/.."
CF_FOR_K8S_DIR="${SCRIPT_DIR}/../../cf-for-k8s"

${SCRIPT_DIR}/bump-cf-for-k8s.sh

kapp deploy -a cf -f <(ytt -f "${CF_FOR_K8S_DIR}/config" -f "$@") -y