#!/bin/bash

set -ex

SCRIPT_DIR=$(dirname $0)
REPO_BASE_DIR="${SCRIPT_DIR}/.."

helm repo add stable https://kubernetes-charts.storage.googleapis.com

#database
helm upgrade --install capi-database stable/postgresql -f "${SCRIPT_DIR}/postgresql-values.yaml"
#minio
helm upgrade --install capi-blobstore stable/minio

# Build the capi image and push it to minkube
docker build -f "${REPO_BASE_DIR}/dockerfiles/cloud_controller_ng/Dockerfile" -t capi -t $(minikube ip):5000/capi "${REPO_BASE_DIR}/src/"
docker build -f "${REPO_BASE_DIR}/dockerfiles/capi-kpack-watcher/Dockerfile" -t capi-kpack-watcher -t $(minikube ip):5000/capi-kpack-watcher "${REPO_BASE_DIR}/src/"
docker push $(minikube ip):5000/capi
docker push $(minikube ip):5000/capi-kpack-watcher

#capi
helm template "${SCRIPT_DIR}/.." --set-string system_domain=minikube.local -f "${SCRIPT_DIR}/capi-values.yaml" | kubectl apply -f -
