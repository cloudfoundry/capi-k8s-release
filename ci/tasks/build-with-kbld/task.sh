#!/bin/bash
set -euo pipefail

trap "pkill dockerd" EXIT

start-docker &
echo 'until docker info; do sleep 5; done' >/usr/local/bin/wait_for_docker
chmod +x /usr/local/bin/wait_for_docker
timeout 300 wait_for_docker

<<<"$DOCKERHUB_PASSWORD" docker login --username "$DOCKERHUB_USERNAME" --password-stdin

pushd capi-k8s-release/build > /dev/null
  ./build.sh $COMPONENT
  image_ref="$(yq -r ".overrides[] | select(.image | test(\"/${IMAGE_NAME}@\")).newImage" kbld.lock.yml)"
popd > /dev/null

docker pull "$image_ref"
docker save "$image_ref" -o kbld-output/image.tar
