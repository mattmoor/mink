# `mink` development instructions

`mink` follows the development practices of Knative upstream as closesly as
practical. You will need `git`, `go`, `ko`, and a kubernetes environment.

## Pre-requisites

All of the below commands (including the CLI!) require the use of `ko`.

Please make sure `ko` is installed, on your path (`which ko`) and that you have
`KO_DOCKER_REPO` pointed at a registry with which you have authenticated.

The simplest way to test things are working is with:

```shell
ko publish ./cmd/kontext-expander
```

## `mink` on-cluster

To deploy all of `mink` simply run:

```shell
ko apply -BRf config/
```

This is the equivalent of the following parts:

```shell
# The "core" of mink
ko apply -BRf config/core

# The in-memory channel
ko apply -BRf config/in-memory
```

## `mink` CLI

To build the mink CLI run:

```shell
./hack/build.sh
```

You can also install this to `$GOPATH/bin` and `~/.config/kn/plugins` with:

```shell
./hack/build.sh
```
