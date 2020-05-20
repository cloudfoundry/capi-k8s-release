#!/bin/bash

set -eux

SCRIPT_DIR="$(dirname $0)"
REPO_BASE_DIR="${SCRIPT_DIR}/../../../"

pack build "${IMAGE_DESTINATION}" --builder cloudfoundry/cnb:bionic
