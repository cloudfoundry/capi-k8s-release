#!/usr/bin/env bash

set -ex

SCRIPT_DIR=$(dirname $0)
REPO_BASE_DIR="${SCRIPT_DIR}/.."

helm repo add stable https://kubernetes-charts.storage.googleapis.com

cat <<EOF | kubectl apply -f -
---
apiVersion: v1
kind: Namespace
metadata:
  name: cf-system
EOF

#database
helm -n cf-system upgrade --install capi-database stable/postgresql -f "${SCRIPT_DIR}/postgresql-values.yaml"
#minio
helm -n cf-system upgrade --install capi-blobstore stable/minio


${SCRIPT_DIR}/rollout.sh "$@"

