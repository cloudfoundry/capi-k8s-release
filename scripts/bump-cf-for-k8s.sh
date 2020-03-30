#!/usr/bin/env bash

set -ex

CF_FOR_K8s_DIR="${CF_FOR_K8s_DIR:-${HOME}/workspace/cf-for-k8s/}"
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
BASE_DIR="${SCRIPT_DIR}/.."

pushd "${CF_FOR_K8s_DIR}"
  vendir sync -d config/_ytt_lib/github.com/cloudfoundry/capi-k8s-release="${BASE_DIR}"
popd
