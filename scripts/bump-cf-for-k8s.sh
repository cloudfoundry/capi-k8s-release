#!/usr/bin/env bash

set -ex

CF_FOR_K8s_DIR="${1-${HOME}/workspace/cf-for-k8s/}"
SCRIPT_DIR="$(dirname $(realpath $0))"
BASE_DIR="${SCRIPT_DIR}/.."

pushd "${CF_FOR_K8s_DIR}"
  vendir sync --use-directory config/_ytt_lib/github.com/cloudfoundry/capi-k8s-release="${BASE_DIR}"
popd
