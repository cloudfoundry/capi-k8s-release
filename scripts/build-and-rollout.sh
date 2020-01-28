#!/usr/bin/env bash

set -ex

SCRIPT_DIR=$(dirname $0)
REPO_BASE_DIR="${SCRIPT_DIR}/.."

eval $(minikube docker-env)

${SCRIPT_DIR}/build.sh

docker push localhost:5000/capi

if [ -n "$1" ]
then
	${SCRIPT_DIR}/rollout.sh $1 
else
	${SCRIPT_DIR}/rollout.sh
fi


