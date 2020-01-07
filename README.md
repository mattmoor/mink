# `mink`: a minimal Knative bundle.

`mink` is a minimal distribution of some of the core Knative components.

The upstream Knative distributions keep themselves intentionally loosely coupled and run extensions as separate deployment processes, which can lead to considerable sprawl.

## What's included?

Current:
 - knative/serving: the core components without any extensions

Planned:
 - mattmoor/net-contour
 - knative/eventing
