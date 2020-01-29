#!/bin/bash

set -eu

check_installed() {
  if ! command -v "${1}" > /dev/null 2>&1; then
    echo "Please install ${1} first"
    exit 1
  fi
}

check_installed minikube
check_installed helm

minikube addons enable registry
minikube addons enable helm-tiller

minikube start
