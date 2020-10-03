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

readonly ROOT_DIR=$(dirname $0)/..
source ${ROOT_DIR}/vendor/knative.dev/test-infra/scripts/library.sh

set -o errexit
set -o nounset
set -o pipefail

cd ${ROOT_DIR}

# We need these flags for things to work properly.
export GO111MODULE=on

# This controls the release branch we track.
VERSION="release-0.17"

# The list of dependencies that we track at HEAD and periodically
# float forward in this repository.
FLOATING_DEPS=(
  "knative.dev/pkg@${VERSION}"
  "knative.dev/test-infra@${VERSION}"

  "knative.dev/serving@${VERSION}"
  "knative.dev/networking@${VERSION}"
  "knative.dev/caching@${VERSION}"
  "knative.dev/net-http01@${VERSION}"
  "knative.dev/net-contour@${VERSION}"

  "github.com/projectcontour/contour@release-1.4"

  "knative.dev/eventing@${VERSION}"

  "github.com/tektoncd/pipeline@v0.16.3"
  "github.com/tektoncd/cli@master"
)

# Parse flags to determine any we should pass to dep.
GO_GET=0
while [[ $# -ne 0 ]]; do
  parameter=$1
  case ${parameter} in
    --upgrade) GO_GET=1 ;;
    *) abort "unknown option ${parameter}" ;;
  esac
  shift
done
readonly GO_GET

if (( GO_GET )); then
  go get -d ${FLOATING_DEPS[@]}
fi


# Prune modules.
go mod tidy
go mod vendor

rm -rf $(find vendor/ -name 'OWNERS')
rm -rf $(find vendor/ -name '*_test.go')
rm -rf $(find vendor/knative.dev/ -type l)
rm -rf $(find vendor/github.com/tektoncd/ -type l)

# TODO(https://github.com/tektoncd/cli/issues/983): CLI isn't up to date on PKG...
sed -i 's/ConvertUp/ConvertTo/g' $(find ./vendor/github.com/tektoncd/cli -name '*.go')
sed -i 's/ConvertDown/ConvertFrom/g' $(find ./vendor/github.com/tektoncd/cli -name '*.go')

# Apply patch to contour
git apply ${ROOT_DIR}/vendor/knative.dev/net-contour/hack/contour.patch

function rewrite_knative_namespace() {
  sed -E 's@knative-(serving|eventing)@mink-system@g'
}

function rewrite_tekton_namespace() {
  sed 's@namespace: tekton-pipelines@namespace: mink-system@g'
}

function rewrite_contour_namespace() {
  sed 's@namespace: projectcontour@namespace: mink-system@g' | \
    sed 's@--namespace=projectcontour@--namespace=mink-system@g'
}

function rewrite_contour_image() {
  sed -E $'s@docker.io/projectcontour/contour:.+@ko://github.com/projectcontour/contour/cmd/contour@g'
}

function rewrite_annotation() {
  sed -E 's@(serving|eventing).knative.dev/release@knative.dev/release@g'
}

function rewrite_webhook() {
  sed 's@webhook.serving.knative.dev@webhook.mink.knative.dev@g' | \
    sed 's@name: eventing-webhook@name: webhook@g' | \
    sed 's@name: tekton-pipelines-webhook@name: webhook@g'
}

function rewrite_common() {
  local readonly INPUT="${1}"
  local readonly OUTPUT_DIR="${2}"

  cat "${INPUT}" | rewrite_knative_namespace | rewrite_tekton_namespace | rewrite_contour_namespace | rewrite_annotation | rewrite_webhook \
    | rewrite_contour_image > "${OUTPUT_DIR}/$(basename ${INPUT})"
}

function list_yamls() {
  find "$1" -type f -name '*.yaml'
}

function replace_text() {
  local readonly INPUT="${1}"
  local readonly ORIG="${2}"
  local readonly REPLACEMENT="${3}"

  sed -i "s/${ORIG}/${REPLACEMENT}/g" ${INPUT}
}

# Remove all of the imported yamls before we start to do our rewrites.
rm $(find config/ -type f | grep imported)

#################################################
#
#
#    Serving
#
#
#################################################

# Do a blanket copy of these resources
for x in $(list_yamls ./vendor/knative.dev/serving/config/core/resources); do
  rewrite_common "$x" "./config/core/200-imported/200-serving/100-resources"
done
for x in $(list_yamls ./vendor/knative.dev/serving/config/core/webhooks); do
  rewrite_common "$x" "./config/core/200-imported/200-serving/webhooks"
done

rewrite_common "./vendor/knative.dev/serving/config/post-install/default-domain.yaml" "./config/core/200-imported/200-serving/deployments"

# We need the Image resource from caching, but used by serving.
rewrite_common "./vendor/knative.dev/caching/config/image.yaml" "./config/core/200-imported/200-serving/100-resources"

# Copy the autoscaler as-is.
rewrite_common "./vendor/knative.dev/serving/config/core/deployments/autoscaler.yaml" "./config/core/200-imported/200-serving/deployments"


#################################################
#
#
#    Eventing
#
#
#################################################

for x in $(list_yamls ./vendor/knative.dev/eventing/config/core/resources); do
  rewrite_common "$x" "./config/core/200-imported/200-eventing/100-resources"
done

rewrite_common "./vendor/knative.dev/eventing/config/core/roles/source-observer-clusterrole.yaml" ./config/core/200-imported/200-eventing/roles


#################################################
#
#
#    Contour and net-contour
#
#
#################################################

# Contour CRDs
rewrite_common "./vendor/github.com/projectcontour/contour/examples/contour/01-crds.yaml" "./config/core/200-imported/100-contour"

# Contour cert-gen Job
rewrite_common "./vendor/github.com/projectcontour/contour/examples/contour/02-job-certgen.yaml" "./config/core/200-imported/100-contour"
replace_text "./config/core/200-imported/100-contour/02-job-certgen.yaml" 65534 65532

#################################################
#
#
#    Tekton
#
#
#################################################

# Do a blanket copy of the resources
for x in $(list_yamls ./vendor/github.com/tektoncd/pipeline/config/ | grep 300-); do
  rewrite_common "$x" "./config/core/200-imported/200-tekton/100-resources"
done


#################################################
#
#
#    In-Memory Channel
#
#
#################################################

# Do a blanket copy of the resources
for x in $(list_yamls ./vendor/knative.dev/eventing/config/channels/in-memory-channel/); do
  rewrite_common "$x" "./config/in-memory/"
done


# Do this for every package under "cmd" except kodata and cmd itself.
# update_licenses third_party/VENDOR-LICENSE "$(find ./cmd -type d | grep -v kodata | grep -vE 'cmd$')"
