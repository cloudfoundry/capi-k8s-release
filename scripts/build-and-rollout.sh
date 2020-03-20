#!/usr/bin/env bash

set -ex

SCRIPT_DIR=$(dirname $0)
REPO_BASE_DIR="${SCRIPT_DIR}/.."

if [ -z "$1" ]
  then
    echo "No image destination was supplied!"
    exit 1
fi

eval $(minikube docker-env)

${SCRIPT_DIR}/build-into-values.sh "${REPO_BASE_DIR}/values.yml" $1

${SCRIPT_DIR}/rollout.sh "$@"


