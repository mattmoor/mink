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

// Code generated by informer-gen. DO NOT EDIT.

package externalversions

import (
	"fmt"

	schema "k8s.io/apimachinery/pkg/runtime/schema"
	cache "k8s.io/client-go/tools/cache"
	v1alpha1 "knative.dev/eventing/pkg/apis/configs/v1alpha1"
	eventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	v1beta1 "knative.dev/eventing/pkg/apis/eventing/v1beta1"
	flowsv1alpha1 "knative.dev/eventing/pkg/apis/flows/v1alpha1"
	flowsv1beta1 "knative.dev/eventing/pkg/apis/flows/v1beta1"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	messagingv1beta1 "knative.dev/eventing/pkg/apis/messaging/v1beta1"
	sourcesv1alpha1 "knative.dev/eventing/pkg/apis/sources/v1alpha1"
	v1alpha2 "knative.dev/eventing/pkg/apis/sources/v1alpha2"
)

// GenericInformer is type of SharedIndexInformer which will locate and delegate to other
// sharedInformers based on type
type GenericInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() cache.GenericLister
}

type genericInformer struct {
	informer cache.SharedIndexInformer
	resource schema.GroupResource
}

// Informer returns the SharedIndexInformer.
func (f *genericInformer) Informer() cache.SharedIndexInformer {
	return f.informer
}

// Lister returns the GenericLister.
func (f *genericInformer) Lister() cache.GenericLister {
	return cache.NewGenericLister(f.Informer().GetIndexer(), f.resource)
}

// ForResource gives generic access to a shared informer of the matching type
// TODO extend this to unknown resources with a client pool
func (f *sharedInformerFactory) ForResource(resource schema.GroupVersionResource) (GenericInformer, error) {
	switch resource {
	// Group=configs.internal.knative.dev, Version=v1alpha1
	case v1alpha1.SchemeGroupVersion.WithResource("configmappropagations"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Configs().V1alpha1().ConfigMapPropagations().Informer()}, nil

		// Group=eventing.knative.dev, Version=v1alpha1
	case eventingv1alpha1.SchemeGroupVersion.WithResource("brokers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Eventing().V1alpha1().Brokers().Informer()}, nil
	case eventingv1alpha1.SchemeGroupVersion.WithResource("eventtypes"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Eventing().V1alpha1().EventTypes().Informer()}, nil
	case eventingv1alpha1.SchemeGroupVersion.WithResource("triggers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Eventing().V1alpha1().Triggers().Informer()}, nil

		// Group=eventing.knative.dev, Version=v1beta1
	case v1beta1.SchemeGroupVersion.WithResource("brokers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Eventing().V1beta1().Brokers().Informer()}, nil
	case v1beta1.SchemeGroupVersion.WithResource("eventtypes"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Eventing().V1beta1().EventTypes().Informer()}, nil
	case v1beta1.SchemeGroupVersion.WithResource("triggers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Eventing().V1beta1().Triggers().Informer()}, nil

		// Group=flows.knative.dev, Version=v1alpha1
	case flowsv1alpha1.SchemeGroupVersion.WithResource("parallels"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Flows().V1alpha1().Parallels().Informer()}, nil
	case flowsv1alpha1.SchemeGroupVersion.WithResource("sequences"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Flows().V1alpha1().Sequences().Informer()}, nil

		// Group=flows.knative.dev, Version=v1beta1
	case flowsv1beta1.SchemeGroupVersion.WithResource("parallels"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Flows().V1beta1().Parallels().Informer()}, nil
	case flowsv1beta1.SchemeGroupVersion.WithResource("sequences"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Flows().V1beta1().Sequences().Informer()}, nil

		// Group=messaging.knative.dev, Version=v1alpha1
	case messagingv1alpha1.SchemeGroupVersion.WithResource("channels"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Messaging().V1alpha1().Channels().Informer()}, nil
	case messagingv1alpha1.SchemeGroupVersion.WithResource("inmemorychannels"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Messaging().V1alpha1().InMemoryChannels().Informer()}, nil
	case messagingv1alpha1.SchemeGroupVersion.WithResource("subscriptions"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Messaging().V1alpha1().Subscriptions().Informer()}, nil

		// Group=messaging.knative.dev, Version=v1beta1
	case messagingv1beta1.SchemeGroupVersion.WithResource("channels"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Messaging().V1beta1().Channels().Informer()}, nil
	case messagingv1beta1.SchemeGroupVersion.WithResource("inmemorychannels"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Messaging().V1beta1().InMemoryChannels().Informer()}, nil
	case messagingv1beta1.SchemeGroupVersion.WithResource("subscriptions"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Messaging().V1beta1().Subscriptions().Informer()}, nil

		// Group=sources.knative.dev, Version=v1alpha1
	case sourcesv1alpha1.SchemeGroupVersion.WithResource("apiserversources"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Sources().V1alpha1().ApiServerSources().Informer()}, nil
	case sourcesv1alpha1.SchemeGroupVersion.WithResource("pingsources"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Sources().V1alpha1().PingSources().Informer()}, nil
	case sourcesv1alpha1.SchemeGroupVersion.WithResource("sinkbindings"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Sources().V1alpha1().SinkBindings().Informer()}, nil

		// Group=sources.knative.dev, Version=v1alpha2
	case v1alpha2.SchemeGroupVersion.WithResource("apiserversources"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Sources().V1alpha2().ApiServerSources().Informer()}, nil
	case v1alpha2.SchemeGroupVersion.WithResource("pingsources"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Sources().V1alpha2().PingSources().Informer()}, nil
	case v1alpha2.SchemeGroupVersion.WithResource("sinkbindings"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Sources().V1alpha2().SinkBindings().Informer()}, nil

	}

	return nil, fmt.Errorf("no informer found for %v", resource)
}
