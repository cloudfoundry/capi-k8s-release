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

test $TRACKER_API_TOKEN

OLD_VERSION="$1"
NEW_VERSION="$2"

pushd "${HOME}/workspace/git_repo_changelog" > /dev/null
  echo "🚀 Generating release notes for capi-k8s-release... 🚀" 1>&2
  bundle exec rake "changelog[${HOME}/workspace/capi-k8s-release,${OLD_VERSION},${NEW_VERSION},]" \
    2>/dev/null | tail -n+2 | sort | uniq

  pushd "${HOME}/workspace/capi-k8s-release" > /dev/null
    git checkout ${OLD_VERSION} 2>/dev/null
    OLD_CCNG_IMAGE_DIGEST="$(bosh int --path=/images/ccng config/values/images.yml)"
    OLD_CCNG_SHA="$(git log --grep="${OLD_CCNG_IMAGE_DIGEST}" | rg "cloud_controller_ng" -A1 | tail -n+2)"

    git checkout ${NEW_VERSION} 2>/dev/null
    NEW_CCNG_IMAGE_DIGEST="$(bosh int --path=/images/ccng config/values/images.yml)"
    NEW_CCNG_SHA="$(git log --grep="${NEW_CCNG_IMAGE_DIGEST}" | rg "cloud_controller_ng" -A1 | tail -n+2)"
  popd > /dev/null

  pushd "${HOME}/workspace/capi-release/src/cloud_controller_ng" > /dev/null
    git checkout ${NEW_CCNG_SHA} 2>/dev/null
    MIGRATIONS=($(git diff --diff-filter=A --name-only $OLD_CCNG_SHA db/migrations))
  popd > /dev/null

  echo ""
  echo "🚀 Generating release notes for cloud_controller_ng... 🚀" 1>&2
  # filters out duplicates and any failures to fetch stories
  # NOTE: if you're not authorized properly, this could mask that (i.e. gave a valid tracker token
  # but don't have permissions on our project)
  CCNG_NOTES="$(bundle exec rake \
    "changelog[${HOME}/workspace/capi-release/src/cloud_controller_ng,${OLD_CCNG_SHA},${NEW_CCNG_SHA},]" \
    2>/dev/null | tail -n+2 | sort | uniq | rg -v "failed to fetch story title")"
  # Repos we care about:
  # - cloudfoundry/cloud_controller_ng
  # - cloudfoundry/capi-k8s-release
  # - cloudfoundry/cli
  REPOS_REGEX="(?:cloudfoundry/cloud_controller_ng|cloudfoundry/capi-k8s-release|cloudfoundry/cli)\s\#\d+:"
  echo "${CCNG_NOTES}" | rg -v "${REPOS_REGEX}"

  echo ""
  echo "🚀 Generating database migrations list for cloud_controller_ng... 🚀" 1>&2
  MIGRATIONS_FORMATTED=()
  if [ ${#MIGRATIONS[@]} -eq '0' ]; then
    MIGRATIONS_FORMATTED+=("None")
  else
    for i in "${MIGRATIONS[@]}"
    do
      MIGRATIONS_FORMATTED+=("- [$(basename $i)](https://github.com/cloudfoundry/cloud_controller_ng/blob/$NEW_CCNG_SHA/$i)")
    done
  fi
  echo "${MIGRATIONS_FORMATTED[*]}"

  echo ""
  echo "🚀 Generating list of related PRs and Issues... 🚀" 1>&2
  echo "${CCNG_NOTES}" | rg "${REPOS_REGEX}"
popd > /dev/null
