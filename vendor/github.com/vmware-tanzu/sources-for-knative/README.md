# VMware Tanzu Sources for Knative

This repo will be the home for VMware-related event sources compatible with the
[Knative](https://knative.dev) project.

[![GoDoc](https://godoc.org/github.com/vmware-tanzu/sources-for-knative?status.svg)](https://godoc.org/github.com/vmware-tanzu/sources-for-knative)
[![Go Report Card](https://goreportcard.com/badge/vmware-tanzu/sources-for-knative)](https://goreportcard.com/report/vmware-tanzu/sources-for-knative)
[![Slack Status](https://img.shields.io/badge/slack-join_chat-white.svg?logo=slack&style=social)](https://knative.slack.com)

This repo is under active development to get a Knative compatible Event Source
for VSphere events, and a Binding to easily access the VSphere API.

To run these examples, you will need [ko](https://github.com/google/ko)
installed.

## Install Source

Install the CRD providing the control / dataplane for the
`VSphere{Source,Binding}`:

```shell
ko apply -f config
```

> Note that currently we require
> [knative/eventing](https://github.com/knative/eventing) to be installed on the
> cluster.

## Samples

To see examples of the Source and Binding in action, check out our
[samples](./samples/README.md) directory.

## Basic `VSphereSource` Example

The `VSphereSource` provides a simple mechanism to enable users to react to
vSphere events.

In order to receive VSphere events there are two key parts:

1. The VSphere address and secret information.
2. Where to send the events.

```yaml
apiVersion: sources.tanzu.vmware.com/v1alpha1
kind: VSphereSource
metadata:
  name: source
spec:
  # Where to fetch the events, and how to auth.
  address: https://my-vsphere-endpoint.local
  skipTLSVerify: true
  secretRef:
    name: vsphere-credentials

  # Where to send the events.
  sink:
    uri: http://where.to.send.stuff
```

Let's walk through each of these.

### Authenticating with vSphere

Let's focus on this part of the sample source:

```yaml
# Where to fetch the events, and how to auth.
address: https://my-vsphere-endpoint.local
skipTLSVerify: true
secretRef:
  name: vsphere-credentials
```

- `address` is the URL of ESXi or vCenter instance to connect to (same as
  `GOVC_URL`).
- `skipTLSVerify` disables certificate verification (same as `GOVC_INSECURE`).
- `secretRef` holds the name of the Kubernetes secret with the following form:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: vsphere-credentials
type: kubernetes.io/basic-auth
stringData:
  # Same as GOVC_USERNAME
  username: ...
  # Same as GOVC_PASSWORD
  password: ...
```

### Delivering Events

Let's focus on this part of the sample source:

```yaml
# Where to send the events.
sink:
  uri: http://where.to.send.stuff
```

The simplest way to deliver events is to simply send them to an accessible
endpoint as above, but there are a few additional options to explore.

To deliver events to a Kubernetes Service in the same namespace you can use:

```yaml
# Where to send the events.
sink:
  ref:
    apiVersion: v1
    kind: Service
    name: the-service-name
```

To deliver events to a [Knative Service](https://knative.dev/docs/serving)
(scales to zero) in the same namespace you can use:

```yaml
# Where to send the events.
sink:
  ref:
    apiVersion: serving.knative.dev/v1
    kind: Service
    name: the-knative-service-name
```

To deliver events to a [Knative Broker](https://knative.dev/docs/eventing) in
the same namespace (e.g. here the `default`) you can use:

```yaml
# Where to send the events.
sink:
  ref:
    apiVersion: eventing.knative.dev/v1beta1
    kind: Broker
    name: default
```

## Basic `VSphereBinding` Example

The `VSphereBinding` provides a simple mechanism for a user application to call
into the vSphere API. In your application code, simply write:

```go
import "github.com/vmware-tanzu/sources-for-knative/pkg/vsphere"

// This returns a github.com/vmware/govmomi.Client
client, err := New(ctx)
if err != nil {
	log.Fatalf("Unable to create vSphere client: %v", err)
}

// This returns a github.com/vmware/govmomi/vapi/rest.Client
restclient, err := New(ctx)
if err != nil {
	log.Fatalf("Unable to create vSphere REST client: %v", err)
}
```

This will authenticate against the bound vSphere environment with the bound
credentials. This same code can be moved to other environments and bound to
different vSphere endpoints without being rebuilt or modified!

Now let's take a look at how `VSphereBinding` makes this possible.

In order to bind an application to a vSphere endpoint, there are two key parts:

1. The VSphere address and secret information (identical to Source above!)
2. The application that is being bound (aka the "subject").

```yaml
apiVersion: sources.tanzu.vmware.com/v1alpha1
kind: VSphereBinding
metadata:
  name: binding
spec:
  # Where to fetch the events, and how to auth.
  address: https://my-vsphere-endpoint.local
  skipTLSVerify: true
  secretRef:
    name: vsphere-credentials

  # Where to bind the endpoint and credential data.
  subject:
    apiVersion: apps/v1
    kind: Deployment
    name: my-simple-app
```

Authentication is identical to source, so let's take a deeper look at subjects.

### Binding applications (aka subjects)

Let's focus on this part of the sample binding:

```yaml
# Where to bind the endpoint and credential data.
subject:
  apiVersion: apps/v1
  kind: Deployment
  name: my-simple-app
```

In this simple example, the binding is going to inject several environment
variables and secret volumes into the containers in this exact Deployment
resource.

If you would like to target a _selection_ of resources you can also write:

```yaml
# Where to bind the endpoint and credential data.
subject:
  apiVersion: batch/v1
  kind: Job
  selector:
    matchLabels:
      foo: bar
```

Here the binding will apply to every `Job` in the same namespace labeled
`foo: bar`, so this can be used to bind every `Job` stamped out by a `CronJob`
resource.

At this point, you might be wondering: what kinds of resources does this
support? We support binding all resources that embed a Kubernetes PodSpec in the
following way (standard Kubernetes shape):

```yaml
spec:
  template:
    spec: # This is a Kubernetes PodSpec.
      containers:
      - image: ...
      ...
```

This has been tested with:

- Knative `Service` and `Configuration`
- `Deployment`
- `Job`
- `DaemonSet`
- `StatefulSet`
