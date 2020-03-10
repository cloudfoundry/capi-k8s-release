#!/usr/bin/env bash

SCRIPT_DIR="$(dirname $0)"
PARENT_DIR="${SCRIPT_DIR}/.."

pushd ${PARENT_DIR}
  docker build src/cloud_controller_ng -f src/cloud_controller_ng/Dockerfile "$@"
popd
