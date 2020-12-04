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
