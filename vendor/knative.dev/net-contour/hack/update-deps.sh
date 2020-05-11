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
  "knative.dev/test-infra"
  "github.com/projectcontour/contour"
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
# Remove unit tests & e2e tests.
rm -rf $(find vendor/ -path '*/pkg/*_test.go')
rm -rf $(find vendor/ -path '*/e2e/*_test.go')

function delete_contour_cluster_role_bindings() {
  sed -e '/apiVersion: rbac.authorization.k8s.io/{' -e ':a' -e '${' -e 'p' -e 'd'  -e '}' -e 'N' -e '/---/!ba' -e '/kind: ClusterRoleBinding/d' -e '}'
}

function rewrite_contour_namespace() {
  sed "s@namespace: projectcontour@namespace: $1@g" \
      | sed "s@name: projectcontour@name: $1@g"
}

function configure_leader_election() {
  sed -e $'s@  contour.yaml: |@  contour.yaml: |\\\n    leaderelection:\\\n      configmap-name: contour\\\n      configmap-namespace: '$1'@g'
}

function rewrite_serve_args() {
  sed -e $'s@        - serve@        - serve\\\n        - --ingress-class-name='$1'@g'
}

function rewrite_image() {
  sed -E $'s@docker.io/projectcontour/contour:.+@ko://knative.dev/net-contour/vendor/github.com/projectcontour/contour/cmd/contour@g'
}

function rewrite_command() {
  sed -e $'s@/bin/contour@contour@g'
}

function disable_hostport() {
  sed -e $'s@hostPort:@# hostPort:@g'
}

function privatize_loadbalancer() {
  sed "s@type: LoadBalancer@type: ClusterIP@g" \
    | sed "s@externalTrafficPolicy: Local@# externalTrafficPolicy: Local@g"
}

rm -rf config/contour/*

# Apply patch to contour
git apply ${ROOT_DIR}/hack/contour.patch

# We do this manually because it's challenging to rewrite
# the ClusterRoleBinding without collateral damage.
cat > config/contour/internal.yaml <<EOF
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: contour-internal
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: contour
subjects:
- kind: ServiceAccount
  name: contour
  namespace: contour-internal
---
EOF

KO_DOCKER_REPO=ko.local ko resolve -f ./vendor/github.com/projectcontour/contour/examples/contour/ \
  | delete_contour_cluster_role_bindings \
  | rewrite_contour_namespace contour-internal \
  | configure_leader_election contour-internal \
  | rewrite_serve_args contour-internal \
  | rewrite_image | rewrite_command | disable_hostport | privatize_loadbalancer >> config/contour/internal.yaml

# We do this manually because it's challenging to rewrite
# the ClusterRoleBinding without collateral damage.
cat > config/contour/external.yaml <<EOF
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: contour-external
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: contour
subjects:
- kind: ServiceAccount
  name: contour
  namespace: contour-external
---
EOF

KO_DOCKER_REPO=ko.local ko resolve -f ./vendor/github.com/projectcontour/contour/examples/contour/ \
  | delete_contour_cluster_role_bindings \
  | rewrite_contour_namespace contour-external \
  | configure_leader_election contour-external \
  | rewrite_serve_args contour-external \
  | rewrite_image | rewrite_command | disable_hostport >> config/contour/external.yaml
