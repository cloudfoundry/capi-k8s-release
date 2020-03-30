#!/bin/bash

set -ex

SCRIPT_DIR="$(dirname $0)"
TEMPLATES_DIR="${SCRIPT_DIR}/../templates/"

"${SCRIPT_DIR}/build.sh"
docker push cloudfoundry/capi-kpack-watcher:dev

kubectl apply -f "${TEMPLATES_DIR}"
