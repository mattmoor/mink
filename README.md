# `mink`: a minimal Knative distribution.

`mink` is a distribution of [Knative](https://knative.dev/) and
[Tekton](https://tekton.dev) components.  The goal of `mink` is
to form a complete foundation for modern application development,
which is simple to install and get started.

[![Releases](https://img.shields.io/github/release-pre/mattmoor/mink.svg?sort=semver)](https://github.com/mattmoor/mink/releases)

You can install `mink` from our [releases](https://github.com/mattmoor/mink/releases) with:

```shell
# Make sure you pick the latest version!
kubectl apply -f https://github.com/mattmoor/mink/releases/download/v0.14.0/release.yaml
```

For basic development that's it!  Watch for the components in `mink-system` to
become ready and then try out one of the Knative or Tekton samples.

* To set up DNS and TLS continue [here](./DNS.md).
* To set up the `mink` CLI continue [here](./CLI.md).