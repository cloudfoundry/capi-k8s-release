#!/usr/bin/env bash

set -ex

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_BASE_DIR="${SCRIPT_DIR}/.."

# set defaults - any of these should be configrable from outside this script
IMAGE_DESTINATION_CCNG="${IMAGE_DESTINATION_CCNG:-gcr.io/cf-capi-arya/dev-ccng}"
IMAGE_DESTINATION_KPACK_WATCHER="${IMAGE_DESTINATION_KPACK_WATCHER:-gcr.io/cf-capi-arya/dev-kpack-watcher}"
CF_FOR_K8s_DIR="${CF_FOR_K8s_DIR:-${HOME}/workspace/cf-for-k8s/}"
CAPI_RELEASE_DIR="${CAPI_RELEASE_DIR:-${HOME}/workspace/capi-release/}"
CCNG_DIR="${CAPI_RELEASE_DIR}src/cloud_controller_ng"
CAPI_KPACK_WATCHER_DIR="${CAPI_KPACK_WATCHER_DIR:-${REPO_BASE_DIR}/src/cf-api-kpack-watcher}"

if [ -z "$1" ]
  then
    echo "No values file was supplied!"
    exit 1
fi

# template the image destination into the kbld yml
KBLD_TMP="$(mktemp)"
ytt -f "${REPO_BASE_DIR}/dev-templates/" \
    -v kbld.destination.ccng="${IMAGE_DESTINATION_CCNG}" \
    -v kbld.destination.capi_kpack_watcher="${IMAGE_DESTINATION_KPACK_WATCHER}" \
    -v src_dirs.ccng="${CCNG_DIR}" \
    -v src_dirs.capi_kpack_watcher="${CAPI_KPACK_WATCHER_DIR}" \
     > "${KBLD_TMP}"

# build a new values file with kbld
pushd "${REPO_BASE_DIR}"
  VALUES_TMP="$(mktemp)"
  echo "#@data/values" > "${VALUES_TMP}"
  kbld --images-annotation=false -f "${KBLD_TMP}" -f "$1" >> "${VALUES_TMP}"
popd

cat "${VALUES_TMP}" > "$1"
