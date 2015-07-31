#!/bin/bash

set -e

THIS_SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
if [ -z "${REPO_ROOT_DIR}" ] ; then
    export REPO_ROOT_DIR="${THIS_SCRIPT_DIR}/.."
fi

echo "=> Switching to REPO_ROOT_DIR: ${REPO_ROOT_DIR}"
cd "${REPO_ROOT_DIR}"
export PATH="$PATH:$GOPATH/bin"

#
# Script for Continuous Integration
#

set -v

# Intsall
go install

# Check for unhandled errors
go get -u github.com/kisielk/errcheck
go install github.com/kisielk/errcheck

# Go lint
go get -u github.com/golang/lint/golint

bash "${THIS_SCRIPT_DIR}/test.sh"

#
# ==> DONE - OK
#
