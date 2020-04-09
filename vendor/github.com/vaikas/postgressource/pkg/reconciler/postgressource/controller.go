/*
Copyright 2019 The Knative Authors

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

package postgressource

import (
	"context"

	"github.com/vaikas/postgressource/pkg/apis/sources/v1alpha1"

	"github.com/kelseyhightower/envconfig"
	"k8s.io/client-go/tools/cache"

	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"

	bindingsql "github.com/mattmoor/bindings/pkg/sql"
	postgressourceinformer "github.com/vaikas/postgressource/pkg/client/injection/informers/sources/v1alpha1/postgressource"
	"github.com/vaikas/postgressource/pkg/client/injection/reconciler/sources/v1alpha1/postgressource"
	"github.com/vaikas/postgressource/pkg/reconciler"
	eventingclient "knative.dev/eventing/pkg/client/injection/client"
	sinkbindinginformer "knative.dev/eventing/pkg/client/injection/informers/sources/v1alpha2/sinkbinding"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	deploymentinformer "knative.dev/pkg/client/injection/kube/informers/apps/v1/deployment"
	secretinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/secret"

	_ "github.com/lib/pq"
)

// NewController initializes the controller and is called by the generated code
// Registers event handlers to enqueue events
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {
	deploymentInformer := deploymentinformer.Get(ctx)
	sinkBindingInformer := sinkbindinginformer.Get(ctx)
	postgresSourceInformer := postgressourceinformer.Get(ctx)
	secretInformer := secretinformer.Get(ctx)

	db, err := bindingsql.Open(ctx, "postgres")

	// If running in a MT environment, we might have more than one secret
	// that we have to use, or the secret might not be available when we
	// start, so muster on if we can't get one.
	if err != nil {
		// Check to see if the secret is even mounted
		if _, err = bindingsql.ReadKey("connectionstr"); err != nil {
			logging.FromContext(ctx).Warnf("Looks like the secret for the db is not mounted: %v", err)
		}
		logging.FromContext(ctx).Warnf("Failed to start db: %v", err)
	}

	r := &Reconciler{
		dr:           &reconciler.DeploymentReconciler{KubeClientSet: kubeclient.Get(ctx)},
		sbr:          &reconciler.SinkBindingReconciler{EventingClientSet: eventingclient.Get(ctx)},
		db:           db,
		secretLister: secretInformer.Lister(),
	}
	impl := postgressource.NewImpl(ctx, r)
	if err := envconfig.Process("", r); err != nil {
		logging.FromContext(ctx).Panicf("required environment variable is not defined: %v", err)
	}

	logging.FromContext(ctx).Info("Setting up event handlers")

	postgresSourceInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	deploymentInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterGroupKind(v1alpha1.Kind("PostgresSource")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	sinkBindingInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterGroupKind(v1alpha1.Kind("PostgresSource")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	secretInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterGroupKind(v1alpha1.Kind("PostgresSource")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	return impl
}
