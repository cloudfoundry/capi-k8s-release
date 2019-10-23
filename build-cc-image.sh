#!/usr/bin/env bash

docker build -f dockerfiles/cloud_controller_ng/Dockerfile -t capi src/
docker save capi -o capi.tar
