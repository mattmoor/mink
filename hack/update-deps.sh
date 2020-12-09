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

source $(dirname "$0")/../vendor/knative.dev/hack/library.sh

CONTOUR_VERSION="v1.10.0"
export FLOATING_DEPS=(
  "github.com/projectcontour/contour@${CONTOUR_VERSION}"

  "github.com/tektoncd/pipeline@master"
  "github.com/tektoncd/cli@master"
)

go_update_deps "$@"

rm -rf $(find vendor/knative.dev/ -type l)
rm -rf $(find vendor/github.com/tektoncd/ -type l)

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

  cat "${INPUT}" | rewrite_knative_namespace | rewrite_tekton_namespace | rewrite_contour_namespace \
    | rewrite_annotation | rewrite_webhook | rewrite_nobody | sed -e's/[[:space:]]*$//' \
    | rewrite_contour_image > "${OUTPUT_DIR}/$(basename ${INPUT})"
}

function list_yamls() {
  find "$1" -type f -name '*.yaml' -mindepth 1 -maxdepth 1
}

function rewrite_nobody() {
  sed -e $'s@65534@65532@g'
}

# Remove all of the imported yamls before we start to do our rewrites.
rm $(find config/ -type f | grep imported)
rm $(find config/in-memory -type f)

#################################################
#
#
#    Serving
#
#
#################################################

# Do a blanket copy of these resources
for x in $(list_yamls ./vendor/knative.dev/serving/config/core/300-resources); do
  rewrite_common "$x" "./config/core/200-imported/200-serving/100-resources"
done
for x in $(list_yamls ./vendor/knative.dev/serving/config/domain-mapping/300-resources); do
  rewrite_common "$x" "./config/core/200-imported/200-serving/100-resources"
done
for x in $(list_yamls ./vendor/knative.dev/serving/config/core/webhooks); do
  rewrite_common "$x" "./config/core/200-imported/200-serving/webhooks"
done

rewrite_common "./vendor/knative.dev/serving/config/post-install/default-domain.yaml" "./config/core/200-imported/200-serving/deployments"

# We need the Image resource from caching, but used by serving.
rewrite_common "./vendor/knative.dev/caching/config/image.yaml" "./config/core/200-imported/200-serving/100-resources"

# We need the resources from networking, but used by serving.
rewrite_common "./vendor/knative.dev/networking/config/certificate.yaml" "./config/core/200-imported/200-serving/100-resources"
rewrite_common "./vendor/knative.dev/networking/config/ingress.yaml" "./config/core/200-imported/200-serving/100-resources"
rewrite_common "./vendor/knative.dev/networking/config/serverlessservice.yaml" "./config/core/200-imported/200-serving/100-resources"
rewrite_common "./vendor/knative.dev/networking/config/domain-claim.yaml" "./config/core/200-imported/200-serving/100-resources"


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
rewrite_common "./vendor/knative.dev/eventing/config/core/roles/channelable-manipulator-clusterrole.yaml" ./config/core/200-imported/200-eventing/roles

rewrite_common "./vendor/knative.dev/eventing/config/core/deployments/pingsource-mt-adapter.yaml" "./config/core/200-imported/200-eventing/deployments"


#################################################
#
#
#    Contour and net-contour
#
#
#################################################

TMP_DIR=$(mktemp -d)

for f in 01-crds 02-job-certgen ; do
  wget -O ${TMP_DIR}/$f.yaml https://raw.githubusercontent.com/projectcontour/contour/${CONTOUR_VERSION}/examples/contour/$f.yaml

  rewrite_common "${TMP_DIR}/$f.yaml" "./config/core/200-imported/100-contour"
done

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
for dir in . resources deployments configmaps roles; do
  for x in $(list_yamls ./vendor/knative.dev/eventing/config/channels/in-memory-channel/$dir); do
    rewrite_common "$x" "./config/in-memory/$dir"
  done
done

# Make sure that all binaries have the appropriate kodata with our version and license data.
for binary in $(find ./config/ -type f | xargs grep ko:// | sed 's@.*ko://@@g' | sed 's@",$@@g' | sort | uniq); do
  if [[ ! -d ./vendor/$binary ]]; then
    echo Skipping $binary, not in vendor.
    continue
  fi
  mkdir ./vendor/$binary/kodata
  pushd ./vendor/$binary/kodata > /dev/null
  ln -s $(echo vendor/$binary/kodata | sed -E 's@[^/]+@..@g')/.git/HEAD .
  ln -s $(echo vendor/$binary/kodata | sed -E 's@[^/]+@..@g')/.git/refs .
  ln -s $(echo vendor/$binary/kodata | sed -E 's@[^/]+@..@g')/LICENSE .
  ln -s $(echo vendor/$binary/kodata | sed -E 's@[^/]+@..@g')/third_party/VENDOR-LICENSE .
  popd > /dev/null
done
