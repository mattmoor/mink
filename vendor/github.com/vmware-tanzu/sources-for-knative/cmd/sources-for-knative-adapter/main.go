/*
Copyright 2020 VMware, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"

	// Uncomment if you want to run locally against remote GKE cluster.
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"k8s.io/client-go/kubernetes"
	"knative.dev/eventing/pkg/adapter/v2"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	"knative.dev/pkg/injection/sharedmain"
	"knative.dev/pkg/signals"

	"github.com/vmware-tanzu/sources-for-knative/pkg/vsphere"
)

func main() {
	ctx := signals.NewContext()
	kc := kubernetes.NewForConfigOrDie(sharedmain.ParseAndGetConfigOrDie())
	ctx = context.WithValue(ctx, kubeclient.Key{}, kc)
	adapter.MainWithContext(ctx, "vspheresource", vsphere.NewEnvConfig, vsphere.NewAdapter)
}
