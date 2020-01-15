#!/usr/bin/env bash

set -ex
SCRIPTS_DIR=$(dirname $0)
# Build the capi image and push it to minkube
${SCRIPTS_DIR}/deploy.sh

# Restart the capi deployment with the new image and wait until the restart is complete
CAPI_JOBS="deployment/capi-api-server deployment/capi-worker deployment/capi-clock deployment/capi-deployment-updater"
kubectl rollout restart ${CAPI_JOBS}
for job in ${CAPI_JOBS}; do
  kubectl rollout status "${job}" -w
done
