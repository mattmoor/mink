# Development

This doc explains how to setup a development environment so you can get started
[contributing](https://github.com/knative/docs/blob/master/community/CONTRIBUTING.md)
to Knative `net-contour`. Also take a look at:

- [The pull request workflow](https://github.com/knative/docs/blob/master/community/CONTRIBUTING.md#pull-requests)

## Getting started

1. Create [a GitHub account](https://github.com/join)
1. Setup
   [GitHub access via SSH](https://help.github.com/articles/connecting-to-github-with-ssh/)
1. Install [requirements](#requirements)
1. Set up your [shell environment](#environment-setup)
1. [Create and checkout a repo fork](#checkout-your-fork)

Before submitting a PR, see also [CONTRIBUTING.md](./CONTRIBUTING.md).

### Requirements

You must install these tools:

1. [`go`](https://golang.org/doc/install): The language Knative `net-contour` is
   built in
1. [`git`](https://help.github.com/articles/set-up-git/): For source control
1. [`dep`](https://github.com/golang/dep): For managing external dependencies.
1. [`ko`](https://github.com/google/ko): For development.
1. [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl/): For
   managing development environments.

### Environment setup

To get started you'll need to set these environment variables (we recommend
adding them to your `.bashrc`):

1. `GOPATH`: If you don't have one, simply pick a directory and add
   `export GOPATH=...`
1. `$GOPATH/bin` on `PATH`: This is so that tooling installed via `go get` will
   work properly.

`.bashrc` example:

```shell
export GOPATH="$HOME/go"
export PATH="${PATH}:${GOPATH}/bin"
```

### Checkout your fork

The Go tools require that you clone the repository to the
`src/knative.dev/net-contour` directory in your
[`GOPATH`](https://github.com/golang/go/wiki/SettingGOPATH).

To check out this repository:

1. Create your own
   [fork of this repo](https://help.github.com/articles/fork-a-repo/)
1. Clone it to your machine:

```shell
mkdir -p ${GOPATH}/src/knative.dev
cd ${GOPATH}/src/knative.dev
git clone git@github.com:${YOUR_GITHUB_USERNAME}/net-contour.git
cd net-contour
git remote add upstream https://knative.dev/net-contour.git
git remote set-url --push upstream no_push
```

_Adding the `upstream` remote sets you up nicely for regularly
[syncing your fork](https://help.github.com/articles/syncing-a-fork/)._

Once you reach this point you are ready to do a full build and deploy as
described below.

### Installing Contour

Before deploying the `net-contour` controller you will need a properly
configured installation of Contour. We provide a version of this that can be
built from source via:

```bash
ko apply -f config/contour
```

### Installing and Iterating on `net-contour`

Once you have a knative/serving installation, and an appropriately configured
Contour installation, you can install the `net-contour` controller via:

```bash
ko apply -f config/
```

### Configuring Knative Serving to use Contour by default

You can configure Serving to use Contour by default with the following command:

```bash
kubectl patch configmap/config-network \
  --namespace knative-serving \
  --type merge \
  --patch '{"data":{"ingress.class":"contour.ingress.networking.knative.dev"}}'
```
