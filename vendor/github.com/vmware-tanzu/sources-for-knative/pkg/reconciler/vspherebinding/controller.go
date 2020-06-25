/*
Copyright 2020 VMware, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package vspherebinding

import (
	"context"

	vsbinformer "github.com/vmware-tanzu/sources-for-knative/pkg/client/injection/informers/sources/v1alpha1/vspherebinding"
	"knative.dev/pkg/client/injection/ducks/duck/v1/podspecable"
	"knative.dev/pkg/client/injection/kube/informers/core/v1/namespace"
	"knative.dev/pkg/reconciler"

	"github.com/vmware-tanzu/sources-for-knative/pkg/apis/sources/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"knative.dev/pkg/apis/duck"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection/clients/dynamicclient"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/tracker"
	"knative.dev/pkg/webhook/psbinding"
)

const (
	controllerAgentName = "vspherebinding-controller"
)

// NewController returns a new VSphereBinding reconciler.
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {
	logger := logging.FromContext(ctx)

	vsbInformer := vsbinformer.Get(ctx)
	dc := dynamicclient.Get(ctx)
	psInformerFactory := podspecable.Get(ctx)
	namespaceInformer := namespace.Get(ctx)

	c := &psbinding.BaseReconciler{
		LeaderAwareFuncs: reconciler.LeaderAwareFuncs{
			PromoteFunc: func(bkt reconciler.Bucket, enq func(reconciler.Bucket, types.NamespacedName)) error {
				all, err := vsbInformer.Lister().List(labels.Everything())
				if err != nil {
					return err
				}
				for _, elt := range all {
					enq(bkt, types.NamespacedName{
						Namespace: elt.GetNamespace(),
						Name:      elt.GetName(),
					})
				}
				return nil
			},
		},
		GVR: v1alpha1.SchemeGroupVersion.WithResource("vspherebindings"),
		Get: func(namespace string, name string) (psbinding.Bindable, error) {
			return vsbInformer.Lister().VSphereBindings(namespace).Get(name)
		},
		DynamicClient: dc,
		Recorder: record.NewBroadcaster().NewRecorder(
			scheme.Scheme, corev1.EventSource{Component: controllerAgentName}),
		NamespaceLister: namespaceInformer.Lister(),
	}
	impl := controller.NewImpl(c, logger, "VSphereBindings")

	logger.Info("Setting up event handlers")

	vsbInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	c.Tracker = tracker.New(impl.EnqueueKey, controller.GetTrackerLease(ctx))
	c.Factory = &duck.CachedInformerFactory{
		Delegate: &duck.EnqueueInformerFactory{
			Delegate:     psInformerFactory,
			EventHandler: controller.HandleAll(c.Tracker.OnChanged),
		},
	}

	return impl
}

func ListAll(ctx context.Context, handler cache.ResourceEventHandler) psbinding.ListAll {
	fbInformer := vsbinformer.Get(ctx)

	// Whenever a VSphereBinding changes our webhook programming might change.
	fbInformer.Informer().AddEventHandler(handler)

	return func() ([]psbinding.Bindable, error) {
		l, err := fbInformer.Lister().List(labels.Everything())
		if err != nil {
			return nil, err
		}
		bl := make([]psbinding.Bindable, 0, len(l))
		for _, elt := range l {
			bl = append(bl, elt)
		}
		return bl, nil
	}

}
