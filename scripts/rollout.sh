#!/usr/bin/env bash

set -ex

SCRIPT_DIR=$(dirname $0)
REPO_BASE_DIR="${SCRIPT_DIR}/.."

if [ -n "$1" ]
then
	ytt -f "${REPO_BASE_DIR}/templates"  -f "${REPO_BASE_DIR}/values.yml" -f $1 | kapp -y deploy -a capi -f -
else
	ytt -f "${REPO_BASE_DIR}/templates"  -f "${REPO_BASE_DIR}/values.yml" | kapp -y deploy -a capi -f -
fi