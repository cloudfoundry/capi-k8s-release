#!/usr/bin/env bash

set -ex

SCRIPT_DIR=$(dirname $0)
REPO_BASE_DIR="${SCRIPT_DIR}/.."

eval $(minikube docker-env)

${SCRIPT_DIR}/rollout.sh "$@"


