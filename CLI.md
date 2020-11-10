# Installing the `mink` CLI

To install the `mink` CLI, download the latest release:

```shell
# Make sure you pick the latest version, and the appropriate platform / architecture.
VERSION=0.19.2

wget https://github.com/mattmoor/mink/releases/download/v${VERSION}/mink_${VERSION}_Linux_x86_64.tar.gz
tar xzvf mink_${VERSION}_Linux_x86_64.tar.gz mink_${VERSION}_Linux_x86_64/mink
sudo mv mink_${VERSION}_Linux_x86_64/mink /usr/local/bin
```

You can then use the `mink` CLI to install `mink` onto your cluster via:

```shell
mink install
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

### Configuration

Mink will read and blend configuration from two files, in addition to
environment variables using [viper](https://github.com/spf13/viper):

Configuration files named `.mink.yaml` are discovered via:

1. the "nearest" file in the working directory or parent directories.
2. the user's home directory

A command line flag, e.g. `--foo` can be configured via either:

```yaml
foo: bar
```

or:

```shell
export MINK_FOO=bar
```

The configuration files are blended because different flags vary in different
ways. For example, some settings like the docker registry to publish source and
binary images may vary by developer, but the developer may use the same settings
across all of the projects they work on. For them, you might find `~/.mink.yaml`
with something like:

```yaml
# Where to upload source (if unspecified)
bundle: gcr.io/mattmoor-knative/mink-bundles

# Where to upload built images (if unspecified)
image: gcr.io/mattmoor-knative/mink-images

# Who to run the build as (if unspecified)
# **NOTE:** The `as` option specifies the service account as which the build
# is run, but `as: me` is a special value that temporarily uploads YOUR local
# docker credentials to the cluster.  I exclusively use sole-tenancy clusters.
as: me
```

However, other settings may vary depending on the project being worked on, and
apply to all developers on the project, such as the buildpack builder image they
use. For these projects you might find `.mink.yaml` in the project root with
something like:

```yaml
# This project uses the GCP buildpacks image.
builder: gcr.io/buildpacks/builder
```

These are simply illustrative examples, all of these settings are configurable
via these mechanisms and follow the same precedence:

1. Flags always win (`--foo`)
2. Environment variables (`MINK_FOO`)
3. Project configuration (`foo: `)
4. User configuration (`foo: `)

Note: User configuration is last here because users could always specify
environment variables to override things as well.

### Bundle

To support building local source, `mink` bundles things into a self-extracting
container image, which when run expands the bundle into the working directory it
is run against.

To **just** produce a bundle, tell `mink` where to put it:

```shell
kn im bundle
gcr.io/mattmoor-knative/bundle@sha256:41c60d8d8a7f5d38e8e63ce04913aded3d0efffbdafa23c835809114eb673f7e
```

### Build

To perform a `Dockerfile` build, `mink` provides the following command:

```shell
kn im build
```

This bundles the local build context and executes a kaniko build on Tekton
steaming the build output back via stderr and emitting the resulting image
digest to stdout. This enables us to easily composed commands:

```shell
kn service create helloworld --image=$(kn im build)
```

Try it out on one of
[our samples](https://github.com/knative/docs/tree/master/docs/serving/samples/hello-world).

### Buildpack

To perform a [cloud native buildpacks](https://buildpacks.io) build, `mink`
provides the following command:

```shell
kn im buildpack
```

By default, this runs the
[Paketo builder](https://github.com/paketo-buildpacks/builder), but this can be
customized via `--builder`:

```shell
# Run the GCP buildpacks
kn im buildpack --builder=gcr.io/buildpacks/builder

# Run the Boson Node.js buildpack
kn im buildpack --builder=quay.io/boson/faas-nodejs-builder
```

As with [build](#build) this streams the output and enables composition with
`kn service` commands:

```shell
kn service create hello-buildpack --image=$(kn im buildpack)
```

Try this out with some of the community samples:

- [Paketo Samples](https://github.com/paketo-buildpacks/samples)
- [GCP Samples](https://github.com/GoogleCloudPlatform/buildpack-samples)
- [Boson Templates](https://github.com/boson-project/faas/tree/main/templates)

### Apply and Resolve.

For more on `mink apply` and `mink resolve` see [here](./APPLY.md).
