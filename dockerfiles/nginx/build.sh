#!/usr/bin/env bash

set -euo pipefail

function main() {
  echo "Downloading latest capi-release..."
  wget -q -O capi-release.tgz https://bosh.io/d/github.com/cloudfoundry/capi-release

  echo "Extracting nginx.tgz..."
  tar xf capi-release.tgz packages/nginx.tgz

  echo "Unpacking nginx.tgz..."
  tar xf packages/nginx.tgz

  echo "Detecting tarballs..."
  set +e
  nginx_version="$(ls -1 nginx | grep -E 'nginx-[0-9]+\.[0-9]+\.[0-9]+\.tar\.gz' | xargs -n1 -I{} basename {} .tar.gz)"

  if [[ -n "${nginx_version}" ]]; then
    echo "Found: ${nginx_version}"
    tar xf "nginx/${nginx_version}.tar.gz"
  else
    echo "ERROR: Failed to detect nginx tarball"
    exit 1
  fi

  nginx_upload_module_version="$(ls -1 nginx | grep -E 'nginx-upload-module-[0-9]+\.[0-9]+\.[0-9]+\.tar\.gz' | xargs -n1 -I{} basename {} .tar.gz)"

  if [[ -n "${nginx_upload_module_version}" ]]; then
    echo "Found: ${nginx_upload_module_version}"
    tar xf "nginx/${nginx_upload_module_version}.tar.gz"
  else
    echo "ERROR: Failed to detect nginx-upload-module tarball"
    exit 1
  fi

  pcre_version="$(ls -1 nginx | grep -E 'pcre-[0-9]+\.[0-9]+\.tar\.gz' | xargs -n1 -I{} basename {} .tar.gz)"

  if [[ -n "${pcre_version}" ]]; then
    echo "Found: ${pcre_version}"
    tar xf "nginx/${pcre_version}.tar.gz"
  else
    echo "ERROR: Failed to detect pcre tarball"
    exit 1
  fi
  set -e

  echo "Patching nginx-upload-module..."
  pushd "${nginx_upload_module_version}"
    patch < ../nginx/upload_module_put_support.patch
  popd

  echo "Building nginx..."
  pushd "${nginx_version}"
    ./configure \
      --with-debug \
      --with-pcre="../${pcre_version}" \
      --add-module="../${nginx_upload_module_version}" && \
    make -j $(nproc) && \
    make -j $(nproc) install
  popd
}

main "$@"