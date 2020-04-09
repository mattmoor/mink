# `mink`: a minimal Knative bundle.

`mink` is a distribution of Knative and Tekton components.

## How?

You can install `mink` by running `ko apply -R -f config/core` (assuming you
have properly configured [`ko`](https://github.com/google/ko)).

`mink` also provides a modified version of `InMemoryChannel` that can run
alongside it which can be installed with `ko apply -R -f config/in-memory`. This
channel is unsuitable for production use cases, but is a nice lightweight option
for development.

## Why?

The upstream Knative distribution keeps itself intentionally loosely coupled and
runs extensions as separate Deployment processes, which can lead to considerable
sprawl.

`mink` folds many of these components back together:

```
NAMESPACE     NAME                              READY   STATUS    RESTARTS   AGE
mink-system   autoscaler-6564969cd6-2r7fg       1/1     Running   0          2m49s
mink-system   controlplane-64787d66cd-xh55w     3/3     Running   0          2m35s
mink-system   dataplane-7lzqd                   5/5     Running   0          2m35s
mink-system   dataplane-kmdvf                   5/5     Running   0          2m35s
mink-system   dataplane-w2d96                   5/5     Running   0          2m35s
```

_With the in-memory channel, you also get the controller and dispatched pods_

The dataplane components, including the Contour envoys, the activator, and the
broker ingress/filter are run as a DaemonSet to scale with the cluster.

## What?

Current (**included**):

- knative/serving: the core components, HPA-class autoscaling, and the
  default-domain job. No cert-manager, no nscert, or Istio controllers are
  included.
- knative/eventing: sink binding, API server source, ping source,
  channel/subscription, broker(mt)/trigger.
- knative/eventing: github, and kafka sources
- knative/net-contour: The Contour KIngress controller is now linked into our
  controller webhook.
- knative/net-http01: A simple ACME HTTP01-based certificate provisioner
  (requires real DNS to be set up).
- projectcontour/contour: A heavily customized Contour installation curated to
  facilitate `mink`.
- vmware-tanzu/sources-for-knative: VMware source and binding.
- mattmoor/bindings: Experimental bindings for Github, Slack, Twitter, and SQL.

Current (**optional**):

- knative/eventing: in-memory channel

Planned:

- knative/eventing: flows
