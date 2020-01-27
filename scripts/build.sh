#!/usr/bin/env bash

SCRIPT_DIR="$(dirname $0)"
PARENT_DIR="${SCRIPT_DIR}/.."

pushd ${PARENT_DIR}
  docker build src/ -f dockerfiles/cloud_controller_ng/Dockerfile -t localhost:5000/capi
popd
