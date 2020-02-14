#!/usr/bin/env bash

pushd "$(dirname "$0")"
  ln -sf \
    "$(git rev-parse --show-toplevel)"/src/capi-kpack-watcher/scripts/goimports-pre-commit-hook.sh \
    "$(git rev-parse --git-dir)"/hooks/pre-commit
popd
