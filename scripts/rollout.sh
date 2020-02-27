#!/usr/bin/env bash

set -ex

CF_FOR_K8s_DIR="${CF_FOR_K8s_DIR:-${HOME}/workspace/cf-for-k8s/}"
SCRIPT_DIR=$(dirname $0)
REPO_BASE_DIR="${SCRIPT_DIR}/.."

${SCRIPT_DIR}/bump-cf-for-k8s.sh
${CF_FOR_K8s_DIR}/bin/install-cf.sh $@
