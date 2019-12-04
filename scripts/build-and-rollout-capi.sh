#!/usr/bin/env bash

# Build the capi image and push it to minkube
docker build -f dockerfiles/cloud_controller_ng/Dockerfile -t $(minikube ip):5000/capi src/
docker push $(minikube ip):5000/capi

# Restart the capi deployment with the new image and wait until the restart is complete
kubectl rollout restart deployment/capi-api-server deployment/capi-worker
kubectl rollout status deployment/capi-api-server -w
kubectl rollout status deployment/capi-worker -w
