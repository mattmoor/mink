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

	tkndefaultconfig "github.com/tektoncd/pipeline/pkg/apis/config"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	knedefaultconfig "knative.dev/eventing/pkg/apis/config"
	"knative.dev/eventing/pkg/apis/eventing"
	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
	eventingv1beta1 "knative.dev/eventing/pkg/apis/eventing/v1beta1"
	"knative.dev/eventing/pkg/apis/flows"
	flowsv1 "knative.dev/eventing/pkg/apis/flows/v1"
	flowsv1beta1 "knative.dev/eventing/pkg/apis/flows/v1beta1"
	"knative.dev/eventing/pkg/apis/messaging"
	channeldefaultconfig "knative.dev/eventing/pkg/apis/messaging/config"
	messagingv1 "knative.dev/eventing/pkg/apis/messaging/v1"
	messagingv1beta1 "knative.dev/eventing/pkg/apis/messaging/v1beta1"
	"knative.dev/eventing/pkg/apis/sources"
	sourcesv1 "knative.dev/eventing/pkg/apis/sources/v1"
	sourcesv1alpha1 "knative.dev/eventing/pkg/apis/sources/v1alpha1"
	sourcesv1alpha2 "knative.dev/eventing/pkg/apis/sources/v1alpha2"
	sourcesv1beta1 "knative.dev/eventing/pkg/apis/sources/v1beta1"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/webhook/resourcesemantics/conversion"
	knsdefaultconfig "knative.dev/serving/pkg/apis/config"
)

