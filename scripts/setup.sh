#!/bin/bash

set -eu

check_installed() {
  if ! command -v "${1}" > /dev/null 2>&1; then
    echo "Please install ${1} first"
    exit 1
  fi
}

check_installed minikube
check_installed helm

DAEMON_JSON="$HOME/.docker/daemon.json"

# `read` returns non-zero when it reaches EOF, so disable error checking.
set +e
read -r -d '' NEW_DAEMON_JSON <<EOF
{
  "insecure-registries": ["$(minikube ip):5000"]
}
EOF
set -e

TEMPFILE="$(mktemp)"

if ! jq -e '.insecure\-registries' "${DAEMON_JSON}" > /dev/null 2>&1; then
  cp "${DAEMON_JSON}" "${DAEMON_JSON}".bak

  # Combine "insecure_registries" key into daemon.json
  jq -s '.[0] * .[1]' "${DAEMON_JSON}" <(echo "${NEW_DAEMON_JSON}")  2> /dev/null > "${TEMPFILE}"
  mv "${TEMPFILE}" "${DAEMON_JSON}"

  echo "Added the following JSON key to ${DAEMON_JSON}:"
  echo "${NEW_DAEMON_JSON}"
  echo "Please restart your Docker daemon"
  echo "If something has gone wrong, a backup file was created in $HOME/.docker/"
else
  echo "${DAEMON_JSON} already contains insecure_registries"
  echo "You're good to go"
fi
