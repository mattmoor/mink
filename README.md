# `mink`: a minimal Knative bundle.

`mink` is a distribution of Knative and Tekton components.

## How?

You can install `mink` by running `ko apply -R -f config/core` (assuming you have
properly configured [`ko`](https://github.com/google/ko)).


`mink` also provides a modified version of `InMemoryChannel` that can run alongside it
which can be installed with `ko apply -R -f config/in-memory`.  This channel is
unsuitable for production use cases, but is a nice lightweight option for development.

## Why?

The upstream Knative distribution keeps itself intentionally loosely coupled and
runs extensions as separate Deployment processes, which can lead to considerable
sprawl.

`mink` folds many of these components back together:

```
NAMESPACE     NAME                              READY   STATUS    RESTARTS   AGE
mink-system   pod/activator-6ss24               3/3     Running   0          12m
mink-system   pod/activator-9crg2               3/3     Running   0          12m
mink-system   pod/activator-tzxsx               3/3     Running   0          12m
mink-system   pod/autoscaler-fdc565c86-frgzf    1/1     Running   0          12m
mink-system   pod/controller-859c5757c8-l9vkl   3/3     Running   0          12m
```

_With the in-memory channel, you also get the controller and dispatched pods_


The dataplane components, including the Contour envoys and the activator are run
as a DaemonSet to scale with the cluster.

## What?

Current (**included**):

- knative/serving: the core components, HPA-class autoscaling, and the
  default-domain job. No cert-manager, no nscert, or Istio controllers are
  included.
- knative/eventing: sink binding, API server source, ping source, channel/subscription, broker/trigger.
- knative/net-contour: The Contour KIngress controller is now linked into our
  controller webhook.
- projectcontour/contour: A heavily customized Contour installation curated to
  facilitate `mink`.
- mattmoor/http01-solver: A simple ACME HTTP01-based certificate provisioner
  (requires real DNS to be set up).

Current (**optional**):
- knative/eventing: in-memory channel


Planned:

- knative/eventing: flows
