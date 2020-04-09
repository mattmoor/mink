## Using `VSphereSource` and `VSphereBinding` to tag new VMs

This builds on the prior samples that demonstrate the
[source](../vcsim/README.md) and [binding](../govc/README.md) in isolation. In
this sample we will combine these concepts to write a microservice that
reacts to `VmCreatedEvent` by tagging the new VM.

### Pre-requisites

This sample assumes that you have a vSphere environment set up already
with credentials in a Secret named `vsphere-credentials`.  For the remainder
of the sample we will assume you are within the environment setup for the
[`vcsim` sample](../vcsim/README.md), and that you have created the tag from
the [`govc` sample](../govc/README.md).

This sample will make use of both Knative Serving and Eventing, so make sure
both are installed, and that you have enabled the Broker on the `default`
namespace.

### Create your Source

Now we are going to create the following source:

```yaml
apiVersion: sources.tanzu.vmware.com/v1alpha1
kind: VSphereSource
metadata:
 name: vcsim-to-broker
spec:
  # Unlike the prior sample, we are going to make use of the
  # Knative Eventing's Broker concept to let us react to specific
  # events.
  sink:
    ref:
      apiVersion: eventing.knative.dev/v1
      kind: Broker
      name: default

  # The connection information for vSphere (we will not cover this in detail)
  address: https://vcsim.default.svc.cluster.local
  skipTLSVerify: true
  secretRef:
    name: vsphere-credentials
```

You can create the source with:

```shell
kubectl apply -f source.yaml
```

### Create your Binding

Now we are going to create the following binding:

```yaml
apiVersion: sources.tanzu.vmware.com/v1alpha1
kind: VSphereBinding
metadata:
  name: vsphere-functions
spec:
  # Apply this binding to all Knative services labeled with:
  #   role: vsphere-fn
  subject:
    apiVersion: serving.knative.dev/v1
    kind: Service
    selector:
      matchLabels:
        role: vsphere-fn

  # The connection information for vSphere (we will not cover this in detail)
  address: https://vcsim.default.svc.cluster.local
  skipTLSVerify: true
  secretRef:
    name: vsphere-credentials
```

You can create the binding with:

```shell
kubectl apply -f binding.yaml
```


### Create your Service

Now we are going to write a small service that we'll use to listen to
`VmCreatedEvent`s and tag the new VMs.  Let's start by looking at the
code to handle the event, and then look at how we wire that up to receive
the appropriate events.

With the binding we can create the client to tag VMs with a few lines:

```go
import (
	...
	"github.com/vmware-tanzu/sources-for-knative/pkg/vsphere"
)

...
func main() {
	...
	// Instantiate a client for interacting with the vSphere APIs.
	client, err := vsphere.NewREST(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}
	r := &receiver{manager: tags.NewManager(client)}
...
```

A few more to tag the VM in the Cloud Event handler:

```go
func (r *receiver) handle(ctx context.Context, event cloudevents.Event) error {
	// Parse the event we received.
	req := &types.VmCreatedEvent{}
	if err := event.DataAs(&req); err != nil {
		return err
	}
	log.Printf("Tagging VM: %v", req.Vm.Vm)
	// Attach the tag from the `govc` sample to the VM!
	return r.manager.AttachTag(ctx, "shrug", req.Vm.Vm)
}
```

We wrap this up in a Knative Service that we have labeled to receive the binding:

```yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: tag-new-vms
  labels:
    # We want the VSphereBinding to give us
    # credentials for talking to VSphere.
    role: vsphere-fn
    # This function should not be exposed
    # outside of the cluster.
    serving.knative.dev/visibility: cluster-local
spec:
  template:
    spec:
      containers:
      - image: ko://github.com/vmware-tanzu/sources-for-knative/samples/tag-new-vms
```

We then deploy with:

```shell
ko apply -f service.yaml
```

### Receiving `VmCreatedEvent`

At this point our `VSphereSource` is dumping events onto the `Broker`, and we
have a `Service` bound and ready to handle events, but we haven't connected the
two.  To connect these two pieces, we are going to create the following trigger:

```yaml
apiVersion: eventing.knative.dev/v1alpha1
kind: Trigger
metadata:
  name: tag-new-vms
spec:
  # We only want to respond to VmCreatedEvent
  filter:
    attributes:
      type: com.vmware.vsphere.VmCreatedEvent

  # Send the event to our service.
  subscriber:
    ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: tag-new-vms
```

Create the trigger with:

```shell
kubectl apply -f service.yaml
```

### Seeing it all in action

If you are using a proper vSphere environment, with a tag named `shrug`, then
you can simply create a new VM and see the new tag applied in the console.


If you are using `vcsim` from the prior sample, then the simplest way to
retrigger its `VmCreatedEvent` is to create the source *last*.  If you already
created it then run:

```shell
kubectl delete -f source.yaml
sleep 30
kubectl apply -f source.yaml
```


If the Service has scaled to zero, you should see it spin up, and if you run
the following see our log output:

```shell
kubectl logs -lserving.knative.dev/service=tag-new-vms -c user-container
2020/04/04 20:21:48 Tagging VM: VirtualMachine:vm-74
```


For some extra fun, modify the [`govc` sample](../govc/README.md) to run the
following command:

```shell
# Change the reference to what you see in the log output above
govc tags.attached.ls -r VirtualMachine:vm-74
```

If you are successful then you should see:

```shell
kubectl logs -lrole=vsphere-job
shrug
```


### Cleanup

```shell
kubectl delete -f trigger.yaml
kubectl delete -f service.yaml
kubectl delete -f binding.yaml
kubectl delete -f source.yaml
```