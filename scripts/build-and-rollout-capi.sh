#!/usr/bin/env bash

# Build the capi image and push it to minkube
docker build -f dockerfiles/cloud_controller_ng/Dockerfile -t $(minikube ip):5000/capi src/
docker push $(minikube ip):5000/capi

# Restart the capi deployment with the new image and wait until the restart is complete
set -x
CAPI_JOBS="deployment/capi-api-server deployment/capi-worker deployment/capi-clock deployment/capi-deployment-updater"
kubectl rollout restart ${CAPI_JOBS}
for job in ${CAPI_JOBS}; do
  kubectl rollout status "${job}" -w
done
