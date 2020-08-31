#!/usr/bin/env bash

set -euo pipefail

if [[ $# -lt 1 ]]; then
  echo "Usage $(basename "$0") <path-to-install-values-yaml>"
  exit 1
fi

install_values_path="$1"

app_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)

kapp deploy -y -a cf-metadata -f <( \
    ytt \
    -f "${app_dir}/config" \
    -f "${install_values_path}"
)
