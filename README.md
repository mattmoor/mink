# `mink`: a minimal Knative bundle.

`mink` is a distribution of Knative and Tekton components.

## How?

You can install `mink` by running `ko apply -R -f config` (assuming you have
properly configured [`ko`](https://github.com/google/ko)).

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

The dataplane components, including the Contour envoys and the activator are run
as a DaemonSet to scale with the cluster.

## What?

Current:

- knative/serving: the core components, HPA-class autoscaling, and the
  default-domain job. No cert-manager, no nscert, or Istio controllers are
  included.
- knative/eventing: sink binding, API server source, and ping source.
- knative/net-contour: The Contour KIngress controller is now linked into our
  controller webhook.
- projectcontour/contour: A heavily customized Contour installation curated to
  facilitate `mink`.
- mattmoor/http01-solver: A simple ACME HTTP01-based certificate provisioner
  (requires real DNS to be set up).

Planned:

- knative/eventing: flows, broker/trigger, channel/subscription.
