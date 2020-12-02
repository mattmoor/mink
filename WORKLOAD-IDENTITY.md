# Using `mink` with workload identity systems

Several cloud vendors support a mechanism for granting IAM capabilities and roles to workloads
at a Pod granularity, and block access to the Node-level equivalent.  This can lead to problems
with both Tekton and Knative controllers, which want to be able to access container image metadata.

This walks through how to configure vendor workload identity mechanisms with `mink`.

# [GKE Workload Identity](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity)

To allow the Tekton and Knative Serving controllers to access container images, run the following:

```
# Create mink-controller IAM SA in the current project
gcloud iam service-accounts create mink-controller

# Grant registry read access to this IAM SA.
gcloud projects add-iam-policy-binding $(gcloud config get-value core/project) \
    --member=serviceAccount:mink-controller@$(gcloud config get-value core/project).iam.gserviceaccount.com \
    --role=roles/storage.objectViewer

# Allow the mink-system/controller SA to impersonate the above SA
gcloud iam service-accounts add-iam-policy-binding \
    --role roles/iam.workloadIdentityUser \
    --member "serviceAccount:$(gcloud config get-value core/project).svc.id.goog[mink-system/controller]" \
    mink-controller@$(gcloud config get-value core/project).iam.gserviceaccount.com

# Tell the workload identity the name of the service account to expose to this workload.
kubectl annotate serviceaccount \
    --namespace mink-system \
    controller \
    iam.gke.io/gcp-service-account=mink-controller@$(gcloud config get-value core/project).iam.gserviceaccount.com

```

# [EKS Fine-Grained IAM](https://aws.amazon.com/blogs/opensource/introducing-fine-grained-iam-roles-service-accounts/)

To allow the Tekton and Knative Serving controllers to access container images, run the following:

```
# Cluster must have oidc enabled:
eksctl utils associate-iam-oidc-provider --name YOUR_CLUSTER_NAME --approve

# Create service account with:
eksctl create iamserviceaccount \
    --name controller \
    --namespace mink-system \
    --cluster YOUR_CLUSTER_NAME \
    --attach-policy-arn arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly \
    --approve
```

