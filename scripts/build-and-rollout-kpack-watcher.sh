#!/usr/bin/env bash
set -x
docker build -f "dockerfiles/capi-kpack-watcher/Dockerfile" -t cloudfoundry/capi:kpack-watcher src/
docker push cloudfoundry/capi:kpack-watcher
kubectl rollout restart deployment/capi-kpack-watcher -n cf-system
kubectl rollout status deployment/capi-kpack-watcher -n cf-system -w

