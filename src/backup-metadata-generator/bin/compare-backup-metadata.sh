#!/usr/bin/env bash

set -euo pipefail

if [ "$#" -lt 1 ]; then
      echo "Usage: compare-backup-metadata file [namespace]"
      echo "file - backup metadata json obtained from get-backup-metadata tool"
      exit 1
fi

if [ -z "${2:-}" ]; then
      NAMESPACE="";
else
      NAMESPACE="-n $2";
      kubectl get namespace "$2" || exit 1
fi

POD_NAME=$(kubectl $NAMESPACE get pods -l 'app.kubernetes.io/name=cf-api-controllers' -o jsonpath='{.items[0].metadata.name}' 2> /dev/null)
kubectl ${NAMESPACE} exec -i "${POD_NAME}" -c backup-metadata-generator -- bash -ce '/cnb/lifecycle/launcher backup-metadata-generator compare' < "$1"
