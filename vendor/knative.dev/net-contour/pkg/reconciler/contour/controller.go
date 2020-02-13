/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package contour

import (
	"context"

	contourclient "knative.dev/net-contour/pkg/client/injection/client"
	proxyinformer "knative.dev/net-contour/pkg/client/injection/informers/projectcontour/v1/httpproxy"
	endpointsinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/endpoints"
	podinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/pod"
	serviceinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/service"
	"knative.dev/pkg/tracker"
	servingclient "knative.dev/serving/pkg/client/injection/client"
	ingressinformer "knative.dev/serving/pkg/client/injection/informers/networking/v1alpha1/ingress"

	"knative.dev/net-contour/pkg/reconciler/contour/config"
	"knative.dev/serving/pkg/network"
	"knative.dev/serving/pkg/network/status"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"knative.dev/serving/pkg/apis/networking"
	"knative.dev/serving/pkg/apis/networking/v1alpha1"
	"knative.dev/serving/pkg/reconciler"
)

const (
	controllerAgentName = "contour-controller"
)

// NewController returns a new Ingress controller for Project Contour.
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {
	logger := logging.FromContext(ctx)

	endpointsInformer := endpointsinformer.Get(ctx)
	serviceInformer := serviceinformer.Get(ctx)
	ingressInformer := ingressinformer.Get(ctx)
	proxyInformer := proxyinformer.Get(ctx)
	podInformer := podinformer.Get(ctx)

	c := &Reconciler{
		client:          servingclient.Get(ctx),
		contourClient:   contourclient.Get(ctx),
		lister:          ingressInformer.Lister(),
		contourLister:   proxyInformer.Lister(),
		serviceLister:   serviceInformer.Lister(),
		endpointsLister: endpointsInformer.Lister(),
		recorder: record.NewBroadcaster().NewRecorder(
			scheme.Scheme, corev1.EventSource{Component: controllerAgentName}),
	}
	impl := controller.NewImpl(c, logger, "ContourIngresses")

	logger.Info("Setting up event handlers")

	myFilterFunc := reconciler.AnnotationFilterFunc(networking.IngressClassAnnotationKey, ContourIngressClassName, false)
	ingressHandler := cache.FilteringResourceEventHandler{
		FilterFunc: myFilterFunc,
		Handler:    controller.HandleAll(impl.Enqueue),
	}
	ingressInformer.Informer().AddEventHandler(ingressHandler)

	proxyInformer.Informer().AddEventHandler(controller.HandleAll(impl.EnqueueControllerOf))

	statusProber := status.NewProber(
		logger.Named("status-manager"),
		&lister{
			ServiceLister:   serviceInformer.Lister(),
			EndpointsLister: endpointsInformer.Lister(),
		},
		func(ia *v1alpha1.Ingress) { impl.Enqueue(ia) })
	c.statusManager = statusProber
	statusProber.Start(ctx.Done())

	ingressInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		// Cancel probing when an Ingress is deleted
		DeleteFunc: statusProber.CancelIngressProbing,
	})
	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		// Cancel probing when a Pod is deleted
		DeleteFunc: statusProber.CancelPodProbing,
	})

	configsToResync := []interface{}{
		&config.Contour{},
		&network.Config{},
	}
	resyncIngressesOnConfigChange := configmap.TypeFilter(configsToResync...)(func(string, interface{}) {
		impl.FilteredGlobalResync(myFilterFunc, ingressInformer.Informer())
	})
	configStore := config.NewStore(logger.Named("config-store"), resyncIngressesOnConfigChange)
	configStore.WatchConfigs(cmw)
	c.configStore = configStore

	// Set up our tracker to facilitate tracking cross-references to objects we don't own.
	c.tracker = tracker.New(impl.EnqueueKey, controller.GetTrackerLease(ctx))
	serviceInformer.Informer().AddEventHandler(controller.HandleAll(
		// Call the tracker's OnChanged method, but we've seen the objects
		// coming through this path missing TypeMeta, so ensure it is properly
		// populated.
		controller.EnsureTypeMeta(
			c.tracker.OnChanged,
			corev1.SchemeGroupVersion.WithKind("Service"),
		),
	))
	endpointsInformer.Informer().AddEventHandler(controller.HandleAll(
		// Call the tracker's OnChanged method, but we've seen the objects
		// coming through this path missing TypeMeta, so ensure it is properly
		// populated.
		controller.EnsureTypeMeta(
			c.tracker.OnChanged,
			corev1.SchemeGroupVersion.WithKind("Endpoints"),
		),
	))

	return impl
}
