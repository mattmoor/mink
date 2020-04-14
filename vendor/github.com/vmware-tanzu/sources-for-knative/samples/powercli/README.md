## `VSphereBinding` with `PowerCLI`

This sample builds on our [previous sample](../govc/README.md) to show how to
use `VSphereBinding` to authenticate `PowerCLI` (a more familiar tool to most
admins).

We are going to make use of the `vmware/powerclicore` container image in this
sample.

### Pre-requisites

**Unlike the previous examples**, this example does not work with `vcsim`, so
you will need a real vSphere environment set up, with credentials in a Secret
named `vsphere-credentials`.

### Create the Binding

We are going to use the following binding to authenticate `PowerCLI`:

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

  address: REPLACE_ME    <-- This must point to a real vSphere environment.
  skipTLSVerify: true
  secretRef:
    name: vsphere-credentials
```

Once you have your binding ready, apply it with:

```shell
kubectl apply -f binding.yaml
```

### Script against vSphere with `PowerCLI`

We are going to run the following Job to script some automation using
`PowerCLI`:

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
        - name: dump-events
          image: docker.io/vmware/powerclicore
          command: ["pwsh", "-Command"]
          args:
            - |
              # Log into the VI Server
              Set-PowerCLIConfiguration -InvalidCertificateAction Ignore -Confirm:$false | Out-Null
              Connect-VIServer -Server ([System.Uri]$env:GOVC_URL).Host -User $env:GOVC_USERNAME -Password $env:GOVC_PASSWORD

              # Get Events and write them out.
              Get-VIEvent | Write-Host
```

This Job authenticates `PowerCLI` using our injected credentials, and dumps the
event logs. You can run it with:

```shell
kubectl apply -f job.yaml
```

When the job completes, check its logs with:

```shell
kubectl logs -lrole=vsphere-job
```

You should see the event logs for your environment!

### Cleanup

```shell
kubectl delete -f binding.yaml
kubectl delete -f job.yaml
```
