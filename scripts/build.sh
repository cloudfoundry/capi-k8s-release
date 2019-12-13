#!/usr/bin/env bash

SCRIPT_DIR="$(dirname $0)"
REPO_BASE_DIR="${SCRIPT_DIR}/.."

# Build the capi image and push it to minkube
docker build -f "${REPO_BASE_DIR}/dockerfiles/cloud_controller_ng/Dockerfile" -t capi -t $(minikube ip):5000/capi "${REPO_BASE_DIR}/src/"
