## `VSphereSource` with `vcsim`

To kick the tires on `VSphereSource` you don't even need a vSphere environment,
you can leverage the `vcsim` simulator!

### Set up `vcsim` on your cluster

You can deploy `vcsim` to your cluster as a Deployment with a Service in front
via:

```shell
ko apply -f vcsim.yaml
```

You can experiment with the arguments passed to `vcsim` to create different
shapes of vSphere environments.

### Set up `vcsim` credentials

To authenticate with `vcsim` we need to create a secret with the fixed
authentication it exposes:

```shell
kubectl apply -f secret.yaml
```

### Set up your receiver

You need a simple HTTP service set up to receive your events. If you have
[knative/serving](https://github.com/knative/serving) installed, we recommend
[sockeye](https://github.com/n3wscott/sockeye):

```shell
kubectl apply -f https://github.com/n3wscott/sockeye/releases/download/v0.4.0/release.yaml
```

If you open sockeye in your browser, it will maintain an open websocket to which
it will forward any events it receives.

### Create your Source!

Now we are going to create the following source:

```yaml
apiVersion: sources.tanzu.vmware.com/v1alpha1
kind: VSphereSource
metadata:
  name: vcsim
spec:
  # If you aren't using Sockeye, then replace this with
  # where you want events sent.
  sink:
    ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: sockeye

  # This points to the Kubernetes service in front of vcsim
  # configured above.
  address: https://vcsim.default.svc.cluster.local
  skipTLSVerify: true
  # This points to the secret created above with the simple
  # vcsim credentials.
  secretRef:
    name: vsphere-credentials
```

> If you are using sockeye, be sure it is open before the next step so that you
> don't miss events!

You can create the source with:

```shell
kubectl apply -f source.yaml
```

### Cleanup

```shell
kubectl delete -f source.yaml
kubectl delete -f secret.yaml
kubectl delete -f vcsim.yaml
```
