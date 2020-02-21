#!/usr/bin/env bash

set -ex

CF_FOR_K8s_DIR="${1-${HOME}/workspace/cf-for-k8s/}"
SCRIPT_DIR="$(dirname $(realpath $0))"
BASE_DIR="${SCRIPT_DIR}/.."

pushd "${BASE_DIR}"
	REF="$(git rev-parse head)"
popd

pushd "${CF_FOR_K8s_DIR}"
  tmp="$(mktemp)"
  ytt -f vendir.yml \
      -f "${BASE_DIR}/overlays/set-vendir-ref/" \
      -v ref="${REF}" > ${tmp}
  cat ${tmp} > vendir.yml
  vendir sync
popd
