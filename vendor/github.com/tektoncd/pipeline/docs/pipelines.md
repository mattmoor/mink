<!--
---
linkTitle: "Pipelines"
weight: 3
---
-->
# Pipelines

This document defines `Pipelines` and their capabilities.

---

- [Pipelines](#pipelines)
  - [Syntax](#syntax)
    - [Description](#description)
    - [Declared resources](#declared-resources)
    - [Workspaces](#workspaces)
    - [Parameters](#parameters)
      - [Usage](#usage)
    - [Pipeline Tasks](#pipeline-tasks)
      - [from](#from)
      - [runAfter](#runafter)
      - [retries](#retries)
      - [conditions](#conditions)
      - [Timeout](#timeout)
    - [Results](#results)
    - [Ordering](#ordering)
  - [Examples](#examples)

## Syntax

To define a configuration file for a `Pipeline` resource, you can specify the
following fields:

- Required:
  - [`apiVersion`][kubernetes-overview] - Specifies the API version, for example
    `tekton.dev/v1beta1`.
  - [`kind`][kubernetes-overview] - Specify the `Pipeline` resource object.
  - [`metadata`][kubernetes-overview] - Specifies data to uniquely identify the
    `Pipeline` resource object, for example a `name`.
  - [`spec`][kubernetes-overview] - Specifies the configuration information for
    your `Pipeline` resource object. In order for a `Pipeline` to do anything,
    the spec must include:
    - [`tasks`](#pipeline-tasks) - Specifies which `Tasks` to run and how to run
      them
- Optional:
  - [`description`](#description) - Description of the Pipeline.
  - [`resources`](#declared-resources) - Specifies which
    [`PipelineResources`](resources.md) of which types the `Pipeline` will be
    using in its [Tasks](#pipeline-tasks)
  - `tasks`
      - `resources.inputs` / `resource.outputs`
          - [`from`](#from) - Used when the content of the
            [`PipelineResource`](resources.md) should come from the
            [output](tasks.md#outputs) of a previous [Pipeline Task](#pipeline-tasks)
      - [`runAfter`](#runAfter) - Used when the [Pipeline Task](#pipeline-tasks)
        should be executed after another Pipeline Task, but there is no
        [output linking](#from) required
      - [`retries`](#retries) - Used when the task is wanted to be executed if
        it fails. Could be a network error or a missing dependency. It does not
        apply to cancellations.
      - [`conditions`](#conditions) - Used when a task is to be executed only if the specified
        conditions are evaluated to be true.
      - [`timeout`](#timeout) - Specifies timeout after which the `TaskRun` for a Pipeline Task will
        fail. There is no default timeout for a Pipeline Task timeout. If no timeout is specified for
        the Pipeline Task, the only timeout taken into account for running a `Pipeline` will be a
        [timeout for the `PipelineRun`](https://github.com/tektoncd/pipeline/blob/master/docs/pipelineruns.md#syntax).
  - [`results`](#pipeline-results) - Specifies which `results` is defined for the pipeline

[kubernetes-overview]:
  https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/#required-fields

### Description

The `description` field is an optional field and can be used to provide description of the Pipeline.

### Declared resources

In order for a `Pipeline` to interact with the outside world, it will probably
need [`PipelineResources`](resources.md) which will be given to
`Tasks` as inputs and outputs.

Your `Pipeline` must declare the `PipelineResources` it needs in a `resources`
section in the `spec`, giving each a name which will be used to refer to these
`PipelineResources` in the `Tasks`.

For example:

```yaml
spec:
  resources:
    - name: my-repo
      type: git
    - name: my-image
      type: image
```

### Workspaces

`workspaces` are a way of declaring volumes you expect to be made available to your
executing `Pipeline` and its `Task`s.

Here's a short example of a Pipeline Spec with `workspaces`:

```yaml
spec:
  workspaces:
    - name: pipeline-ws1 # The name of the workspace in the Pipeline
  tasks:
    - name: use-ws-from-pipeline
      taskRef:
        name: gen-code # gen-code expects a workspace with name "output"
      workspaces:
        - name: output
          workspace: pipeline-ws1
    - name: use-ws-again
      taskRef:
        name: commit # commit expects a workspace with name "src"
      runAfter:
        - use-ws-from-pipeline # important: use-ws-from-pipeline writes to the workspace first
      workspaces:
        - name: src
          workspace: pipeline-ws1
```

For complete documentation on using `workspaces` in `Pipeline`s, see
[workspaces.md](./workspaces.md#declaring-workspaces-in-pipelines).

_For a complete example see [the Workspaces PipelineRun](../examples/v1beta1/pipelineruns/workspaces.yaml)
in the examples directory._

### Parameters

`Pipeline`s can declare input parameters that must be supplied to the `Pipeline`
during a `PipelineRun`. Pipeline parameters can be used to replace template
values in [`PipelineTask` parameters' values](#pipeline-tasks).

Parameter names are limited to alpha-numeric characters, `-` and `_` and can
only start with alpha characters and `_`. For example, `fooIs-Bar_` is a valid
parameter name, `barIsBa$` or `0banana` are not.

Each declared parameter has a `type` field, assumed to be `string` if not provided by the user. The other possible type is `array` — useful, for instance, when a dynamic number of string arguments need to be supplied to a task. When the actual parameter value is supplied, its parsed type is validated against the `type` field.

#### Usage

The following example shows how `Pipeline`s can be parameterized, and these
parameters can be passed to the `Pipeline` from a `PipelineRun`.

Input parameters in the form of `$(params.foo)` are replaced inside of the
[`PipelineTask` parameters' values](#pipeline-tasks) (see also
[variable substitution](tasks.md#variable-substitution)).

The following `Pipeline` declares an input parameter called 'context', and uses
it in the `PipelineTask`'s parameter. The `description` and `default` fields for
a parameter are optional, and if the `default` field is specified and this
`Pipeline` is used by a `PipelineRun` without specifying a value for 'context',
the `default` value will be used.

```yaml
apiVersion: tekton.dev/v1beta1
kind: Pipeline
metadata:
  name: pipeline-with-parameters
spec:
  params:
    - name: context
      type: string
      description: Path to context
      default: /some/where/or/other
  tasks:
    - name: build-skaffold-web
      taskRef:
        name: build-push
      params:
        - name: pathToDockerFile
          value: Dockerfile
        - name: pathToContext
          value: "$(params.context)"
```

The following `PipelineRun` supplies a value for `context`:

```yaml
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  name: pipelinerun-with-parameters
spec:
  pipelineRef:
    name: pipeline-with-parameters
  params:
    - name: "context"
      value: "/workspace/examples/microservices/leeroy-web"
```

### Pipeline Tasks

A `Pipeline` will execute a graph of [`Tasks`](tasks.md) (see
[ordering](#ordering) for how to express this graph). A valid `Pipeline`
declaration must include a reference to at least one [`Task`](tasks.md). Each
`Task` within a `Pipeline` must have a
[valid](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names)
name and task reference, for example:

```yaml
tasks:
  - name: build-the-image
    taskRef:
      name: build-push
```

[Declared `PipelineResources`](#declared-resources) can be given to `Task`s in
the `Pipeline` as inputs and outputs, for example:

```yaml
spec:
  tasks:
    - name: build-the-image
      taskRef:
        name: build-push
      resources:
        inputs:
          - name: workspace
            resource: my-repo
        outputs:
          - name: image
            resource: my-image
```

[Parameters](tasks.md#parameters) can also be provided:

```yaml
spec:
  tasks:
    - name: build-skaffold-web
      taskRef:
        name: build-push
      params:
        - name: pathToDockerFile
          value: Dockerfile
        - name: pathToContext
          value: /workspace/examples/microservices/leeroy-web
```

#### from

Sometimes you will have [Pipeline Tasks](#pipeline-tasks) that need to take as
input the output of a previous `Task`, for example, an image built by a previous
`Task`.

Express this dependency by adding `from` on [`PipelineResources`](resources.md)
that your `Tasks` need.

- The (optional) `from` key on an `input source` defines a set of previous
  `PipelineTasks` (i.e. the named instance of a `Task`) in the `Pipeline`
- When the `from` key is specified on an input source, the version of the
  resource that is from the defined list of tasks is used
- `from` can support fan in and fan out
- The `from` clause [expresses ordering](#ordering), i.e. the
  [Pipeline Task](#pipeline-tasks) which provides the `PipelineResource` must run
  _before_ the Pipeline Task which needs that `PipelineResource` as an input
  - The name of the `PipelineResource` must correspond to a `PipelineResource`
    from the `Task` that the referenced `PipelineTask` gives as an output

For example see this `Pipeline` spec:

```yaml
- name: build-app
  taskRef:
    name: build-push
  resources:
    outputs:
      - name: image
        resource: my-image
- name: deploy-app
  taskRef:
    name: deploy-kubectl
  resources:
    inputs:
      - name: image
        resource: my-image
        from:
          - build-app
```

The resource `my-image` is expected to be given to the `deploy-app` `Task` from
the `build-app` `Task`. This means that the `PipelineResource` `my-image` must
also be declared as an output of `build-app`.

This also means that the `build-app` Pipeline Task will run before `deploy-app`,
regardless of the order they appear in the spec.

#### runAfter

Sometimes you will need to have [Pipeline Tasks](#pipeline-tasks) that need to
run in a certain order, but they do not have an explicit
[output](tasks.md#outputs) to [input](tasks.md#inputs) dependency (which is
expressed via [`from`](#from)). In this case you can use `runAfter` to indicate
that a Pipeline Task should be run after one or more previous Pipeline Tasks.

For example see this `Pipeline` spec:

```yaml
- name: test-app
  taskRef:
    name: make-test
  resources:
    inputs:
      - name: workspace
        resource: my-repo
- name: build-app
  taskRef:
    name: kaniko-build
  runAfter:
    - test-app
  resources:
    inputs:
      - name: workspace
        resource: my-repo
```

In this `Pipeline`, we want to test the code before we build from it, but there
is no output from `test-app`, so `build-app` uses `runAfter` to indicate that
`test-app` should run before it, regardless of the order they appear in the
spec.

#### retries

Sometimes you need a policy for retrying tasks which have problems such as
network errors, missing dependencies or upload problems. Any of those issues must
be reflected as False (corev1.ConditionFalse) within the TaskRun Status
Succeeded Condition. For that reason there is an optional attribute called
`retries` which declares how many times that task should be retried in case of
failure.

By default and in its absence there are no retries; its value is 0.

```yaml
tasks:
  - name: build-the-image
    retries: 1
    taskRef:
      name: build-push
```

In this example, the task "build-the-image" will be executed and if the first
run fails a second one would triggered. But, if that fails no more would
triggered: a max of two executions.

#### conditions

Sometimes you will need to run tasks only when some conditions are true. The `conditions` field
allows you to list a series of references to [`Conditions`](./conditions.md) that are run before the task
is run. If all of the conditions evaluate to true, the task is run. If any of the conditions are false,
the Task is not run. Its status.ConditionSucceeded is set to False with the reason set to  `ConditionCheckFailed`.
However, unlike regular task failures, condition failures do not automatically fail the entire pipeline
run -- other tasks that are not dependent on the task (via `from` or `runAfter`) are still run.

```yaml
tasks:
  - name: conditional-task
    taskRef:
      name: build-push
    conditions:
      - conditionRef: my-condition
        params:
          - name: my-param
            value: my-value
        resources:
          - name: workspace
            resource: source-repo
```

In this example, `my-condition` refers to a [Condition](conditions.md) custom resource. The `build-push`
task will only be executed if the condition evaluates to true.

Resources in conditions can also use the [`from`](#from) field to indicate that they
expect the output of a previous task as input. As with regular Pipeline Tasks, using `from`
implies ordering --  if task has a condition that takes in an output resource from
another task, the task producing the output resource will run first:

```yaml
tasks:
  - name: first-create-file
    taskRef:
      name: create-file
    resources:
      outputs:
        - name: workspace
          resource: source-repo
  - name: then-check
    conditions:
      - conditionRef: "file-exists"
        resources:
          - name: workspace
            resource: source-repo
            from: [first-create-file]
    taskRef:
      name: echo-hello
```

#### Timeout

The Timeout property of a Pipeline Task allows a timeout to be defined for a `TaskRun` that
is part of a `PipelineRun`. If the `TaskRun` exceeds the amount of time specified, the `TaskRun`
will fail and the `PipelineRun` associated with a `Pipeline` will fail as well.

There is no default timeout for Pipeline Tasks, so a timeout must be specified with a Pipeline Task
when defining a `Pipeline` if one is needed. An example of a Pipeline Task with a Timeout is shown below:

```yaml
spec:
  tasks:
    - name: build-the-image
      taskRef:
        name: build-push
      Timeout: "0h1m30s"
```

The Timeout property is specified as part of the Pipeline Task on the `Pipeline` spec. The above
example has a timeout of one minute and 30 seconds.

#### Results

Tasks can declare [results](./tasks.md#results) that they will emit during their execution. These results can be used as values for params in subsequent tasks of a Pipeline. Tekton will infer the ordering of these Tasks to ensure that the Task emitting the results runs before the Task consuming those results in its parameters.

Using a Task result as a value for another Task's parameter is done with variable substitution. Here is what a Pipeline Task's param looks like with a result wired into it:

```yaml
params:
  - name: foo
    value: "$(tasks.previous-task-name.results.bar-result)"
```

In this example the previous pipeline task has name "previous-task-name" and its result is declared in the Task definition as having name "bar-result".

For a complete example demonstrating Task Results in a Pipeline see the [pipelinerun example](../examples/v1beta1/pipelineruns/task_results_example.yaml).

#### Ordering

The [Pipeline Tasks](#pipeline-tasks) in a `Pipeline` can be connected and run
in a graph, specifically a _Directed Acyclic Graph_ or DAG. Each of the Pipeline
Tasks is a node, which can be connected with an edge (i.e. a _Graph_) such that one will run
before another (i.e. _Directed_), and the execution will eventually complete
(i.e. _Acyclic_, it will not get caught in infinite loops).

This is done using:

- [`from`](#from) clauses on the [`PipelineResources`](resources.md) needed by a
  `Task`
- [`runAfter`](#runAfter) clauses on the [Pipeline Tasks](#pipeline-tasks)

For example see this `Pipeline` spec:

```yaml
- name: lint-repo
  taskRef:
    name: pylint
  resources:
    inputs:
      - name: workspace
        resource: my-repo
- name: test-app
  taskRef:
    name: make-test
  resources:
    inputs:
      - name: workspace
        resource: my-repo
- name: build-app
  taskRef:
    name: kaniko-build-app
  runAfter:
    - test-app
  resources:
    inputs:
      - name: workspace
        resource: my-repo
    outputs:
      - name: image
        resource: my-app-image
- name: build-frontend
  taskRef:
    name: kaniko-build-frontend
  runAfter:
    - test-app
  resources:
    inputs:
      - name: workspace
        resource: my-repo
    outputs:
      - name: image
        resource: my-frontend-image
- name: deploy-all
  taskRef:
    name: deploy-kubectl
  resources:
    inputs:
      - name: my-app-image
        resource: my-app-image
        from:
          - build-app
      - name: my-frontend-image
        resource: my-frontend-image
        from:
          - build-frontend
```

This will result in the following execution graph:

```none
        |            |
        v            v
     test-app    lint-repo
    /        \
   v          v
build-app  build-frontend
   \          /
    v        v
    deploy-all
```

1. The `lint-repo` and `test-app` Pipeline Tasks will begin executing
   simultaneously. (They have no `from` or `runAfter` clauses.)
1. Once `test-app` completes, both `build-app` and `build-frontend` will begin
   executing simultaneously (both `runAfter` `test-app`).
1. When both `build-app` and `build-frontend` have completed, `deploy-all` will
   execute (it requires `PipelineResources` from both Pipeline Tasks).
1. The entire `Pipeline` will be finished executing after `lint-repo` and
   `deploy-all` have completed.

### Pipeline Results

A pipeline can declare results that they will emit during their execution. These results can be defined as reference to task results executed
during the pipeline execution.

```yaml
  results:
    - name: sum
      description: the sum of all three operands
      value: $(tasks.second-add.results.sum)
```

In this example the pipeline result has name "sum" and its result is declared as the task result value from the tasks named `second-add`.

For a complete example demonstrating pipeline Results in a Pipeline see the [pipeline example](../examples/pipelineruns/pipelinerun-results.yaml).

## Examples

For complete examples, see
[the examples folder](https://github.com/tektoncd/pipeline/tree/master/examples).

---

Except as otherwise noted, the content of this page is licensed under the
[Creative Commons Attribution 4.0 License](https://creativecommons.org/licenses/by/4.0/),
and code samples are licensed under the
[Apache 2.0 License](https://www.apache.org/licenses/LICENSE-2.0).
