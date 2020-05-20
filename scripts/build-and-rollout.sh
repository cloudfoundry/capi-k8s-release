#!/usr/bin/env bash

set -ex

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_BASE_DIR="${SCRIPT_DIR}/.."

# set defaults - any of these should be configrable from outside this script
IMAGE_DESTINATION="${IMAGE_DESTINATION:-gcr.io/cf-capi-arya/shared-dev-capi}"
CF_FOR_K8S_DIR="${CF_FOR_K8s_DIR:-${HOME}/workspace/cf-for-k8s}"
CCNG_DIR="${CCNG_DIR:-${REPO_BASE_DIR}/../capi-release/src/cloud_controller_ng}"
: "${@?No values file supplied.}"

# build parameters into the kbld configuration
KBLD_TMP="$(mktemp)"
ytt -f "${REPO_BASE_DIR}/dev-templates/" \
    -v kbld.destination="${IMAGE_DESTINATION}" \
    -v ccng_dir="${CCNG_DIR}" \
     > "${KBLD_TMP}"

${SCRIPT_DIR}/bump-cf-for-k8s.sh

# use kbld to process templated cf-for-k8s yml -
# it'll find cloudfoundry/cloud_controller_ng and replace it w/ kbld-built image
kapp deploy -y -a cf \
  -f <(kbld -f "${KBLD_TMP}" -f <(ytt -f "${CF_FOR_K8S_DIR}/config" -f "$@" ))
