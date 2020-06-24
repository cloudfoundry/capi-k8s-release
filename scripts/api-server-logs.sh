#!/usr/bin/env bash

set -x

kapp -a cf logs -m %server% -c %server% "$@"
