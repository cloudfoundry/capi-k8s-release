#!/bin/bash
# Copyright 2012 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# git goimports pre-commit hook
#
# To use, store as .git/hooks/pre-commit inside your repository and make sure
# it has execute permissions.
#
# This script does not handle file names that contain spaces.

gofiles=$(git diff --cached --name-only --diff-filter=ACM | grep -v vendor | grep '.go$')
[ -z "${gofiles}" ] && exit 0

unformatted=$(goimports -l $gofiles)
[ -z "${unformatted}" ] && exit 0

echo >&2 "Go files must be formatted with goimports. Please run:"
for fn in ${unformatted}; do
    echo >&2 "  goimports -w ${PWD}/${fn}"
done

exec < /dev/tty

read -r -p "Autocorrect (Y/n)? " yn
case "${yn}" in
Y|y)
    for fn in ${unformatted}; do
        goimports -w "${PWD}/${fn}" && echo "Formatted ${PWD}/${fn} ..."
    done
    ;;
*)
    exit 1
    ;;
esac

git add -u ${unformatted}

exit 0
