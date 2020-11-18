# Using `mink apply` and `mink resolve`

Starting in the 0.19 release of `mink`, we will have a new feature based around
the user experience of [`ko`](https://github.com/google/ko).

Suppose my repository is laid out like:

```
  foo/
    Dockerfile
    main.js
  bar/
    main.go
  baz/
    main.ru
    overrides.toml
  config/
    lots-of.yaml
```

I would like to build, containerize and deploy all of this with a simple:
`mink apply`.

### Authoring configs for `mink apply`

In the `ko`-style of development, instead of referencing images, config files
reference Go-importpaths:

```
  image: ko://github.com/mattmoor/mink/foo/bar
```

`mink apply` adopts a similar strategy, where image URIs are prefixed with a
scheme that correlates with the build mode, following the above structure you
would use something like:

```
  image: dockerfile:///foo
  image: ko://bar
  image: buildpacks:///baz
```

> Note: currently dockerfile/buildpacks requires triple-slashes for
> `file:///`-style URIs

### How this works

`mink` will upload a single "bundle" of the entire repository (guided by
`--directory` for the "root", for more see
[Complex directory structures](#complex-directory-structures)). It will then
pass this bundle to all of the different builds (see detailed sections).

As each build completes, the resulting image digest is substituted into the
source yaml. For `apply` this is piped to `kubectl apply`, and for `resolve`
this is printed to stdout (for more, see
[What about releases?](#what-about-releases)).

#### `ko://` semantics

The semantics of `ko://a/b/c` are equivalent to that of `github.com/google/ko`.

This build may be reproduced with:

```shell
ko publish --bare a/b/c
```

#### `dockerfile:///` semantics

`dockerfile:///a/b/c` will trigger a Dockerfile build within the uploaded
context with `a/b/c/Dockerfile`. If `--dockerfile=Dockerfile.blah` is passed
then the build will use `a/b/c/Dockerfile.blah`. There is not currently a way to
scope the build context differently or supply different Dockerfile names per
build target.

This build may be reproduced with:

```shell
mink build --dockerfile=a/b/c/Dockerfile
```

#### `buildpack:///` semantics

`buildpack:///a/b/c` will trigger a buildpack build within the uploaded context
with a set of optional overrides to `project.toml` supplied via
`a/b/c/overrides.toml`. If `--overrides=blah.toml` is passed then the build will
use `a/b/c/blah.toml` for the overrides. There is not currently a way to scope
the build context differently or supply different `--overrides` per build
target.

This build may be reproduced with:

```shell
mink buildpack --overrides=a/b/c/blah.toml
```

> **NOTE:** `overrides.toml` is a `mink`-specific concept that builds around the
> buildpack construct of `project.toml`, **it is not portable**.

Example using `overrides.toml` to select the Go package to build:

```toml
[[build.env]]
name = "BP_GO_TARGETS"
value = "./cmd/foo"
```

Example using `overrides.toml` to specify the GCP function to target:

```toml
[[build.env]]
name = "GOOGLE_FUNCTION_TARGET"
value = "bar"
```

### What about releases?

Similar to `ko`, this can be used to produce releases as well via
`mink resolve`:

```
mink resolve -f config > release.yaml
```

This will produce a version of the input yaml with image references substituted
in their digest form.

### Advanced configuration

Unlike `ko`, `mink`'s style of configuration allows projects to drop the
`-f path/to/config`! In the root of your repo, simply add the filenames to your
`.mink.yaml` configuration:

```yaml
filename:
  - ./path/to/my/config
  - ./another/path/to/some/config
```

Then folks can just type `mink apply`.

> Note: Currently there is no way to pass configuration options for individual
> builds (e.g. different buildpack builder per build)

### Complex directory structures

Suppose we have complex directory structure:

```
  deep/
    deeper/
      deepest/
        foo/
          Dockerfile
          main.js
        bar/
          main.go
      baz/
        main.ru
	overrides.toml
    config/
      lots-of.yaml
```

If you supply the command: `mink apply -f deep/config --directory=deep/deeper`,
then the bundle uploaded will contain:

```
  deepest/
    foo/
      Dockerfile
      main.js
    bar/
      main.go
  baz/
    main.ru
    overrides.toml
```

So the config references should take this into account and use:

```
  image: dockerfile:///deepest/foo
  image: ko://deepest/bar
  image: buildpacks:///baz

```
