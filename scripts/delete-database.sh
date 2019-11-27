#!/bin/bash

set -e

RELEASE=${1:-capi-database}

helm delete $RELEASE --purge || true
kubectl delete pvc -l release=$RELEASE || true
