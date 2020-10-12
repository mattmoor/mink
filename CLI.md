# Installing the `mink` CLI

**We have not yet released the `mink` CLI binary, or the `kn` plugin.**

You can set these up from HEAD with:

```shell
./hack/build.sh --install
```

## Try it!

The `mink` CLI is available in two forms:

1. a `kn` plugin called `im` (`kn im` is `mink` backwards!), and
2. a standalone `mink` binary.

`mink` is a superset of `kn im`, so in the examples below we will use `kn im`
where either is acceptable, and reserve `mink` for contexts where that name must
be used.

### Authenticating

The `mink` CLI makes extensive use of the container registry as a ubiquitous and
standard object store. However, the typical model for authenticating with a
container registry is via `docker login`, and `mink` does not require users to
install `docker` locally. To facilitate logging in without `docker` we expose:

```shell
mink auth login my.registry.io -u username --password-stdin
```

### Bundle

To support building local source, `mink` bundles things into a self-extracting
container image, which when run expands the bundle into the working directory it
is run against.

To **just** produce a bundle, tell `mink` where to put it:

```shell
kn im bundle --image=gcr.io/mattmoor-knative/bundle
gcr.io/mattmoor-knative/bundle@sha256:41c60d8d8a7f5d38e8e63ce04913aded3d0efffbdafa23c835809114eb673f7e
```

### Build

To perform a `Dockerfile` build, `mink` provides the following command:

```shell
kn im build --as=me --image=gcr.io/mattmoor-knative/helloworld
```

This bundles the local build context and executes a kaniko build on Tekton
steaming the build output back via stderr and emitting the resulting image
digest to stdout. This enables us to easily composed commands:

```shell
kn service create helloworld --image=$(kn im build --as=me --image=gcr.io/mattmoor-knative/helloworld)
```

**NOTE:** The `--as=` command specifies the service account as which the build
is run, but `--as=me` is a special value that temporarily uploads YOUR local
docker credentials to the cluster. Please use this carefully in shared
environments.

Try it out on one of
[our samples](https://github.com/knative/docs/tree/master/docs/serving/samples/hello-world).

### Buildpack

To perform a [cloud native buildpacks](https://buildpacks.io) build, `mink`
provides the following command:

```shell
kn im buildpack --as=me --image=gcr.io/mattmoor-knative/helloworld
```

By default, this runs the [GCP](https://github.com/GoogleCloudPlatform/buildpacks#google-cloud-buildpacks) builder, but this can be
customized via `--builder`:

```shell
# Run the Paketo buildpacks
kn im buildpack --as=me --builder=gcr.io/paketo-buildpacks/builder:base --image=gcr.io/mattmoor-knative/hello-buildpack
```

As with [build](#build) this streams the output and enables composition with
`kn service` commands:

```shell
kn service create hello-buildpack --image=$(kn im buildpack --as=me --image=gcr.io/mattmoor-knative/hello-buildpack)
```

Try this out with some of the community samples:

- [Paketo Samples](https://github.com/paketo-buildpacks/samples)
- [GCP Samples](https://github.com/GoogleCloudPlatform/buildpack-samples)
