#!/bin/bash

set -ex

SCRIPT_DIR="$(dirname $0)"
REPO_BASE_DIR="${SCRIPT_DIR}/../../../"

docker build -t cloudfoundry/capi-kpack-watcher "${REPO_BASE_DIR}/src/capi-kpack-watcher"
