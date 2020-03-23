#!/usr/bin/env bash

set -ex

CF_FOR_K8s_DIR="${CF_FOR_K8s_DIR:-${HOME}/workspace/cf-for-k8s/}"
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_BASE_DIR="${SCRIPT_DIR}/.."

if [ -z "$1" ]
  then
    echo "No values file was supplied!"
		exit 1
fi

if [ -z "$2" ]
  then
    echo "No image destination was supplied!"
		exit 1
fi

# template the image destination into the kbld yml
KBLD_TMP="$(mktemp)"
ytt -f "${REPO_BASE_DIR}/dev-templates/" -v kbld.destination="${2}" > "${KBLD_TMP}"

# build a new values file with kbld
pushd "${REPO_BASE_DIR}"
  VALUES_TMP="$(mktemp)"
  echo "#@data/values" > "${VALUES_TMP}"
  kbld -f "${KBLD_TMP}" -f "$1" >> "${VALUES_TMP}"
popd

cat "${VALUES_TMP}" > "$1"
