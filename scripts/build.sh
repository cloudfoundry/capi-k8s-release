#!/usr/bin/env bash

SCRIPT_DIR="$(dirname $0)"
PARENT_DIR="${SCRIPT_DIR}/.."

pushd ${PARENT_DIR}
  docker build src/ -f dockerfiles/cloud_controller_ng/Dockerfile -t capi
  docker build src/ -f dockerfiles/capi-kpack-watcher/Dockerfile -t capi-kpack-watcher
popd
