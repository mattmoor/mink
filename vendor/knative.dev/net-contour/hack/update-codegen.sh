#!/usr/bin/env bash

# Copyright 2019 The Knative Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

export GO111MODULE=on
# If we run with -mod=vendor here, then generate-groups.sh looks for vendor files in the wrong place.
export GOFLAGS=-mod=

if [ -z "${GOPATH:-}" ]; then
  export GOPATH=$(go env GOPATH)
fi

REPO_ROOT=$(dirname ${BASH_SOURCE})/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd ${REPO_ROOT}; ls -d -1 $(dirname ${BASH_SOURCE})/../vendor/k8s.io/code-generator 2>/dev/null || echo ../code-generator)}

KNATIVE_CODEGEN_PKG=${KNATIVE_CODEGEN_PKG:-$(cd ${REPO_ROOT}; ls -d -1 $(dirname ${BASH_SOURCE})/../vendor/knative.dev/pkg 2>/dev/null || echo ../pkg)}

# Generate our own client for istio (otherwise injection won't work)
${CODEGEN_PKG}/generate-groups.sh "client,informer,lister" \
  knative.dev/net-contour/pkg/client github.com/projectcontour/contour/apis \
  "projectcontour:v1" \
  --go-header-file ${REPO_ROOT}/hack/boilerplate/boilerplate.go.txt

# Knative Injection (for istio)
${KNATIVE_CODEGEN_PKG}/hack/generate-knative.sh "injection" \
  knative.dev/net-contour/pkg/client github.com/projectcontour/contour/apis \
  "projectcontour:v1" \
  --go-header-file ${REPO_ROOT}/hack/boilerplate/boilerplate.go.txt

# Depends on generate-groups.sh to install bin/deepcopy-gen
${GOPATH}/bin/deepcopy-gen \
  -O zz_generated.deepcopy \
  --go-header-file ${REPO_ROOT}/hack/boilerplate/boilerplate.go.txt \
  -i knative.dev/net-contour/pkg/reconciler/contour/config

# Make sure our dependencies are up-to-date
${REPO_ROOT}/hack/update-deps.sh
