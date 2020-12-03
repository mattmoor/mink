# `mink`: a minimal Knative distribution (min-K)

`mink` is a distribution of [Knative](https://knative.dev/) and
[Tekton](https://tekton.dev) components. The goal of `mink` is to form a
complete foundation for modern application development, which is simple to
install and get started.

[![Releases](https://img.shields.io/github/release-pre/mattmoor/mink.svg?sort=semver)](https://github.com/mattmoor/mink/releases)

You can install `mink` directly from our
[releases](https://github.com/mattmoor/mink/releases) with:

```shell
kubectl apply -f https://github.com/mattmoor/mink/releases/download/v0.19.0/release.yaml
```

> NOTE: You can also install `mink` via the [CLI](./CLI.md).

For basic development that's it! Watch for the components in `mink-system` to
become ready and then try out one of the Knative or Tekton samples.

- Additional `mink` setup for clusters with workload-identity enabled, continue
  [here](./WORKLOAD-IDENTITY.md).
- To set up DNS and TLS continue [here](./DNS.md).
- To set up the `mink` CLI continue [here](./CLI.md).
