#!/usr/bin/env bash

set -euo pipefail

if [ "$#" -lt 1 ]; then
   echo "Usage: get-backup-metadata veleroBackupName [file]"
   echo "file - file name where metadata is output to"
   exit 1
fi

if ! command -v velero &> /dev/null
then
    echo "velero is not installed. Please install velero and re-run this script."
    exit 1
fi

if [ -z "${2:-}" ]; then
   velero backup logs "$1" | grep -o 'CF Metadata: {.*}\\n'  \
     | sed 's/CF Metadata: //g' \
     | sed 's/\\n//g' \
     | sed 's/\\//g' \
     | jq .
else
   velero backup logs "$1" | grep -o 'CF Metadata: {.*}\\n'  \
     | sed 's/CF Metadata: //g' \
     | sed 's/\\n//g' \
     | sed 's/\\//g' \
     | jq . > "$2"
fi

