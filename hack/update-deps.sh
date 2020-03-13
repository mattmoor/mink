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

# The list of dependencies that we track at HEAD and periodically
# float forward in this repository.
FLOATING_DEPS=(
  "knative.dev/pkg"
  "knative.dev/serving"
  "knative.dev/eventing"
  "github.com/mattmoor/http01-solver"
  "github.com/tektoncd/pipeline"
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
rm -rf $(find vendor/knative.dev/ -type l)
rm -rf $(find vendor/github.com/tektoncd/ -type l)

# HACK HACK HACK
# The only way we found to create a consistent Trace tree without any missing Spans is to
# artificially set the SpanId. See pkg/tracing/traceparent.go for more details.
# See: https://github.com/knative/eventing/issues/2052
git apply ${REPO_ROOT_DIR}/vendor/knative.dev/eventing/hack/set-span-id.patch

function rewrite_knative_namespace() {
  sed 's@knative-serving@mink-system@g'
}

function rewrite_tekton_namespace() {
  sed 's@namespace: tekton-pipelines@namespace: mink-system@g'
}

function rewrite_contour_namespace() {
  sed 's@namespace: projectcontour@namespace: mink-system@g' | \
    sed 's@--namespace=projectcontour@--namespace=mink-system@g'
}

function rewrite_contour_image() {
  sed -E $'s@docker.io/projectcontour/contour:.+@github.com/mattmoor/mink/vendor/github.com/projectcontour/contour/cmd/contour@g'
}

function rewrite_annotation() {
  sed -E 's@(serving|eventing).knative.dev/release@knative.dev/release@g'
}

function rewrite_importpaths() {
  # TODO(mattmoor): Adopting ko:// would be helpful here.
  sed 's@knative.dev/serving/cmd@github.com/mattmoor/mink/vendor/knative.dev/serving/cmd@g' |\
    sed 's@knative.dev/net-contour/vendor@github.com/mattmoor/mink/vendor@g'
}

function rewrite_ingress_class() {
  sed -e $'s@    ingress.class: "istio.ingress.networking.knative.dev"@  ingress.class: "contour.ingress.networking.knative.dev"\\\n  _other: |@g'
}

function rewrite_certificate_class() {
  sed -e $'s@    certificate.class: "cert-manager.certificate.networking.knative.dev"@  certificate.class: "mattmoor-http01.certificate.networking.knative.dev"\\\n  _other2: |@g'
}

function enable_auto_tls() {
  sed -e $'s@    autoTLS: "Disabled"@  autoTLS: "Enabled"\\\n  _other3: |@g'
}

function rewrite_webhook() {
  sed 's@webhook.serving.knative.dev@webhook.mink.knative.dev@g'
}

function rewrite_deploy_to_daemon() {
  sed 's@kind: Deployment@kind: DaemonSet@g'
}

function rewrite_common() {
  local readonly INPUT="${1}"
  local readonly OUTPUT_DIR="${2}"

  cat "${INPUT}" | rewrite_knative_namespace | rewrite_tekton_namespace | rewrite_contour_namespace | rewrite_annotation | rewrite_webhook \
    | rewrite_importpaths | rewrite_ingress_class | rewrite_certificate_class | rewrite_contour_image | enable_auto_tls > "${OUTPUT_DIR}/$(basename ${INPUT})"
}

function list_yamls() {
  find "$1" -type f -name '*.yaml'
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
for dir in configmaps webhooks ; do
  for x in $(list_yamls ./vendor/knative.dev/serving/config/core/$dir | grep -v config-defaults); do
    rewrite_common "$x" "./config/core/200-imported/200-serving/$dir"
  done
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
# TODO(mattmoor): We'll need this once we pull in the broker stuff.
# rewrite_common "./vendor/knative.dev/eventing/config/core/configmaps/default-channel.yaml" "./config/core/200-imported/200-eventing/configmaps"


#################################################
#
#
#    Contour and net-contour
#
#
#################################################

# This is designed to live alongside of the serving stuff.
rewrite_common "./vendor/knative.dev/net-contour/config/200-clusterrole.yaml" "./config/core/200-imported/net-contour/rbac"

# Contour CRDs
rewrite_common "./vendor/github.com/projectcontour/contour/examples/contour/01-crds.yaml" "./config/core/200-imported/100-contour"

# Contour cert-gen Job
rewrite_common "./vendor/github.com/projectcontour/contour/examples/contour/02-job-certgen.yaml" "./config/core/200-imported/100-contour"


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

# ConfigMaps
rewrite_common "./vendor/github.com/tektoncd/pipeline/config/config-artifact-bucket.yaml" "./config/core/200-imported/200-tekton/configmaps"
rewrite_common "./vendor/github.com/tektoncd/pipeline/config/config-artifact-pvc.yaml" "./config/core/200-imported/200-tekton/configmaps"
