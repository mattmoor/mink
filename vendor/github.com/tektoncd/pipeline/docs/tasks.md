<!--
---
linkTitle: "Tasks"
weight: 1
---
-->
# Tasks

- [Overview](#overview)
- [Configuring a `Task`](#configuring-a-task)
  - [`Task` vs. `ClusterTask`](#task-vs-clustertask)
  - [Defining `Steps`](#defining-steps)
    - [Running scripts within `Steps`](#running-scripts-within-steps)
  - [Specifying `Parameters`](#specifying-parameters)
  - [Specifying `Resources`](#specifying-resources)
  - [Specifying `Workspaces`](#specifying-workspaces)
  - [Storing execution results](#storing-execution-results)
  - [Specifying `Volumes`](#specifying-volumes)
  - [Specifying a `Step` template](#specifying-a-step-template)
  - [Specifying `Sidecars`](#specifying-sidecars)
  - [Adding a description](#adding-a-description)
  - [Using variable substitution](#using-variable-substitution)
    - [Substituting parameters and resources](#substituting-parameters-and-resources)
    - [Substituting `Array` parameters](#substituting-array-parameters)
    - [Substituting `Workspace` paths](#substituting-workspace-paths)
    - [Substituting `Volume` names and types](#substituting-volume-names-and-types)
- [Code examples](#code-examples)
  - [Building and pushing a Docker image](#building-and-pushing-a-docker-image)
  - [Mounting multiple `Volumes`](#mounting-multiple-volumes)
  - [Mounting a `ConfigMap` as a `Volume` source](#mounting-a-configmap-as-a-volume-source)
  - [Using a `Secret` as an environment source](#using-a-secret-as-an-environment-source)
  - [Using a `Sidecar` in a `Task`](#using-a-sidecar-in-a-task)
- [Debugging](#debugging)
  - [Inspecting the file structure](#inspecting-the-file-structure)
  - [Inspecting the `Pod`](#inspecting-the-pod)

## Overview

A `Task` is a collection of `Steps` that you
define and arrange in a specific order of execution as part of your continuous integration flow. 
A `Task` executes as a Pod on your Kubernetes cluster. A `Task` is available within a specific
namespace, while a `ClusterTask` is available across the entire cluster.

A `Task` declaration includes the following elements:

- [Parameters](#specifying-parameters)
- [Resources](#specifying-resources)
- [Steps](#defining-steps)
- [Workspaces](#specifying-workspaces)
- [Results](#storing-execution-results)

## Configuring a `Task`

A `Task` definition supports the following fields:

- Required:
  - [`apiVersion`][kubernetes-overview] - Specifies the API version. For example, 
    `tekton.dev/v1beta`.
  - [`kind`][kubernetes-overview] - Identifies this resource object as a `Task` object.
  - [`metadata`][kubernetes-overview] - Specifies metadata that uniquely identifies the
    `Task` resource object. For example, a `name`.
  - [`spec`][kubernetes-overview] - Specifies the configuration information for
    this `Task` resource object. 
  - [`steps`](#defining-steps) - Specifies one or more container images to run in the `Task`.
- Optional:
  - [`description`](#adding-a-description) - An informative description of the `Task`.
  - [`params`](#specifying-parameters) - Specifies execution parameters for the `Task`.
  - [`resources`](#specifying-resources) - **alpha only** Specifies
    [`PipelineResources`](resources.md) needed or created by your`Task`.
    - [`inputs`](#specifying-resources) - Specifies the resources ingested by the `Task`.
    - [`outputs`](#specifying-resources) - Specifies the resources produced by the `Task`.
  - [`workspaces`](#specifying-workspaces) - Specifies paths to volumes required by the `Task`.
  - [`results`](#storing-execution-results) - Specifies the file to which the `Tasks` writes its execution results.
  - [`volumes`](#specifying-volumes) - Specifies one or more volumes that will be available available to the `Steps` in the `Task`.
  - [`stepTemplate`](#specifying-a-step-template) - Specifies a `Container` step definition to use as the basis for all `Steps` in the `Task`.
  - [`sidecars`](#specifying-sidecars) - Specifies `Sidecar` containers to run alongside the `Steps` in the `Task.

[kubernetes-overview]:
  https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/#required-fields

The non-functional example below demonstrates the use of most of the above-mentioned fields:

```yaml
apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: example-task-name
spec:
  params:
    - name: pathToDockerFile
      type: string
      description: The path to the dockerfile to build
      default: /workspace/workspace/Dockerfile
  resources:
    inputs:
      - name: workspace
        type: git
    outputs:
      - name: builtImage
        type: image
  steps:
    - name: ubuntu-example
      image: ubuntu
      args: ["ubuntu-build-example", "SECRETS-example.md"]
    - image: gcr.io/example-builders/build-example
      command: ["echo"]
      args: ["$(params.pathToDockerFile)"]
    - name: dockerfile-pushexample
      image: gcr.io/example-builders/push-example
      args: ["push", "$(resources.outputs.builtImage.url)"]
      volumeMounts:
        - name: docker-socket-example
          mountPath: /var/run/docker.sock
  volumes:
    - name: example-volume
      emptyDir: {}
```

### `Task` vs. `ClusterTask`

A `ClusterTask` is a `Task` scoped to the entire cluster instead of a single namespace.
A `ClusterTask` behaves identically to a `Task` and therefore everything in this document
applies to both.

**Note:** When using a `ClusterTask`, you must explicitly set the `kind` sub-field in the `taskRef` field to `ClusterTask`.
          If not specified, the `kind` sub-field defaults to `Task.` 

Below is an example of a Pipeline declaration that uses a `ClusterTask`:

```yaml
apiVersion: tekton.dev/v1beta1
kind: Pipeline
metadata:
  name: demo-pipeline
  namespace: default
spec:
  tasks:
    - name: build-skaffold-web
      taskRef:
        name: build-push
        kind: ClusterTask
      params: ....
```

### Defining `Steps`

A `Step` is a reference to a container image that executes a specific tool on a
specific input and produces a specific output. To add `Steps` to a `Task` you
define a `steps` field (required) containing a list of desired `Steps`. The order in
which the `Steps` appear in this list is the order in which they will execute.

The following requirements apply to each container image referenced in a `steps` field:

- The container image must abide by the [container contract](./container-contract.md). 
- Each container image runs to completion or until the first failure occurs.
- The CPU, memory, and ephemeral storage resource requests will be set to zero, or, if
  specified, the minimums set through `LimitRanges` in that `Namespace`,
  if the container image does not have the largest resource request out of all
  container images in the `Task.` This ensures that the Pod that executes the `Task`
  only requests enough resources to run a single container image in the `Task` rather
  than hoard resources for all container images in the `Task` at once.

#### Running scripts within `Steps`

A step can specify a `script` field, which contains the body of a script. That script is
invoked as if it were stored inside the container image, and any `args` are passed directly
to it.

**Note:** If the `script` field is present, the step cannot also contain a `command` field.

Scripts that do not start with a [shebang](https://en.wikipedia.org/wiki/Shebang_(Unix))
line will have the following default preamble prepended:

```bash
#!/bin/sh
set -xe
```

You can override this default preamble by prepending a shebang that specifies the desired parser. 
This parser must be present within that `Step's` container image.

The example below executes a Bash script:

```yaml
steps:
- image: ubuntu  # contains bash
  script: |
    #!/usr/bin/env bash
    echo "Hello from Bash!"
```

The example below executes a Python script:

```yaml
steps:
- image: python  # contains python
  script: |
    #!/usr/bin/env python3
    print("Hello from Python!")
```

The example below executes a Node script:

```yaml
steps:
- image: node  # contains node
  script: |
    #!/usr/bin/env node
    console.log("Hello from Node!")
```

You can execute scripts directly in the workspace:

```yaml
steps:
- image: ubuntu
  script: |
    #!/usr/bin/env bash
    /workspace/my-script.sh  # provided by an input resource
```

You can also execute scripts within the container image:

```yaml
steps:
- image: my-image  # contains /bin/my-binary
  script: |
    #!/usr/bin/env bash
    /bin/my-binary
```

### Specifying `Parameters`

You can specify parameters, such as compilation flags or artifact names, that you want to supply to the `Task` at execution time.
 `Parameters` are passed to the `Task` from its corresponding `TaskRun`.

Parameter names:
- Must only contain alphanumeric characters, hyphens (`-`), and underscores (`_`).
- Must begin with a letter or an underscore (`_`).

For example, `fooIs-Bar_` is a valid parameter name, but `barIsBa$` or `0banana` are not.

Each declared parameter has a `type` field, which can be set to either `array` or `string`. `array` is useful in cases where the number
of compiliation flags being supplied to a task varies throughout the `Task's` execution. If not specified, the `type` field defaults to
`string`. When the actual parameter value is supplied, its parsed type is validated against the `type` field.

The following example illustrates the use of `Parameters` in a `Task`. The `Task` declares two input parameters named `flags`
(of type `array`) and `someURL` (of type `string`), and uses them in the `steps.args` list. You can expand parameters of type `array`
inside an existing array using the star operator. In this example, `flags` contains the star operator: `$(params.flags[*])`.

**Note:** Input parameter values can be used as variables throughout the `Task` by using [variable substitution](#using-variable-substitution).

```yaml
apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: task-with-parameters
spec:
  params:
    - name: flags
      type: array
    - name: someURL
      type: string
  steps:
    - name: build
      image: my-builder
      args: ["build", "$(params.flags[*])", "url=$(params.someURL)"]
```

The following `TaskRun` supplies a dynamic number of strings within the `flags` parameter:

```yaml
apiVersion: tekton.dev/v1beta1
kind: TaskRun
metadata:
  name: run-with-parameters
spec:
  taskRef:
    name: task-with-parameters
  params:
    - name: flags
      value:
        - "--set"
        - "arg1=foo"
        - "--randomflag"
        - "--someotherflag"
    - name: someURL
      value: "http://google.com"
```

### Specifying `Resources`

A `Task` definition can specify input and output resources supplied by
a `[`PipelineResources`](resources.md#using-resources) entity.

Use the `input` field to supply your `Task` with the context and/or data it needs to execute.
If the output of your `Task` is also the input of the next `Task` that executes, you must
make that data available to that `Task` at `/workspace/output/resource_name/`. For example:

```yaml
resources:
  outputs:
    name: storage-gcs
    type: gcs
steps:
  - image: objectuser/run-java-jar #https://hub.docker.com/r/objectuser/run-java-jar/
    command: [jar]
    args:
      ["-cvf", "-o", "/workspace/output/storage-gcs/", "projectname.war", "*"]
    env:
      - name: "FOO"
        value: "world"
```

**Note**: If the `Task` relies on output resource functionality then the
containers in the `Task's` `steps` field cannot mount anything in the path
`/workspace/output`.

In the following example, the `tar-artifact` resource is used as both input and
output. Thus, the input resource is copied into the `customworkspace` directory,
as specified in the `targetPath` field. The `untar` `Step` extracts the tarball
into the `tar-scratch-space` directory. The `edit-tar` `Step` adds a new file,
and the `tar-it-up` `Step` creates a new tarball and places it in the
`/workspace/customworkspace/` directory. When the `Task` completes execution,
it places the resulting tarball in the `/workspace/customworkspace` directory
and uploads it to the bucket defined in the `tar-artifact` field.

```yaml
resources:
  inputs:
    name: tar-artifact
    targetPath: customworkspace
  outputs:
    name: tar-artifact
steps:
 - name: untar
    image: ubuntu
    command: ["/bin/bash"]
    args: ['-c', 'mkdir -p /workspace/tar-scratch-space/ && tar -xvf /workspace/customworkspace/rules_docker-master.tar -C /workspace/tar-scratch-space/']
 - name: edit-tar
    image: ubuntu
    command: ["/bin/bash"]
    args: ['-c', 'echo crazy > /workspace/tar-scratch-space/rules_docker-master/crazy.txt']
 - name: tar-it-up
   image: ubuntu
   command: ["/bin/bash"]
   args: ['-c', 'cd /workspace/tar-scratch-space/ && tar -cvf /workspace/customworkspace/rules_docker-master.tar rules_docker-master']
```

### Specifying `Workspaces`

`Workspaces`[workspaces.md#declaring-workspaces-in-tasks] allow you to specify
one or more volumes that your `Task` requires during execution. For example:

```yaml
spec:
  steps:
  - name: write-message
    image: ubuntu
    script: |
      #!/usr/bin/env bash
      set -xe
      echo hello! > $(workspaces.messages.path)/message
  workspaces:
  - name: messages
    description: The folder where we write the message to
    mountPath: /custom/path/relative/to/root
```

For more information, see [Using `Workspaces` in `Tasks`](workspaces.md#using-workspaces-in-tasks)
and the [`Workspaces` in a `TaskRun`](../examples/v1beta1/taskruns/workspace.yaml) example YAML file.

### Storing execution results

Use the `results` field to specify one or more files in which the `Task` stores its execution results. These files are
stored in the `/tekton/results` directory. This directory is created automatically at execution time if at least one file
is specified in the `results` field. To specify a file, provide its `name` and `description`.

In the example below, the `Task` specifies two files in the `results` field:
`current-date-unix-timestamp` and `current-date-human-readable`.

```yaml
apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: print-date
  annotations:
    description: |
      A simple task that prints the date
spec:
  results:
    - name: current-date-unix-timestamp
      description: The current date in unix timestamp format
    - name: current-date-human-readable
      description: The current date in human readable format
  steps:
    - name: print-date-unix-timestamp
      image: bash:latest
      script: |
        #!/usr/bin/env bash
        date +%s | tee /tekton/results/current-date-unix-timestamp
    - name: print-date-humman-readable
      image: bash:latest
      script: |
        #!/usr/bin/env bash
        date | tee /tekton/results/current-date-human-readable
```

**Note:** The maximum size of a `Task's` results is limited by the container termination log feature of Kubernetes,
as results are passed back to the controller via this mechanism. At present, the limit is
["2048 bytes or 80 lines, whichever is smaller."](https://kubernetes.io/docs/tasks/debug-application-cluster/determine-reason-pod-failure/#customizing-the-termination-message).
Results are written to the termination log encoded as JSON objects and Tekton uses those objects
to pass additional information to the controller. As such, `Task` results are best suited for holding
small amounts of data, such as commit SHAs, branch names, ephemeral space names, and so on.

If your `Task` writes a large number of small results, you can work around this limitation
by writing each result from a separate `Step` so that each `Step` has its own termination log.
However, for results larger than a kilobyte, use a [`Workspace`](#specifying-workspaces) to
shuttle data between `Tasks` within a `Pipeline`.

### Specifying `Volumes`

Specifies one or more [`Volumes`](https://kubernetes.io/docs/concepts/storage/volumes/) that the `Steps` in your
`Task` require to execute in addition to volumes that are implicitly created for input and output resources.

For example, you can use `Volumes` to do the following:

- [Mount a Kubernetes `Secret`](auth.md).
- Create an `emptyDir` persistent `Volume` that caches data across multiple `Steps`.
- Mount a [Kubernetes `ConfigMap`](https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/)
  as `Volume` source.
- Mount a host's Docker socket to use a `Dockerfile` for building container images.
  **Note:** Building a container image on-cluster using `docker build` is **very
  unsafe** and is mentioned only for the sake of the example. Use [kaniko](https://github.com/GoogleContainerTools/kaniko) instead.

### Specifying a `Step` template

The `stepTemplate` field specifies a [`Container`](https://kubernetes.io/docs/concepts/containers/)
configuration that will be used as the starting point for all of the `Steps` in your
`Task`. Individual configurations specified within `Steps` supersede the template wherever
overlap occurs.

In the example below, the `Task` specifies a `stepTemplate` field with the environment variable
`FOO` set to `bar`. The first `Step` in the `Task` uses that value for `FOO`, but the second `Step`
overrides the value set in the template with `baz`.

```yaml
stepTemplate:
  env:
    - name: "FOO"
      value: "bar"
steps:
  - image: ubuntu
    command: [echo]
    args: ["FOO is ${FOO}"]
  - image: ubuntu
    command: [echo]
    args: ["FOO is ${FOO}"]
    env:
      - name: "FOO"
        value: "baz"
```

### Specifying `Sidecars`

The `sidecars` field specifies a list of [`Containers`](https://kubernetes.io/docs/concepts/containers/)
to run alongside the `Steps` in your `Task`. You can use `Sidecars` to provide auxiliary functionality, such as
[Docker in Docker](https://hub.docker.com/_/docker) or running a mock API server that your app can hit during testing.
`Sidecars` spin up before your `Task` executes and are deleted after the `Task` execution completes.
For further information, see [`Sidecars` in `TaskRuns`](taskruns.md#sidecars).

In the example below, a `Step` uses a Docker-in-Docker `Sidecar` to build a Docker image:

```yaml
steps:
  - image: docker
    name: client
    script: |
        #!/usr/bin/env bash
        cat > Dockerfile << EOF
        FROM ubuntu
        RUN apt-get update
        ENTRYPOINT ["echo", "hello"]
        EOF
        docker build -t hello . && docker run hello
        docker images
    volumeMounts:
      - mountPath: /var/run/
        name: dind-socket
sidecars:
  - image: docker:18.05-dind
    name: server
    securityContext:
      privileged: true
    volumeMounts:
      - mountPath: /var/lib/docker
        name: dind-storage
      - mountPath: /var/run/
        name: dind-socket
volumes:
  - name: dind-storage
    emptyDir: {}
  - name: dind-socket
    emptyDir: {}
```

Sidecars, just like `Steps`, can also run scripts:

```yaml
sidecars:
  image: busybox
  name: hello-sidecar
  script: |
    echo 'Hello from sidecar!'
```
**Note:** Tekton's current `Sidecar` implementation contains a bug.
Tekton uses a container image named `nop` to terminate `Sidecars`.
That image is configured by passing a flag to the Tekton controller.
If the configured `nop` image contains the exact command the `Sidecar`
was executing before receiving a "stop" signal, the `Sidecar` keeps
running, eventually causing the `TaskRun` to time out with an error.
For more information, see [issue 1347](https://github.com/tektoncd/pipeline/issues/1347).

### Adding a description

The `description` field is an optional field that allows you to add an informative description to the `Task`.

### Using variable substitution

`Tasks` allow you to substitute variable names for the following entities:

- [Parameters and resources]](#substituting-parameters-and-resources)
- [`Array` parameters](#substituting-array-parameters)
- [`Workspaces`](#substituting-workspace-paths)
- [`Volume` names and types](#substituting-volume-names-and-paths)

#### Substituting parameters and resources

[`params`](#specifying-parameters) and [`resources`](#specifying-resources) attributes can replace
variable values as follows:

- To reference a parameter in a `Task`, use the following syntax, where `<name>` is the name of the parameter:
  ```shell
  $(params.<name>)
  ```
- To access parameter values from resources, see [variable substitution](resources.md#variable-substitution)  

#### Substituting `Array' parameters

You can expand referenced paramters of type `array` using the star operator. To do so, add the operator (`[*]`)
to the named parameter to insert the array elements in the spot of the reference string.

For example, given a `params` field with the contents listed below, you can expand
`command: ["first", "$(params.array-param[*])", "last"]` to `command: ["first", "some", "array", "elements", "last"]`:

```yaml
params:
  - name: array-param
    value:
      - "some"
      - "array"
      - "elements"
```

You **must** reference parameters of type `array` in a completely isolated string within a larger `string` array.
Referencing an `array` parameter in any other way will result in an error. For example, if `build-args` is a parameter of
type `array`, then the following example is an invalid `Step` because the string isn't isolated:

```yaml
 - name: build-step
      image: gcr.io/cloud-builders/some-image
      args: ["build", "additionalArg $(params.build-args[*])"]
```

Similarly, referencing `build-args` in a non-`array` field is also invalid:

```yaml
 - name: build-step
      image: "$(params.build-args[*])"
      args: ["build", "args"]
```

A valid reference to the `build-args` parameter is isolated and in an eligible field (`args`, in this case):

```yaml
 - name: build-step
      image: gcr.io/cloud-builders/some-image
      args: ["build", "$(params.build-args[*])", "additonalArg"]
```

#### Substituting `Workspace` paths

You can substitute paths to `Workspaces` specified within a `Task` as follows:

```yaml
$(workspaces.myworkspace.path)
```

Since the `Volume` name is randomized and only set when the `Task` executes, you can also
substitute the volume name as follows:

```yaml
$(workspaces.myworkspace.volume)
```

#### Substituting `Volume` names and types

You can substitute `Volume` names and [types](https://kubernetes.io/docs/concepts/storage/volumes/#types-of-volumes)
by parameterizing them. Tekton supports popular `Volume` types such as `ConfigMap`, `Secret`, and `PersistentVolumeClaim`.
See this [example](#using-kubernetes-configmap-as-volume-source) to find out how to perform this type of substitution
in your `Task.`

## Code examples

Study the following code examples to better understand how to configure your `Tasks`:

- [Building and pushing a Docker image](#building-and-pushing-a-docker-image)
- [Mounting multiple `Volumes`](#mounting-multiple-volumes)
- [Mounting a `ConfigMap` as a `Volume` source](#mounting-a-configmap-as-a-volume-source)
- [Using a `Secret` as an environment source](#using-a-secret-as-an-environment-source)
- [Using a `Sidecar` in a `Task`](#using-a-sidecar-in-a-task)

_Tip: See the collection of simple
[examples](https://github.com/tektoncd/pipeline/tree/master/examples) for
additional code samples._

### Building and pushing a Docker image

The following example `Task` builds and pushes a `Dockerfile`-built image.

**Note:** Building a container image using `docker build` on-cluster is **very
unsafe** and is shown here only as a demonstration. Use [kaniko](https://github.com/GoogleContainerTools/kaniko) instead.

```yaml
spec:
  params:
    # These may be overridden, but provide sensible defaults.
    - name: directory
      type: string
      description: The directory containing the build context.
      default: /workspace
    - name: dockerfileName
      type: string
      description: The name of the Dockerfile
      default: Dockerfile
  resources:
    inputs:
      - name: workspace
        type: git
    outputs:
      - name: builtImage
        type: image
  steps:
    - name: dockerfile-build
      image: gcr.io/cloud-builders/docker
      workingDir: "$(params.directory)"
      args:
        [
          "build",
          "--no-cache",
          "--tag",
          "$(resources.outputs.image.url)",
          "--file",
          "$(params.dockerfileName)",
          ".",
        ]
      volumeMounts:
        - name: docker-socket
          mountPath: /var/run/docker.sock

    - name: dockerfile-push
      image: gcr.io/cloud-builders/docker
      args: ["push", "$(resources.outputs.image.url)"]
      volumeMounts:
        - name: docker-socket
          mountPath: /var/run/docker.sock

  # As an implementation detail, this Task mounts the host's daemon socket.
  volumes:
    - name: docker-socket
      hostPath:
        path: /var/run/docker.sock
        type: Socket
```

#### Mounting multiple `Volumes`

The example below illustrates mounting multiple `Volumes`:

```yaml
spec:
  steps:
    - image: ubuntu
      script: |
        #!/usr/bin/env bash
        curl https://foo.com > /var/my-volume
      volumeMounts:
        - name: my-volume
          mountPath: /var/my-volume

    - image: ubuntu
      script: |
        #!/usr/bin/env bash
        cat /etc/my-volume
      volumeMounts:
        - name: my-volume
          mountPath: /etc/my-volume

  volumes:
    - name: my-volume
      emptyDir: {}
```

#### Mounting a `ConfigMap` as a `Volume` source

The example below illustrates how to mount a `ConfigMap` to act as a `Volume` source:

```yaml
spec:
  params:
    - name: CFGNAME
      type: string
      description: Name of config map
    - name: volumeName
      type: string
      description: Name of volume
  steps:
    - image: ubuntu
      script: |
        #!/usr/bin/env bash
        cat /var/configmap/test
      volumeMounts:
        - name: "$(params.volumeName)"
          mountPath: /var/configmap

  volumes:
    - name: "$(params.volumeName)"
      configMap:
        name: "$(params.CFGNAME)"
```

#### Using a `Secret` as an environment source

The example below illustrates how to use a `Secret` as an environment source:

```yaml
apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: goreleaser
spec:
  params:
  - name: package
    type: string
    description: base package to build in
  - name: github-token-secret
    type: string
    description: name of the secret holding the github-token
    default: github-token
  resources:
    inputs:
    - name: source
      type: git
      targetPath: src/$(params.package)
  steps:
  - name: release
    image: goreleaser/goreleaser
    workingDir: /workspace/src/$(params.package)
    command:
    - goreleaser
    args:
    - release
    env:
    - name: GOPATH
      value: /workspace
    - name: GITHUB_TOKEN
      valueFrom:
        secretKeyRef:
          name: $(params.github-token-secret)
          key: bot-token
```

#### Using a `Sidecar` in a `Task`

The example below illustrates how to use a `Sidecar` in your `Task`:

```yaml
apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: with-sidecar-task
spec:
  params:
  - name: sidecar-image
    type: string
    description: Image name of the sidecar container
  - name: sidecar-env
    type: string
    description: Environment variable value
  sidecars:
  - name: sidecar
    image: $(params.sidecar-image)
    env:
    - name: SIDECAR_ENV
      value: $(params.sidecar-env)
  steps:
  - name: test
    image: hello-world
```

## Debugging

This section describes techniques for debugging the most common issues in `Tasks`.

### Inspecting the file structure

A common issue when configuring `Tasks` stems from not knowing the location of your data.
For the most part, files ingested and output by your `Task` live in the `/workspace` directory,
but the specifics can vary. To inspect the file structure of your `Task`, add a step that outputs
the name of every file stored in the `/workspace` directory to the build log. For example:

```yaml
- name: build-and-push-1
  image: ubuntu
  command:
  - /bin/bash
  args:
  - -c
  - |
    set -ex
    find /workspace
```

You can also choose to examine the *contents* of every file used by your `Task`:

```yaml
- name: build-and-push-1
  image: ubuntu
  command:
  - /bin/bash
  args:
  - -c
  - |
    set -ex
    find /workspace | xargs cat
```

### Inspecting the `Pod`

To inspect the contents of the `Pod` used by your `Task` at a specific stage in the `Task's` execution,
log into the `Pod` and add a `Step` that pauses the `Task` at the desired stage. For example:

```yaml
- name: pause
  image: docker
  args: ["sleep", "6000"]

```

Except as otherwise noted, the contents of this page are licensed under the
[Creative Commons Attribution 4.0 License](https://creativecommons.org/licenses/by/4.0/).
Code samples are licensed under the [Apache 2.0 License](https://www.apache.org/licenses/LICENSE-2.0).
