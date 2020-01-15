#!/bin/bash

set -ex

SCRIPT_DIR="$(dirname $0)"
REPO_BASE_DIR="${SCRIPT_DIR}/../../../"

docker build -f "${REPO_BASE_DIR}/dockerfiles/capi-kpack-watcher/Dockerfile" -t capi-kpack-watcher -t $(minikube ip):5000/capi-kpack-watcher "${REPO_BASE_DIR}/src/"
