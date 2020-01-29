#!/usr/bin/env bash

SCRIPT_DIR=$(dirname $0)
REPO_BASE_DIR="${SCRIPT_DIR}/.."

eval $(minikube docker-env)
${SCRIPT_DIR}/rollout.sh "$@" \
	-f dev-templates/ \
	-v kbld.destination="$(minikube ip)":5000/cloudfoundry/cloud-controller-ng
