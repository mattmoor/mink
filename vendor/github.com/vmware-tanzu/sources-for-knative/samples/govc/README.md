## `VSphereBinding` with `govc`

To kick the tires on `VSphereBinding` you don't even need to write code!
In this sample we are going to demonstrate how the `VSphereBinding` sets
up the environment so that you can use the `govc` CLI without any additional
setup.


### Pre-requisites

This sample assumes that you have a vSphere environment set up already
with credentials in a Secret named `vsphere-credentials`.  For the remainder
of the sample we will assume you are within the environment setup for the
[`vcsim` sample](../vcsim/README.md).


### Create the Binding

We are going to use the following binding to authenticate `govc`:

```yaml
apiVersion: sources.tanzu.vmware.com/v1alpha1
kind: VSphereBinding
metadata:
  name: vsphere-jobs
spec:

  # Apply to every Job labeled "role: vsphere-jobs" in
  # this namespace
  subject:
    apiVersion: batch/v1
    kind: Job
    selector:
      matchLabels:
        role: vsphere-jobs

  # The address and credentials for vSphere.
  # If you aren't using the simulator, change this!
  address: https://vcsim.default.svc.cluster.local
  skipTLSVerify: true
  secretRef:
    name: vsphere-credentials
```

Once you have your binding ready, apply it with:

```shell
kubectl apply -f binding.yaml
```

### Script against vSphere with `govc` 

We are going to run the following Job to script some automation using `govc`:

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: vsphere-script
  labels:
    # Make this Job match the binding!
    role: vsphere-job
spec:
  template:
    metadata:
      labels:
        # So it's easier to list the Pods
        role: vsphere-job        
    spec:
      restartPolicy: Never
      containers:
      - name: create-tag
        image: ko://github.com/vmware-tanzu/sources-for-knative/vendor/github.com/vmware/govmomi/govc
        command: ["/bin/bash", "-c"]
        args:
        - |
          # Just invoke govc and it works like magic!
          govc tags.category.create testing
          govc tags.create -c testing shrug
```

This Job creates a tag category called `testing` and a tag named `shrug`.  Run it with:

```shell
ko apply -f job.yaml
```

When the job completes, check its logs with:
```shell
kubectl logs -lrole=vsphere-job
urn:vmomi:InventoryServiceCategory:3c8d271f-0f6d-4af0-b4c2-a612aa6b390a:GLOBAL
urn:vmomi:InventoryServiceTag:5db2c0fa-fe61-42fa-a1fa-90a5f72b8648:GLOBAL
```


### Cleanup

```shell
kubectl delete -f binding.yaml
kubectl delete -f job.yaml
```