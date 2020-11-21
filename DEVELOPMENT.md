# `mink` development instructions

`mink` follows the development practices of Knative upstream as closely as
practical. You will need `git`, `go`, `ko`, and a kubernetes environment.

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

You can also install this to `$GOPATH/bin` and `~/.kn/plugins` with:

```shell
./hack/build.sh --install
```
