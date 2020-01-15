#!/bin/bash

set -ex

SCRIPT_DIR="$(dirname $0)"
TEMPLATES_DIR="${SCRIPT_DIR}/../templates/"

"${SCRIPT_DIR}/build.sh"
docker push $(minikube ip):5000/capi-kpack-watcher

kubectl apply -f "${TEMPLATES_DIR}"
