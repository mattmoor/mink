# Development

This doc explains how to setup a development environment so you can get started
contributing to the VMware Sources for Knative. As much as possible we aim to
mirror the standard Knative development workflow.

## Getting started

1. Create [a GitHub account](https://github.com/join)
1. Setup
   [GitHub access via SSH](https://help.github.com/articles/connecting-to-github-with-ssh/)
1. Install [requirements](#requirements)
1. Set up your [shell environment](#environment-setup)
1. [Create and checkout a repo fork](#checkout-your-fork)

### Requirements

You must install these tools:

1. [`go`](https://golang.org/doc/install): The language Knative is built in
1. [`git`](https://help.github.com/articles/set-up-git/): For source control.
1. [`dep`](https://github.com/golang/dep): For managing external dependencies.
1. [`ko`](https://github.com/google/ko): The primary Knative development tool.

### Environment setup

To get started you'll need to set these environment variables (we recommend
adding them to your `.bashrc`):

1. `GOPATH`: If you don't have one, simply pick a directory and add
   `export GOPATH=...`

1. `$GOPATH/bin` on `PATH`: This is so that tooling installed via `go get` will
   work properly.

1. `KO_DOCKER_REPO`: This should be set to an authenticated Docker repo where
   you can publish container images during development, e.g. `docker.io/${USER}`

`.bashrc` example:

```shell
export GOPATH="$HOME/go"
export PATH="${PATH}:${GOPATH}/bin"
export KO_DOCKER_REPO="docker.io/${USER}"
```

### Checkout your fork

The Go tools require that you clone the repository to the
`src/github.con/vmware-tanzu/sources-for-knative` directory in your
[`GOPATH`](https://github.com/golang/go/wiki/SettingGOPATH).

To check out this repository:

1. Create your own
   [fork of this repo](https://help.github.com/articles/fork-a-repo/)

1. Clone it to your machine:

```shell
mkdir -p ${GOPATH}/src/github.com/vmware-tanzu
cd ${GOPATH}/src/github.com/vmware-tanzu
git clone git@github.com:${USER}/sources-for-knative.git
cd sources-for-knative
git remote add upstream https://github.com/vmware-tanzu/sources-for-knative.git
git remote set-url --push upstream no_push
```

_Adding the `upstream` remote sets you up nicely for regularly
[syncing your fork](https://help.github.com/articles/syncing-a-fork/)._

Once you reach this point you are ready to do a full build and deploy as
described below.

### Deploy `sources-for-knative` to a Kubernetes cluster

To deploy to the active `kubectl` context, run the following:

```shell
ko apply -f config
```

This will build all of the Go binaries into containers, publish them to your
`KO_DOCKER_REPO` and deploy them to the active `kubectl` context.

### Running the adapter on your local machine

Sometimes you might want to develop against a VSphere server that is not
accessible from your development cluster. So you can run the receive adapter
(the data plane for the events) locally like so. Note that you will still need a
kubernetes cluster to use for the ConfigMap-based bookmarking
([issue](https://github.com/vmware-tanzu-private/sources-for-knative/issues/16)).

Store the credentials on the filesystem:

```shell
mkdir -p /var/bindings/vsphere
echo -n 'administrator@Vsphere.local' > /var/bindings/vsphere/username
echo -n 'mysuper$ecretPassword' > /var/bindings/vsphere/password
```

Point at a configmap to use on your active `kubectl` context for bookmarking:

```shell
export NAMESPACE=default
export VSPHERE_KVSTORE_CONFIGMAP=vsphere-test
```

Then set up the necessary env variables:

```shell
export K_METRICS_CONFIG={}
export K_LOGGING_CONFIG={}
export GOVC_URL=<your vsphere url>
export GOVC_INSECURE=true
```

Then specify where the source should send events to

```shell
export K_SINK=http://localhost:8080
```

If you are using GKE for your bookmarking configmap, uncomment the following
line in `cmd/adapter/main.go`:

```go
	// Uncomment if you want to run locally against remote GKE cluster.
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
```

And then finally run the receive adapter, pointing to your kubeconfig file

```shell
go run ./cmd/adapter/main.go
```

### Local development notes with KinD

This section describes how to develop with KinD as your Kubernetes cluster, you
will have to have Docker running.

First install KinD, and create a cluster:

```shell
GO111MODULE="on" go get sigs.k8s.io/kind@v0.7.0 && kind create cluster
```

Make sure the KinD cluster is your active `kubectl` context:

```shell
kubectl config use-context kind-kind
```

Then install knative/eventing on it following
[the standard instructions](https://knative.dev/docs/install/any-kubernetes-cluster/#installing-the-eventing-component).

**If you are using a private registry for development** you will need to grant
the ServiceAccount access to your private repository. For GKE you would do it
like so:

```shell
SA_EMAIL=$(gcloud iam service-accounts --format='value(email)' create k8s-gcr-auth-ro)
gcloud iam service-accounts keys create k8s-gcr-auth-ro.json --iam-account=$SA_EMAIL

PROJECT=$(gcloud config list core/project --format='value(core.project)')
gcloud projects add-iam-policy-binding $PROJECT --member serviceAccount:$SA_EMAIL --role roles/storage.objectViewer

kubectl --context kind-kind -n vmware-sources create secret docker-registry image-secrets   --docker-server=https://gcr.io   --docker-username=_json_key   --docker-email=user@example.com   --docker-password="$(cat k8s-gcr-auth-ro.json)"
kubectl --context kind-kind -n vmware-sources patch serviceaccount controller -p "{\"imagePullSecrets\": [{\"name\": \"image-secrets\"}]}"
```

You can then iterate using the standard workflow:

```shell
ko apply -f ./config
```
