#!/usr/bin/env bash

SCRIPT_DIR=$(dirname $0)
REPO_BASE_DIR="${SCRIPT_DIR}/.."

helm template "${SCRIPT_DIR}/.." --set-string system_domain=minikube.local -f "${SCRIPT_DIR}/capi-values.yaml" \
  | tee last-apply.yaml \
  | kapp -y deploy -a capi -f -
