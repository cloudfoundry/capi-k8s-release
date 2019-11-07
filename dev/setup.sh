#!/bin/bash

set -ex

# minikube start -p capi
# minikube addons enable registry



if ! jq '.insecure_registries' $HOME/.docker/daemon.json -e; then
  echo "no insecure registries"
else
  echo "it's there"
fi


#.insecure_registries == null or select(.insecure_registries | contains(["234"]) | not)
