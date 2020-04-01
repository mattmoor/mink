# Tag Created VMs

This sample shows how to do several things:
1. Listen to VM Created Events,
2. Bind vSphere credentials into our Service, and
3. Combine the above to tag newly created VMs with a precreated label.

## Setup

This sample assumes that you have already:
1. Installed Knative Serving/Eventing,
2. Stood up `ko apply -f vcsim.yaml` to simulate a vSphere endpoint
   depositing events on the default Broker,
3. Created our tag with `ko create -f tags.yaml`.


## Triggering on VM Creations

To start our tour, let's first look at how we consume VM creation events (a subset of
what our setup is putting on the Broker).
```yaml
apiVersion: eventing.knative.dev/v1alpha1
kind: Trigger
metadata:
  name: to-sample
spec:
  filter:
    attributes:
      type: com.vmware.vsphere.VmCreatedEvent
  subscriber:
    ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: sample
```

This `Trigger` resource sets up a standing query on the default `Broker` for any Cloud Events
with type `com.vmware.vsphere.VmCreatedEvent`, and forwards all that it finds on to our sample
`Service`.


## Binding vSphere credentials into our Service.

To allow our Service to call back into the vSphere API we bind credential data into
our `Service` with:
```yaml
apiVersion: sources.knative.dev/v1alpha1
kind: VSphereBinding
metadata:
  name: sample-binding
spec:
  # Our Service is going to receive the credential data.
  subject:
    apiVersion: serving.knative.dev/v1
    kind: Service
    name: sample

  # What vSphere endpoint to connect to, and the credentials to use.
  address: https://vcsim.default.svc.cluster.local
  skipTLSVerify: true
  secretRef:
    name: vsphere-credentials
```

In our Go code we can use this to get an authenticated client by writing:
```go

import "github.com/mattmoor/vmware-sources/pkg/vsphere"

...

client, err := vsphere.NewREST(ctx)
if err != nil {
	log.Fatal(err.Error())
}
```

This sets up a `rest.Client`, but we also expose `vsphere.New` for getting a `govmomi.Client`.


## Tagging newly created VMs

Now let's put it all together:
```go
package main

import (
	"context"
	"log"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/mattmoor/vmware-sources/pkg/vsphere"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/vim25/types"
)

type receiver struct {
	manager *tags.Manager
}

func main() {
	ctx := context.Background()

	ceclient, err := cloudevents.NewDefaultClient()
	if err != nil {
		log.Fatal(err.Error())
	}

	// Instantiate a client for interacting with the vSphere APIs.
	client, err := vsphere.NewREST(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}
	r := &receiver{manager: tags.NewManager(client)}

	if err := ceclient.StartReceiver(ctx, r.handle); err != nil {
		log.Fatal(err)
	}
}

func (r *receiver) handle(ctx context.Context, event cloudevents.Event) error {
	// Parse the VmCreatedEvent payload
	req := &types.VmCreatedEvent{}
	if err := event.DataAs(&req); err != nil {
		return err
	}
	// Attach the "shrug" tag to the ManagedObjectReference for the
	// Vm embedded in our event payload.
	return r.manager.AttachTag(ctx, "shrug", req.Vm.Vm)
}

```
