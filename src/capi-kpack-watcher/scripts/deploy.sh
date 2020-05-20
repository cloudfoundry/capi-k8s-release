#!/bin/bash

set -eux

SCRIPT_DIR="$(dirname $0)"
TEMPLATES_DIR="${SCRIPT_DIR}/../templates/"

"${SCRIPT_DIR}/build.sh"

docker push "${IMAGE_DESTINATION}"
