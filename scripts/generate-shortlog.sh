#!/usr/bin/env bash

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_BASE_DIR="${SCRIPT_DIR}/.."

CAPI_RELEASE_DIR="${CAPI_RELEASE_DIR:-${HOME}/workspace/capi-release/}"
CCNG_DIR="${CCNG_DIR:-${CAPI_RELEASE_DIR}src/cloud_controller_ng}"
CF_API_CONTROLLERS_DIR="${CAPI_CF_API_CONTROLLERS_DIR:-${REPO_BASE_DIR}/src/cf-api-controllers}"
REGISTRY_BUDDY_DIR="${REGISTRY_BUDDY_DIR:-${REPO_BASE_DIR}/src/registry-buddy}"
BACKUP_METADATA_GENERATOR_DIR="${BACKUP_METADATA_GENERATOR_DIR:-${REPO_BASE_DIR}/src/backup-metadata-generator}"
NGINX_DIR="${NGINX_DIR:-${REPO_BASE_DIR}/dockerfiles/nginx}"

function git_sha() {
  local repo=${1}
  pushd "${repo}" > /dev/null
    git rev-parse HEAD
  popd >/dev/null
}

function git_remote() {
  local repo=${1}
  pushd "${repo}" > /dev/null
    git remote get-url origin
  popd >/dev/null
}

function get_image() {
  local name=${1}
  pushd "${REPO_BASE_DIR}" >/dev/null
    yq -r ".images.${name}" config/values/images.yml
  popd >/dev/null
}

function github_commit_link() {
  local repo=${1}
  # strip "git@github" style urls and .git extensions, then append sha
  echo "https://github.$(git_remote "${dir}" | cut -f2 -d. | sed 's|:|/|g')/commit/$(git_sha "${repo}")"
}

function image_changed() {
  local name=${1}
  pushd "${REPO_BASE_DIR}" >/dev/null
    ! git log -n 1 --grep="^${name}:$" | grep "$(get_image "${name}")" > /dev/null
    local result=$?
  popd >/dev/null
  return ${result}
}

function chunk() {
  local name=${1}
  local dir=${2}
  if image_changed "${name}"; then
    cat <<- EOF
${name}:
  image: $(get_image "${name}")
  sha: $(git_sha "${dir}")
  remote: $(git_remote "${dir}")
  link: $(github_commit_link "${dir}")
EOF
  fi
}

cat <<- EOF
images.yml updated by CI
---
EOF
chunk ccng "${CCNG_DIR}"
chunk nginx "${NGINX_DIR}"
chunk cf_api_controllers "${CF_API_CONTROLLERS_DIR}"
chunk registry_buddy "${REGISTRY_BUDDY_DIR}"
chunk backup_metadata_generator "${BACKUP_METADATA_GENERATOR_DIR}"
