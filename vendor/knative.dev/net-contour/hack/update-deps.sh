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

CONTOUR_VERSION="v1.9.0" # This is for controlling which version of contour we want to use.

FLOATING_DEPS=(
  "github.com/projectcontour/contour@${CONTOUR_VERSION}"
)

go_update_deps "$@"

# Remove unit tests & e2e tests.
rm -rf $(find vendor/ -path '*/pkg/*_test.go')
rm -rf $(find vendor/ -path '*/e2e/*_test.go')

# Add permission for shell scripts
chmod +x $(find vendor -type f -name '*.sh')

function add_ingress_provider_labels() {
  sed '${/---/d;}' | go run ${REPO_ROOT_DIR}/vendor/github.com/mikefarah/yq/v3 m - ./hack/labels.yaml -d "*"
}

function delete_contour_cluster_role_bindings() {
  sed -e '/apiVersion: rbac.authorization.k8s.io/{' -e ':a' -e '${' -e 'p' -e 'd'  -e '}' -e 'N' -e '/---/!ba' -e '/kind: ClusterRoleBinding/d' -e '}'
}

function rewrite_contour_namespace() {
  sed "s@namespace: projectcontour@namespace: $1@g" \
      | sed "s@name: projectcontour@name: $1@g"
}

function rewrite_serve_args() {
  sed -e $'s@        - serve@        - serve\\\n        - --ingress-class-name='$1'@g'
}

function rewrite_image() {
  sed -E $'s@docker.io/projectcontour/contour:.+@ko://github.com/projectcontour/contour/cmd/contour@g'
}

function rewrite_command() {
  sed -e $'s@/bin/contour@contour@g'
}

function disable_hostport() {
  sed -e $'s@hostPort:@# hostPort:@g'
}

function rewrite_user() {
  sed -e $'s@65534@65532@g'
}

function privatize_loadbalancer() {
  sed "s@type: LoadBalancer@type: ClusterIP@g" \
    | sed "s@externalTrafficPolicy: Local@# externalTrafficPolicy: Local@g"
}

function contour_yaml() {
  # Used to be: KO_DOCKER_REPO=ko.local ko resolve -f ./vendor/github.com/projectcontour/contour/examples/contour/
  curl "https://raw.githubusercontent.com/projectcontour/contour/${CONTOUR_VERSION}/examples/render/contour.yaml"
}

rm -rf config/contour/*

# We do this manually because it's challenging to rewrite
# the ClusterRoleBinding without collateral damage.
cat > config/contour/internal.yaml <<EOF
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: contour-internal
  labels:
    networking.knative.dev/ingress-provider: contour
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

contour_yaml \
  | delete_contour_cluster_role_bindings \
  | rewrite_contour_namespace contour-internal \
  | rewrite_serve_args contour-internal | rewrite_user \
  | rewrite_image | rewrite_command | disable_hostport | privatize_loadbalancer \
  | add_ingress_provider_labels  >> config/contour/internal.yaml

# We do this manually because it's challenging to rewrite
# the ClusterRoleBinding without collateral damage.
cat > config/contour/external.yaml <<EOF
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: contour-external
  labels:
    networking.knative.dev/ingress-provider: contour
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

contour_yaml \
  | delete_contour_cluster_role_bindings \
  | rewrite_contour_namespace contour-external \
  | rewrite_serve_args contour-external | rewrite_user \
  | rewrite_image | rewrite_command | disable_hostport \
  | add_ingress_provider_labels >> config/contour/external.yaml
