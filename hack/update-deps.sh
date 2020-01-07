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

# Ensure we have everything we need under vendor/
dep ensure

rm -rf $(find vendor/ -name 'OWNERS')
rm -rf $(find vendor/ -name '*_test.go')
rm -rf $(find vendor/knative.dev/ -type l)

function rewrite_namespace() {
  sed 's@knative-serving@knative-system@g'
}

function rewrite_annotation() {
  sed 's@serving.knative.dev/release@knative.dev/release@g'
}

function rewrite_importpaths() {
  # TODO(mattmoor): Adopting ko:// would be helpful here.
  sed 's@knative.dev/serving/cmd@github.com/mattmoor/mink/vendor/knative.dev/serving/cmd@g'
}

function rewrite_ingress_class() {
  sed -e $'s@    ingress.class: "istio.ingress.networking.knative.dev"@  ingress.class: "contour.ingress.networking.knative.dev"\\\n  _other: |@g'
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

  cat "${INPUT}" | rewrite_namespace | rewrite_annotation | rewrite_webhook \
    | rewrite_importpaths | rewrite_ingress_class > "${OUTPUT_DIR}/$(basename ${INPUT})"
}

function rewrite_daemonset() {
  local readonly INPUT="${1}"
  local readonly OUTPUT_DIR="${2}"

  cat "${INPUT}" | rewrite_namespace | rewrite_annotation | rewrite_webhook \
    | rewrite_importpaths | rewrite_deploy_to_daemon | rewrite_ingress_class > "${OUTPUT_DIR}/$(basename ${INPUT})"
}

function list_yamls() {
  find "$1" -type f -name '*.yaml'
}

# Do a blanket copy of these resources
for dir in resources rbac configmaps webhooks ; do
  for x in $(list_yamls ./vendor/knative.dev/serving/config/core/$dir); do
    rewrite_common "$x" "./config/core/imported/serving/$dir"
  done
done

rewrite_common "./vendor/knative.dev/serving/config/post-install/default-domain.yaml" "./config/core/imported/serving/deployments"

# We need the Image resource from caching, but used by serving.
rewrite_common "./vendor/knative.dev/caching/config/image.yaml" "./config/core/imported/serving/resources"

# Rewrite the activator to a DaemonSet.
# TODO(mattmoor): perhaps stop auto-rewriting it to do this and combine with Contour?
rewrite_daemonset "./vendor/knative.dev/serving/config/core/deployments/activator.yaml" "./config/core/imported/serving/deployments"

# Copy the autoscaler as-is.
rewrite_common "./vendor/knative.dev/serving/config/core/deployments/autoscaler.yaml" "./config/core/imported/serving/deployments"
