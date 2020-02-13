# `mink`: a minimal Knative bundle.

`mink` is a minimal distribution of some of the core Knative components.

## How?

You can install `mink` by running `ko apply -R -f config` (assuming you have properly configured [`ko`](https://github.com/google/ko)).


## Why?

The upstream Knative distributions keep themselves intentionally loosely coupled and run extensions as separate deployment processes, which can lead to considerable sprawl.

`mink` folds many of these components together:

```
NAMESPACE        NAME                              READY   STATUS    RESTARTS   AGE
knative-system   pod/activator-6ss24               3/3     Running   0          12m
knative-system   pod/activator-9crg2               3/3     Running   0          12m
knative-system   pod/activator-tzxsx               3/3     Running   0          12m
knative-system   pod/autoscaler-fdc565c86-frgzf    1/1     Running   0          12m
knative-system   pod/contour-76778db4c5-25g5m      2/2     Running   0          13m
knative-system   pod/controller-859c5757c8-l9vkl   1/1     Running   0          12m
```

The dataplane components, including the Contour envoys and the activator are run as a DaemonSet to scale with the cluster.


## What?

Current:
 - knative/serving: the core components, HPA-class autoscaling, namespace wildcard cert controller, and the default-domain job.  No cert-manager, or Istio controllers are included.
 - knative/eventing: sink binding, API server source, and ping source.
 - knative/net-contour: The Contour KIngress controller is now linked into our controller webhook.
 - projectcontour/contour: A heavily customized Contour installation curated to facilitate `mink`.

> Plans for TLS are evolving (see https://github.com/mattmoor/mink/issues/4), so we may drop the namespace wildcard cert controller.

Planned:
 - knative/eventing: flows, broker/trigger, channel/subscription.
 - tekton/pipelines: This is blocked on them updating to the latest knative/pkg, so they can simply be linked in.
 - I'd love to see a simple Certificate controller for HTTP01 without pulling in all of cert-manager (ideally it would fold into our shared controller)
