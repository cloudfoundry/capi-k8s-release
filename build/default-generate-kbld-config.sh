#! /bin/bash

set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

function generate_kbld_config() {
  local default_template_path="${SCRIPT_DIR}/kbld.yml"
  local kbld_config_path="${1}"
  local kbld_template_path="${2:-$default_template_path}"

  local source_path
  source_path="${SCRIPT_DIR}/.."

  pushd "${source_path}" > /dev/null
    local git_ref
    git_ref=$(git rev-parse HEAD)
  popd > /dev/null

  echo "Creating CAPI K8s release kbld config with ytt..."
  local kbld_config_values
  kbld_config_values=$(cat <<EOF
#@data/values
---
git_ref: ${git_ref}
git_url: https://github.com/cloudfoundry/capi-k8s-release
EOF
)

  echo "${kbld_config_values}" | ytt -f "${kbld_template_path}" -f - > "${kbld_config_path}"
}

function main() {
  local kbld_config_path="${1}"
  local kbld_template_path="${2}"

  generate_kbld_config "${kbld_config_path}" "${kbld_template_path}"
}

main "$@"
