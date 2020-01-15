#!/usr/bin/env bash

set -ex

SCRIPTS_DIR="$(dirname $0)"

"${SCRIPTS_DIR}/deploy.sh"

kubectl rollout restart capi-kpack-watcher
kubectl rollout status capi-kpack-watcher -w