func newConversionController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	// Decorate contexts with the current state of the config.
	knsstore := knsdefaultconfig.NewStore(logging.FromContext(ctx).Named("config-store"))
	knsstore.WatchConfigs(cmw)

	knestore := knedefaultconfig.NewStore(logging.FromContext(ctx).Named("config-store"))
	knestore.WatchConfigs(cmw)

	tknstore := tkndefaultconfig.NewStore(logging.FromContext(ctx).Named("config-store"))
	tknstore.WatchConfigs(cmw)

	channelStore := channeldefaultconfig.NewStore(logging.FromContext(ctx).Named("channel-config-store"))
	channelStore.WatchConfigs(cmw)

	var (
		eventingv1beta1Version  = eventingv1beta1.SchemeGroupVersion.Version
		eventingv1Version       = eventingv1.SchemeGroupVersion.Version
		messagingv1beta1Version = messagingv1beta1.SchemeGroupVersion.Version
		messagingv1Version      = messagingv1.SchemeGroupVersion.Version
		flowsv1beta1Version     = flowsv1beta1.SchemeGroupVersion.Version
		flowsv1Version          = flowsv1.SchemeGroupVersion.Version
		sourcesv1alpha1Version  = sourcesv1alpha1.SchemeGroupVersion.Version
		sourcesv1alpha2Version  = sourcesv1alpha2.SchemeGroupVersion.Version
		sourcesv1beta1Version   = sourcesv1beta1.SchemeGroupVersion.Version
		sourcesv1Version        = sourcesv1.SchemeGroupVersion.Version

		tektonv1alpha1Version = tektonv1alpha1.SchemeGroupVersion.Version
		tektonv1beta1Version  = tektonv1beta1.SchemeGroupVersion.Version
	)

	return conversion.NewConversionController(ctx,
		// The path on which to serve the webhook
		"/resource-conversion",

		// Specify the types of custom resource definitions that should be converted
		map[schema.GroupKind]conversion.GroupKindConversion{
			// Eventing
			eventingv1.Kind("Trigger"): {
				DefinitionName: eventing.TriggersResource.String(),
				HubVersion:     eventingv1beta1Version,
				Zygotes: map[string]conversion.ConvertibleObject{
					eventingv1beta1Version: &eventingv1beta1.Trigger{},
					eventingv1Version:      &eventingv1.Trigger{},
				},
			},
			eventingv1.Kind("Broker"): {
				DefinitionName: eventing.BrokersResource.String(),
				HubVersion:     eventingv1beta1Version,
				Zygotes: map[string]conversion.ConvertibleObject{
					eventingv1beta1Version: &eventingv1beta1.Broker{},
					eventingv1Version:      &eventingv1.Broker{},
				},
			},

			// Messaging
			messagingv1.Kind("Channel"): {
				DefinitionName: messaging.ChannelsResource.String(),
				HubVersion:     messagingv1beta1Version,
				Zygotes: map[string]conversion.ConvertibleObject{
					messagingv1beta1Version: &messagingv1beta1.Channel{},
					messagingv1Version:      &messagingv1.Channel{},
				},
			},
			messagingv1.Kind("Subscription"): {
				DefinitionName: messaging.SubscriptionsResource.String(),
				HubVersion:     messagingv1beta1Version,
				Zygotes: map[string]conversion.ConvertibleObject{
					messagingv1beta1Version: &messagingv1beta1.Subscription{},
					messagingv1Version:      &messagingv1.Subscription{},
				},
			},
			// TODO(mattmoor): Can we split this out?
			messagingv1.Kind("InMemoryChannel"): {
				DefinitionName: messaging.InMemoryChannelsResource.String(),
				HubVersion:     messagingv1beta1Version,
				Zygotes: map[string]conversion.ConvertibleObject{
					messagingv1beta1Version: &messagingv1beta1.InMemoryChannel{},
					messagingv1Version:      &messagingv1.InMemoryChannel{},
				},
			},

			// flows
			flowsv1.Kind("Sequence"): {
				DefinitionName: flows.SequenceResource.String(),
				HubVersion:     flowsv1beta1Version,
				Zygotes: map[string]conversion.ConvertibleObject{
					flowsv1beta1Version: &flowsv1beta1.Sequence{},
					flowsv1Version:      &flowsv1.Sequence{},
				},
			},
			flowsv1.Kind("Parallel"): {
				DefinitionName: flows.ParallelResource.String(),
				HubVersion:     flowsv1beta1Version,
				Zygotes: map[string]conversion.ConvertibleObject{
					flowsv1beta1Version: &flowsv1beta1.Parallel{},
					flowsv1Version:      &flowsv1.Parallel{},
				},
			},

			// Sources
			sourcesv1beta1.Kind("ApiServerSource"): {
				DefinitionName: sources.ApiServerSourceResource.String(),
				HubVersion:     sourcesv1alpha1Version,
				Zygotes: map[string]conversion.ConvertibleObject{
					sourcesv1alpha1Version: &sourcesv1alpha1.ApiServerSource{},
					sourcesv1alpha2Version: &sourcesv1alpha2.ApiServerSource{},
					sourcesv1beta1Version:  &sourcesv1beta1.ApiServerSource{},
					sourcesv1Version:       &sourcesv1.ApiServerSource{},
				},
			},
			sourcesv1beta1.Kind("PingSource"): {
				DefinitionName: sources.PingSourceResource.String(),
				HubVersion:     sourcesv1alpha2Version,
				Zygotes: map[string]conversion.ConvertibleObject{
					sourcesv1alpha2Version: &sourcesv1alpha2.PingSource{},
					sourcesv1beta1Version:  &sourcesv1beta1.PingSource{},
					// sourcesv1Version:       &sourcesv1.PingSource{},
				},
			},
			sourcesv1beta1.Kind("SinkBinding"): {
				DefinitionName: sources.SinkBindingResource.String(),
				HubVersion:     sourcesv1alpha1Version,
				Zygotes: map[string]conversion.ConvertibleObject{
					sourcesv1alpha1Version: &sourcesv1alpha1.SinkBinding{},
					sourcesv1alpha2Version: &sourcesv1alpha2.SinkBinding{},
					sourcesv1beta1Version:  &sourcesv1beta1.SinkBinding{},
					sourcesv1Version:       &sourcesv1.SinkBinding{},
				},
			},
			sourcesv1beta1.Kind("ContainerSource"): {
				DefinitionName: sources.ContainerSourceResource.String(),
				HubVersion:     sourcesv1alpha2Version,
				Zygotes: map[string]conversion.ConvertibleObject{
					sourcesv1alpha2Version: &sourcesv1alpha2.ContainerSource{},
					sourcesv1beta1Version:  &sourcesv1beta1.ContainerSource{},
					sourcesv1Version:       &sourcesv1.ContainerSource{},
				},
			},

			// Tekton
			tektonv1beta1.Kind("Task"): {
				DefinitionName: pipeline.TaskResource.String(),
				HubVersion:     tektonv1alpha1Version,
				Zygotes: map[string]conversion.ConvertibleObject{
					tektonv1alpha1Version: &tektonv1alpha1.Task{},
					tektonv1beta1Version:  &tektonv1beta1.Task{},
				},
			},
			tektonv1beta1.Kind("ClusterTask"): {
				DefinitionName: pipeline.ClusterTaskResource.String(),
				HubVersion:     tektonv1alpha1Version,
				Zygotes: map[string]conversion.ConvertibleObject{
					tektonv1alpha1Version: &tektonv1alpha1.ClusterTask{},
					tektonv1beta1Version:  &tektonv1beta1.ClusterTask{},
				},
			},
			tektonv1beta1.Kind("TaskRun"): {
				DefinitionName: pipeline.TaskRunResource.String(),
				HubVersion:     tektonv1alpha1Version,
				Zygotes: map[string]conversion.ConvertibleObject{
					tektonv1alpha1Version: &tektonv1alpha1.TaskRun{},
					tektonv1beta1Version:  &tektonv1beta1.TaskRun{},
				},
			},
			tektonv1beta1.Kind("Pipeline"): {
				DefinitionName: pipeline.PipelineResource.String(),
				HubVersion:     tektonv1alpha1Version,
				Zygotes: map[string]conversion.ConvertibleObject{
					tektonv1alpha1Version: &tektonv1alpha1.Pipeline{},
					tektonv1beta1Version:  &tektonv1beta1.Pipeline{},
				},
			},
			tektonv1beta1.Kind("PipelineRun"): {
				DefinitionName: pipeline.PipelineRunResource.String(),
				HubVersion:     tektonv1alpha1Version,
				Zygotes: map[string]conversion.ConvertibleObject{
					tektonv1alpha1Version: &tektonv1alpha1.PipelineRun{},
					tektonv1beta1Version:  &tektonv1beta1.PipelineRun{},
				},
			},
		},

		// A function that infuses the context passed to ConvertUp/ConvertDown/SetDefaults with
		// custom metadata.
		func(ctx context.Context) context.Context {
			return channelStore.ToContext(tknstore.ToContext(knestore.ToContext(knsstore.ToContext(ctx))))
		},
	)
}
