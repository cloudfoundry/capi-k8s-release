#!/usr/bin/env bash

set -euo pipefail

function usage_text() {
  cat <<EOF
Usage:
  $(basename "$0") OLD_TAG NEW_TAG
EOF
  exit 1
}

if [[ $# -lt 1 ]]; then
  usage_text >&2
fi

OLD_VERSION="$1"
NEW_VERSION="$2"

pushd "${HOME}/workspace/git_repo_changelog" > /dev/null
  echo "🚀 Generating release notes for capi-k8s-release... 🚀" 1>&2
  bundle exec rake "changelog[${HOME}/workspace/capi-k8s-release,${OLD_VERSION},${NEW_VERSION},]" \
    2>/dev/null | tail -n+2 | sort | uniq

  pushd "${HOME}/workspace/capi-k8s-release" > /dev/null
    git checkout ${OLD_VERSION} 2>/dev/null
    old_ccng_image_digest="$(bosh int --path=/images/ccng values/images.yml)"
    old_ccng_sha="$(git log --grep="${old_ccng_image_digest}" | rg "cloud_controller_ng" -A1 | tail -n+2)"

    git checkout ${NEW_VERSION} 2>/dev/null
    new_ccng_image_digest="$(bosh int --path=/images/ccng values/images.yml)"
    new_ccng_sha="$(git log --grep="${new_ccng_image_digest}" | rg "cloud_controller_ng" -A1 | tail -n+2)"
  popd > /dev/null

  echo ""
  echo "🚀 Generating release notes for cloud_controller_ng... 🚀" 1>&2
  # filters out duplicates and any failures to fetch stories
  # NOTE: if you're not authorized properly, this could mask that (i.e. gave a valid tracker token
  # but don't have permissions on our project)
  bundle exec rake \
    "changelog[${HOME}/workspace/capi-release/src/cloud_controller_ng,${old_ccng_sha},${new_ccng_sha},]" \
    2>/dev/null | tail -n+2 | sort | uniq | rg -v "failed to fetch story title"
popd > /dev/null
