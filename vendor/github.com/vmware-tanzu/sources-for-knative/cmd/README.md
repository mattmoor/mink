## Binary names

The binaries in this directory are prefixed with `sources-for-knative-{foo}` so
that when published via `KO_DOCKER_REPO=docker.io/vmware ko apply -Bf config`
the resulting images are named `vmware/sources-for-knative-{foo}`.

