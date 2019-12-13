#!/bin/bash

set -ex

SCRIPT_DIR=$(dirname $0)
REPO_BASE_DIR="${SCRIPT_DIR}/.."

#database
helm upgrade --install capi-database stable/postgresql -f "${SCRIPT_DIR}/postgresql-values.yaml"
#minio
helm upgrade --install capi-blobstore stable/minio

# Build the capi image, push it to minikube, kapp deploy
${SCRIPT_DIR}/build-and-rollout-capi.sh
