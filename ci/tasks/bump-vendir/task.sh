#!/bin/bash

tag="$(cat component-repo/.git/ref)"
echo "Updating vendir.yml to use ${tag}"
pushd capi-k8s-release/build > /dev/null
  sed -i "s|ref:.*|ref: $tag|" vendir.yml
  vendir sync
popd > /dev/null

cp -R capi-k8s-release/. vendir-bumped-capi-k8s-release
