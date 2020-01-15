#!/usr/bin/env bash

set -ex

SCRIPTS_DIR="$(dirname $0)"

"${SCRIPTS_DIR}/deploy.sh"

kubectl rollout restart deployment/capi-kpack-watcher -n cf-system
kubectl rollout status deployment/capi-kpack-watcher -w -n cf-system
