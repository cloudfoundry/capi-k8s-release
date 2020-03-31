#!/bin/bash

set -ex

SCRIPT_DIR="$(dirname $0)"
REPO_BASE_DIR="${SCRIPT_DIR}/../../../"

docker build -f "${REPO_BASE_DIR}/dockerfiles/capi-kpack-watcher/Dockerfile" -t gcr.io/cf-capi-arya/dev-capi-kpack-watcher "${REPO_BASE_DIR}/src/"
