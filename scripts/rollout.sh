#!/usr/bin/env bash

set -ex

SCRIPT_DIR=$(dirname $0)
REPO_BASE_DIR="${SCRIPT_DIR}/.."

ytt -f "${REPO_BASE_DIR}/templates"  -f "${REPO_BASE_DIR}/values.yml" | kapp -y deploy -a capi -f -
