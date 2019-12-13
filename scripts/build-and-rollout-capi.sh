#!/usr/bin/env bash

set -ex

SCRIPT_DIR=$(dirname $0)
REPO_BASE_DIR="${SCRIPT_DIR}/.."

# Build the capi image and push it to minkube
${SCRIPT_DIR}/build.sh
docker push $(minikube ip):5000/capi

#capi
helm template "${SCRIPT_DIR}/.." --set-string system_domain=minikube.local -f "${SCRIPT_DIR}/capi-values.yaml" \
  | tee last-apply.yaml \
  | kapp -y deploy -a capi -f -
