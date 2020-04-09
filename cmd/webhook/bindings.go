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

package main

import (
	"context"
	"fmt"

	"github.com/vmware-tanzu/sources-for-knative/pkg/reconciler/vspherebinding"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/eventing/pkg/reconciler/sinkbinding"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection"
	"knative.dev/pkg/webhook/psbinding"
)

func NewSinkBindingWebhook(opts ...psbinding.ReconcilerOption) injection.ControllerConstructor {
	return func(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
		sbresolver := sinkbinding.WithContextFactory(ctx, func(types.NamespacedName) {})

		return psbinding.NewAdmissionController(ctx,
			// Name of the resource webhook.
			"sinkbindings.webhook.mink.knative.dev",

			// The path on which to serve the webhook.
			"/sinkbindings",

			// How to get all the Bindables for configuring the mutating webhook.
			sinkbinding.ListAll,

			// How to setup the context prior to invoking Do/Undo.
			sbresolver,

			// Pass through options from our caller.
			opts...,
		)
	}
}

func NewVSphereBindingWebhook(opts ...psbinding.ReconcilerOption) injection.ControllerConstructor {
	return func(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
		return psbinding.NewAdmissionController(ctx,
			// Name of the resource webhook.
			"vspherebindings.webhook.mink.knative.dev",

			// The path on which to serve the webhook.
			"/vspherebindings",

			// How to get all the Bindables for configuring the mutating webhook.
			vspherebinding.ListAll,

			// A function that infuses the context passed to Validate/SetDefaults with custom metadata.
			func(ctx context.Context, _ psbinding.Bindable) (context.Context, error) {
				// Here is where you would infuse the context with state
				// (e.g. attach a store with configmap data)
				return ctx, nil
			},
			opts...,
		)
	}
}

func NewBindingWebhook(resource string, gla psbinding.GetListAll, wc psbinding.BindableContext) injection.ControllerConstructor {
	return func(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
		return psbinding.NewAdmissionController(ctx,
			// Name of the resource webhook.
			fmt.Sprintf("%s.webhook.mink.knative.dev", resource),

			// The path on which to serve the webhook.
			fmt.Sprintf("/%s", resource),

			// How to get all the Bindables for configuring the mutating webhook.
			gla,

			// How to setup the context prior to invoking Do/Undo.
			wc,
		)
	}
}
