#!/bin/bash

set -ex

SCRIPT_DIR="$(dirname $0)"
TEMPLATES_DIR="${SCRIPT_DIR}/../templates/"

"${SCRIPT_DIR}/build.sh"
docker push gcr.io/cf-capi-arya/dev-capi-kpack-watcher

kubectl apply -f "${TEMPLATES_DIR}"
