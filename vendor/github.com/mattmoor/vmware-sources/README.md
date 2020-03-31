# VSphere Knative Event source

[![GoDoc](https://godoc.org/knative.dev/sample-controller?status.svg)](https://godoc.org/knative.dev/sample-controller)
[![Go Report Card](https://goreportcard.com/badge/knative/sample-controller)](https://goreportcard.com/report/knative/sample-controller)

This repo is under active development to get a Knative compatible Event Source
for VSphere events.

To run these examples, you will need ko installed.

## Install Source

Install the CRD providing the control / dataplane for the VSphereSource:

```shell
ko apply -f config
```

## (Optional) Install a VSphere simulator (vcsim)

If you do not have access to a VSphere environment, you can use
[vcsim](https://github.com/vmware/govmomi/tree/master/vcsim) to
simulate events. If you have access to VSphere environment, you
can skip this step and use that instead. This installs the
simulator to the default namespace and exposes it as a k8s service
called `vcsim`.

```shell
ko apply -f samples/vcsim.yaml
```

### (Optional) Visualizing events with [sockeye](https://github.com/n3wscott/sockeye)

Sockeye is a tool for inspecting CloudEvents. This is an easy way to verify
source is configured correctly and producing events.

```shell
kubectl apply -f https://github.com/n3wscott/sockeye/releases/download/v0.3.0/release.yaml
```

### Install Secrets for accessing VSphere

Source needs k8s secrets with the Credentials for the VSphere to be able
to receive events from it. There's a sample in samples/secret.yaml configured
to connect to vcsim, so if you are trying to connect to existing VSphere,
you should modify it with your actual credentials.

```
ko apply -f ./samples/secret.yaml
```

### Install Source

We need to tell the Source where to get the VSphere events from,
If you are trying to connect to existing VSphere, you should modify
it accordingly in the spec.address field in the
samples/vsphere-source.yaml.

```
ko apply -f ./samples/vsphere-source.yaml
```

### Consume events

In order to consume events, you need to create a Trigger. This example
sends events to Knative Service that just dumps them into logs. Note
that if you're using the vcsim source, you're going to have to kill
the pods to get the events as it only sends a batch of events and will
not emit more events.

```
ko apply -f ./samples/event-display.yaml
```

You can see the events in the logs:

```
kubectl logs -l 'serving.knative.dev/service=event-display' -c user-container
```

To learn more about Knative, please visit our
[Knative docs](https://github.com/knative/docs) repository.

If you are interested in contributing, see [CONTRIBUTING.md](./CONTRIBUTING.md)
and [DEVELOPMENT.md](./DEVELOPMENT.md).
