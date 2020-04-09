#!/usr/bin/env bash

# Copyright 2020 VMware, Inc.
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

REPO_ROOT=$(dirname ${BASH_SOURCE})/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd ${REPO_ROOT}; ls -d -1 $(dirname $0)/../vendor/k8s.io/code-generator 2>/dev/null || echo ../code-generator)}

KNATIVE_CODEGEN_PKG=${KNATIVE_CODEGEN_PKG:-$(cd ${REPO_ROOT}; ls -d -1 $(dirname $0)/../vendor/knative.dev/pkg 2>/dev/null || echo ../pkg)}

# generate the code with:
# --output-base    because this script should also be able to run inside the vendor dir of
#                  k8s.io/kubernetes. The output-base is needed for the generators to output into the vendor dir
#                  instead of the $GOPATH directly. For normal projects this can be dropped.
${CODEGEN_PKG}/generate-groups.sh "deepcopy,client,informer,lister" \
  github.com/vmware-tanzu/sources-for-knative/pkg/client github.com/vmware-tanzu/sources-for-knative/pkg/apis \
  "sources:v1alpha1" \
  --go-header-file ${REPO_ROOT}/hack/boilerplate/boilerplate.go.txt

# Knative Injection
${KNATIVE_CODEGEN_PKG}/hack/generate-knative.sh "injection" \
  github.com/vmware-tanzu/sources-for-knative/pkg/client github.com/vmware-tanzu/sources-for-knative/pkg/apis \
  "sources:v1alpha1" \
  --go-header-file ${REPO_ROOT}/hack/boilerplate/boilerplate.go.txt

# Make sure our dependencies are up-to-date
${REPO_ROOT}/hack/update-deps.sh
