#!/usr/bin/env bash

# Copyright 2020 VMware, Inc.
# SPDX-License-Identifier: Apache-2.0

readonly ROOT_DIR=$(dirname $0)/..
source ${ROOT_DIR}/vendor/knative.dev/test-infra/scripts/library.sh

set -o errexit
set -o nounset
set -o pipefail

cd ${ROOT_DIR}

# The list of dependencies that we track at HEAD and periodically
# float forward in this repository.
FLOATING_DEPS=(
  "knative.dev/pkg"
  "knative.dev/eventing"
  "knative.dev/test-infra"
)

# Parse flags to determine any we should pass to dep.
DEP_FLAGS=()
while [[ $# -ne 0 ]]; do
  parameter=$1
  case ${parameter} in
    --upgrade) DEP_FLAGS=( -update ${FLOATING_DEPS[@]} ) ;;
    *) abort "unknown option ${parameter}" ;;
  esac
  shift
done
readonly DEP_FLAGS

# Ensure we have everything we need under vendor/
dep ensure ${DEP_FLAGS[@]}

rm -rf $(find vendor/ -name 'OWNERS')
rm -rf $(find vendor/ -name '*_test.go')
