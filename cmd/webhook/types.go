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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/webhook/resourcesemantics"

	tknv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	tknv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
	eventingv1beta1 "knative.dev/eventing/pkg/apis/eventing/v1beta1"
	flowsv1 "knative.dev/eventing/pkg/apis/flows/v1"
	messagingv1 "knative.dev/eventing/pkg/apis/messaging/v1"
	sourcesv1 "knative.dev/eventing/pkg/apis/sources/v1"
	sourcesv1beta2 "knative.dev/eventing/pkg/apis/sources/v1beta2"
	net "knative.dev/networking/pkg/apis/networking/v1alpha1"
	autoscalingv1alpha1 "knative.dev/serving/pkg/apis/autoscaling/v1alpha1"
	v1 "knative.dev/serving/pkg/apis/serving/v1"
	"knative.dev/serving/pkg/apis/serving/v1alpha1"
)

var ourTypes = map[schema.GroupVersionKind]resourcesemantics.GenericCRD{
	v1alpha1.SchemeGroupVersion.WithKind("DomainMapping"): &v1alpha1.DomainMapping{},
	v1.SchemeGroupVersion.WithKind("Revision"):            &v1.Revision{},
	v1.SchemeGroupVersion.WithKind("Configuration"):       &v1.Configuration{},
	v1.SchemeGroupVersion.WithKind("Route"):               &v1.Route{},
	v1.SchemeGroupVersion.WithKind("Service"):             &v1.Service{},

	autoscalingv1alpha1.SchemeGroupVersion.WithKind("PodAutoscaler"): &autoscalingv1alpha1.PodAutoscaler{},
	autoscalingv1alpha1.SchemeGroupVersion.WithKind("Metric"):        &autoscalingv1alpha1.Metric{},

	net.SchemeGroupVersion.WithKind("Certificate"):       &net.Certificate{},
	net.SchemeGroupVersion.WithKind("Ingress"):           &net.Ingress{},
	net.SchemeGroupVersion.WithKind("ServerlessService"): &net.ServerlessService{},

	// For group eventing.knative.dev.
	// v1beta1
	eventingv1beta1.SchemeGroupVersion.WithKind("EventType"): &eventingv1beta1.EventType{},
	// v1
	eventingv1.SchemeGroupVersion.WithKind("Broker"):  &eventingv1.Broker{},
	eventingv1.SchemeGroupVersion.WithKind("Trigger"): &eventingv1.Trigger{},

	// For group messaging.knative.dev.
	// v1
	messagingv1.SchemeGroupVersion.WithKind("Channel"):      &messagingv1.Channel{},
	messagingv1.SchemeGroupVersion.WithKind("Subscription"): &messagingv1.Subscription{},

	// For group sources.knative.dev.
	// v1beta2
	sourcesv1beta2.SchemeGroupVersion.WithKind("PingSource"): &sourcesv1beta2.PingSource{},
	// v1
	sourcesv1.SchemeGroupVersion.WithKind("ApiServerSource"): &sourcesv1.ApiServerSource{},
	sourcesv1.SchemeGroupVersion.WithKind("PingSource"):      &sourcesv1.PingSource{},
	sourcesv1.SchemeGroupVersion.WithKind("SinkBinding"):     &sourcesv1.SinkBinding{},
	sourcesv1.SchemeGroupVersion.WithKind("ContainerSource"): &sourcesv1.ContainerSource{},

	// For group flows.knative.dev
	// v1
	flowsv1.SchemeGroupVersion.WithKind("Parallel"): &flowsv1.Parallel{},
	flowsv1.SchemeGroupVersion.WithKind("Sequence"): &flowsv1.Sequence{},

	// For group tekton.dev
	// v1alpha1
	tknv1alpha1.SchemeGroupVersion.WithKind("Run"):              &tknv1alpha1.Run{},
	tknv1alpha1.SchemeGroupVersion.WithKind("Pipeline"):         &tknv1alpha1.Pipeline{},
	tknv1alpha1.SchemeGroupVersion.WithKind("Task"):             &tknv1alpha1.Task{},
	tknv1alpha1.SchemeGroupVersion.WithKind("ClusterTask"):      &tknv1alpha1.ClusterTask{},
	tknv1alpha1.SchemeGroupVersion.WithKind("TaskRun"):          &tknv1alpha1.TaskRun{},
	tknv1alpha1.SchemeGroupVersion.WithKind("PipelineRun"):      &tknv1alpha1.PipelineRun{},
	tknv1alpha1.SchemeGroupVersion.WithKind("Condition"):        &tknv1alpha1.Condition{},
	tknv1alpha1.SchemeGroupVersion.WithKind("PipelineResource"): &tknv1alpha1.PipelineResource{},
	// v1beta1
	tknv1beta1.SchemeGroupVersion.WithKind("Pipeline"):    &tknv1beta1.Pipeline{},
	tknv1beta1.SchemeGroupVersion.WithKind("Task"):        &tknv1beta1.Task{},
	tknv1beta1.SchemeGroupVersion.WithKind("ClusterTask"): &tknv1beta1.ClusterTask{},
	tknv1beta1.SchemeGroupVersion.WithKind("TaskRun"):     &tknv1beta1.TaskRun{},
	tknv1beta1.SchemeGroupVersion.WithKind("PipelineRun"): &tknv1beta1.PipelineRun{},
}
