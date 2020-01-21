#!/usr/bin/env bash

set -ex

SCRIPT_DIR=$(dirname $0)
REPO_BASE_DIR="${SCRIPT_DIR}/.."

${SCRIPT_DIR}/build.sh

docker push $(minikube ip):5000/capi

${SCRIPT_DIR}/rollout.sh
