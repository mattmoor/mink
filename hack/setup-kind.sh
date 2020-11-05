#!/usr/bin/env bash

# Copyright 2020 The Knative Authors
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

source "$(dirname "$0")/../vendor/knative.dev/hack/library.sh"

if [[ -z "${GITHUB_WORKSPACE}" ]]; then
  echo "This script is expected to run in the context of GitHub Actions."
  exit 1
fi

# Defaults
K8S_VERSION="1.17.x"
REGISTRY_NAME="registry.local"
REGISTRY_PORT="5000"
CLUSTER_SUFFIX="cluster.local"
NODE_COUNT="1"

while [[ $# -ne 0 ]]; do
  parameter="$1"
  case "${parameter}" in
    --k8s-version)
      shift
      K8S_VERSION="$1"
      ;;
    --registry-url)
      shift
      REGISTRY_NAME="$(echo "$1" | cut -d':' -f 1)"
      REGISTRY_PORT="$(echo "$1" | cut -d':' -f 2)"
      ;;
    --cluster-suffix)
      shift
      CLUSTER_SUFFIX="$1"
      ;;
    --nodes)
      shift
      NODE_COUNT="$1"
      ;;
    *) abort "unknown option ${parameter}" ;;
  esac
  shift
done

# The version map correlated with this version of KinD
KIND_VERSION="v0.9.0"
case ${K8S_VERSION} in
  v1.17.x)
    K8S_VERSION="1.17.11"
    KIND_IMAGE_SHA="sha256:5240a7a2c34bf241afb54ac05669f8a46661912eab05705d660971eeb12f6555"
    ;;
  v1.18.x)
    K8S_VERSION="1.18.8"
    KIND_IMAGE_SHA="sha256:f4bcc97a0ad6e7abaf3f643d890add7efe6ee4ab90baeb374b4f41a4c95567eb"
    ;;
  v1.19.x)
    K8S_VERSION="1.19.1"
    KIND_IMAGE_SHA="sha256:98cf5288864662e37115e362b23e4369c8c4a408f99cbc06e58ac30ddc721600"
    ;;
  *) abort "Unsupported version: ${K8S_VERSION}" ;;
esac

#############################################################
#
#    Install KinD
#
#############################################################
echo '::group:: Install KinD'

# Disable swap otherwise memory enforcement does not work
# See: https://kubernetes.slack.com/archives/CEKK1KTN2/p1600009955324200
sudo swapoff -a
sudo rm -f /swapfile

curl -Lo ./kind "https://github.com/kubernetes-sigs/kind/releases/download/${KIND_VERSION}/kind-$(uname)-amd64"
chmod +x ./kind
sudo mv kind /usr/local/bin

echo '::endgroup::'


#############################################################
#
#    Setup KinD cluster.
#
#############################################################
echo '::group:: Setup KinD Cluster'

cat > kind.yaml <<EOF
apiVersion: kind.x-k8s.io/v1alpha4
kind: Cluster
nodes:
- role: control-plane
  image: kindest/node:${K8S_VERSION}@${KIND_IMAGE_SHA}
EOF

for i in $(seq 1 1 "${NODE_COUNT}");
do
  cat >> kind.yaml <<EOF
- role: worker
  image: kindest/node:${K8S_VERSION}@${KIND_IMAGE_SHA}
EOF
done

cat >> kind.yaml <<EOF
kubeadmConfigPatches:
  # This is needed in order to support projected volumes with service account tokens.
  # See: https://kubernetes.slack.com/archives/CEKK1KTN2/p1600268272383600
  - |
    apiVersion: kubeadm.k8s.io/v1beta2
    kind: ClusterConfiguration
    metadata:
      name: config
    apiServer:
      extraArgs:
        "service-account-issuer": "kubernetes.default.svc"
        "service-account-signing-key-file": "/etc/kubernetes/pki/sa.key"
    networking:
      dnsDomain: "${CLUSTER_SUFFIX}"

  # This is needed to avoid filling our disk.
  # See: https://kubernetes.slack.com/archives/CEKK1KTN2/p1603391142276400
  - |
    kind: KubeletConfiguration
    metadata:
      name: config
    imageGCHighThresholdPercent: 90

# Support a local registry
# Support many layered images: https://kubernetes.slack.com/archives/CEKK1KTN2/p1602770111199000
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."$REGISTRY_NAME:$REGISTRY_PORT"]
    endpoint = ["http://$REGISTRY_NAME:$REGISTRY_PORT"]
  [plugins."io.containerd.grpc.v1.cri".containerd]
    disable_snapshot_annotations = true
EOF

# Create a cluster!
kind create cluster --config kind.yaml

echo '::endgroup::'


#############################################################
#
#    Setup metallb
#
#############################################################
echo '::group:: Setup metallb'

kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.9.3/manifests/namespace.yaml
kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.9.3/manifests/metallb.yaml
kubectl create secret generic -n metallb-system memberlist --from-literal=secretkey="$(openssl rand -base64 128)"

network=$(docker network inspect kind -f "{{(index .IPAM.Config 0).Subnet}}" | cut -d '.' -f1,2)
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  namespace: metallb-system
  name: config
data:
  config: |
    address-pools:
    - name: default
      protocol: layer2
      addresses:
      - $network.255.1-$network.255.250
EOF

echo '::endgroup::'


#############################################################
#
#    Setup container registry
#
#############################################################
echo '::group:: Setup container registry'

# Run a registry.
docker run -d --restart=always \
  -p "$REGISTRY_PORT:$REGISTRY_PORT" --name "$REGISTRY_NAME" registry:2
# Connect the registry to the KinD network.
docker network connect "kind" "$REGISTRY_NAME"
# Make the $REGISTRY_NAME -> 127.0.0.1, to tell `ko` to publish to
# local reigstry, even when pushing $REGISTRY_NAME:$REGISTRY_PORT/some/image
sudo echo "127.0.0.1 $REGISTRY_NAME" | sudo tee -a /etc/hosts

echo '::endgroup::'
