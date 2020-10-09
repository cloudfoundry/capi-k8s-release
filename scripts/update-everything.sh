#! /bin/bash

set -eux

pushd ~/workspace/capi-release/
  git co develop
  git pull
  ./scripts/update
  git status
popd

pushd ~/workspace/capi-release/src/cloud_controller_ng
  git co honeycomb
  git rebase master
  git status
popd

pushd ~/workspace/capi-k8s-release/
  git co master
  git pull
  git co honeycomb
  git rebase master
  git status
popd

pushd ~/workspace/cf-for-k8s/
  git co master
  git pull
popd
