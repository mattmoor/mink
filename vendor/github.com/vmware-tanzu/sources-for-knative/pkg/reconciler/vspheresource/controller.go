/*
Copyright 2020 VMware, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package vspheresource

import (
	"context"

	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"

	"github.com/kelseyhightower/envconfig"
	"github.com/vmware-tanzu/sources-for-knative/pkg/apis/sources/v1alpha1"
	"github.com/vmware-tanzu/sources-for-knative/pkg/client/injection/client"
	vspherebindinginformer "github.com/vmware-tanzu/sources-for-knative/pkg/client/injection/informers/sources/v1alpha1/vspherebinding"
	vsphereinformer "github.com/vmware-tanzu/sources-for-knative/pkg/client/injection/informers/sources/v1alpha1/vspheresource"
	vspherereconciler "github.com/vmware-tanzu/sources-for-knative/pkg/client/injection/reconciler/sources/v1alpha1/vspheresource"
	eventingclient "knative.dev/eventing/pkg/client/injection/client"
	sinkbindinginformer "knative.dev/eventing/pkg/client/injection/informers/sources/v1alpha1/sinkbinding"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	deploymentinformer "knative.dev/pkg/client/injection/kube/informers/apps/v1/deployment"
	cminformer "knative.dev/pkg/client/injection/kube/informers/core/v1/configmap"
	sainformer "knative.dev/pkg/client/injection/kube/informers/core/v1/serviceaccount"
	rbacinformer "knative.dev/pkg/client/injection/kube/informers/rbac/v1/rolebinding"
)

type envConfig struct {
	VSphereAdapter string `envconfig:"VSPHERE_ADAPTER" required:"true"`
}

// NewController creates a Reconciler and returns the result of NewImpl.
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {
	logger := logging.FromContext(ctx)

	vsphereInformer := vsphereinformer.Get(ctx)
	deploymentInformer := deploymentinformer.Get(ctx)
	sinkbindingInformer := sinkbindinginformer.Get(ctx)
	rbacInformer := rbacinformer.Get(ctx)
	cmInformer := cminformer.Get(ctx)
	vspherebindingInformer := vspherebindinginformer.Get(ctx)
	saInformer := sainformer.Get(ctx)

	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		logger.Fatalf("Unable to read environment config: %v", err)
	}

	r := &Reconciler{
		adapterImage:         env.VSphereAdapter,
		kubeclient:           kubeclient.Get(ctx),
		eventingclient:       eventingclient.Get(ctx),
		client:               client.Get(ctx),
		deploymentLister:     deploymentInformer.Lister(),
		vspherebindingLister: vspherebindingInformer.Lister(),
		sinkbindingLister:    sinkbindingInformer.Lister(),
		cmLister:             cmInformer.Lister(),
		rbacLister:           rbacInformer.Lister(),
		saLister:             saInformer.Lister(),
	}
	impl := vspherereconciler.NewImpl(ctx, r)

	logger.Info("Setting up event handlers.")

	vsphereInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	deploymentInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterGroupKind(v1alpha1.Kind("VSphereSource")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	saInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterGroupKind(v1alpha1.Kind("VSphereSource")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	rbacInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterGroupKind(v1alpha1.Kind("VSphereSource")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	// Don't trigger off of CM updates because we don't care about the content
	// and it is high churn.

	sinkbindingInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterGroupKind(v1alpha1.Kind("VSphereSource")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	vspherebindingInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterGroupKind(v1alpha1.Kind("VSphereSource")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	return impl
}
