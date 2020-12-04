# Using `mink run`

`mink run` allows users to instantiate named Tekton Task and Pipeline resources
via `mink run task NAME` and `mink run pipeline NAME` respectively. How they
work is largely the same, just Tasks vs. Pipelines.

## Usage

If a task (similarly for pipelines) takes no arguments, it can be invoked with:

```shell
$ mink run task NAME
```

The usage of the task (or pipeline) itself can be examined with:

```shell
$ mink run task hello -- --help
Says hello and stuff

Usage:
  mink run task hello [flags]

Flags:
      --greeting string   The greeting to use (default "Hello")
  -h, --help              help for mink
      --name string       The name of the person to greet.
  -o, --output string     options: message
```

This usage draws all of its metadata from the task definition itself. The task
description, the parameters (descriptions and defaults), and the outputs
(results).

## Example

To try things out, you can install the task `examples/task-hello.yaml` and
invoke it with:

```shell
$  mink run task hello -- --name Bill
[echo] Hello, Bill
```

You can project named results with `-oNAME`:

```shell
$ mink run task hello -- --name Bill -omessage
[echo] Hello, Bill

Hello, Bill
```

The result (`-oNAME`) will be sent to stdout, where the log output will be sent
to stderr, so you can capture or compose the result while still seeing logs.


## Deeper Task/Pipeline Integration

`mink` takes the simple interface above one step further, and provides a set
of parameters and results that can activate additional functionality in the
`mink` CLI.  These special names have been intentionally namespaced to avoid
collisions with pre-existing parameters.

### Example

For example, if you want to take advantage of `mink bundle`'s ability to upload
source context, you can add the following to your task or pipelines signature:

```yaml
  params:
    - name: mink-source-bundle
      description: A self-extracting container image.
```

This parameter name instructs `mink run` to bundle up the source context using
`mink bundle` and pass the resulting image digest as this parameter for your
task to consume.

A typical task would then make its first step:

```yaml
  steps:
    - name: extract-bundle
      image: $(params.mink-source-bundle)
```

Subsequent steps will see the source context in the working directory.

For a more in-depth example that puts several of these constructs together,
see the [kaniko task](./examples/kaniko.yaml) example.

### Special Parameters

#### `mink-source-bundle`

As outlined above, when this parameter is part of the task or pipeline's
signature it will trigger the `mink bundle` functionality.  `mink bundle`
produces a container image with source from either a local `--directory`
or a remote `--git-url`, which is extracted into the working directory
the resulting container is run in.

For examples of how to use this functionality see
[task](./examples/task-bundle.yaml) or
[pipeline](./examples/pipeline-bundle.yaml).


#### `mink-image-target`

When this parameter is present, the template in `--image` will be instantiated
and the resulting URI will be passed through this parameter to the Task
or Pipeline to use as it sees fite.

For examples of how to use this functionality see
[task](./examples/task-image.yaml) or
[pipeline](./examples/pipeline-image.yaml).

### Special Results

#### `mink-image-digest`

This does not yet activate any special functionality in `mink run`, but is
the result name used in `mink build` and friends.
