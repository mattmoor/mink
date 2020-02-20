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

package certificate

import (
	context "context"
	"time"

	"github.com/mattmoor/http01-solver/pkg/challenger"
	"github.com/mattmoor/http01-solver/pkg/ordermanager"
	"k8s.io/api/discovery/v1alpha1"
	"k8s.io/client-go/tools/cache"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	endpointsinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/endpoints"
	secretinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/secret"
	serviceinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/service"
	configmap "knative.dev/pkg/configmap"
	controller "knative.dev/pkg/controller"
	logging "knative.dev/pkg/logging"
	"knative.dev/pkg/reconciler"
	"knative.dev/serving/pkg/apis/networking"
	certificate "knative.dev/serving/pkg/client/injection/informers/networking/v1alpha1/certificate"
	v1alpha1certificate "knative.dev/serving/pkg/client/injection/reconciler/networking/v1alpha1/certificate"
)

const CertificateClassName = "mattmoor-http01.certificate.networking.knative.dev"

// NewController creates a Reconciler for Certificate and returns the result of NewImpl.
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
	chlr challenger.Interface,
) *controller.Impl {
	logger := logging.FromContext(ctx)

	certificateInformer := certificate.Get(ctx)
	secretInformer := secretinformer.Get(ctx)
	serviceInformer := serviceinformer.Get(ctx)
	endpointsInformer := endpointsinformer.Get(ctx)

	r := &Reconciler{
		kubeClient:      kubeclient.Get(ctx),
		secretLister:    secretInformer.Lister(),
		serviceLister:   serviceInformer.Lister(),
		endpointsLister: endpointsInformer.Lister(),
	}
	impl := v1alpha1certificate.NewImpl(ctx, r, CertificateClassName)

	logger.Info("Setting up event handlers.")

	classFilterFunc := reconciler.AnnotationFilterFunc(
		networking.CertificateClassAnnotationKey, CertificateClassName, true)
	certHandler := cache.FilteringResourceEventHandler{
		FilterFunc: classFilterFunc,
		Handler:    controller.HandleAll(impl.Enqueue),
	}
	certificateInformer.Informer().AddEventHandler(certHandler)

	secretInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterGroupKind(v1alpha1.Kind("Certificate")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})
	serviceInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterGroupKind(v1alpha1.Kind("Certificate")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})
	endpointsInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterGroupKind(v1alpha1.Kind("Certificate")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	// Don't let the OrderManager hang on client calls.
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	om, err := ordermanager.New(ctx, impl.Enqueue, chlr)
	if err != nil {
		logging.FromContext(ctx).Fatalf("Error creating OrderManager: %v", err)
	}
	r.orderManager = om

	// TODO: add additional informer event handlers here.

	return impl
}
