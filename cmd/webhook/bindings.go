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

	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/webhook/psbinding"

	"knative.dev/eventing/pkg/reconciler/sinkbinding"
)

func NewSinkBindingWebhook(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
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
	)
}
