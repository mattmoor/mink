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

	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	messagingv1beta1 "knative.dev/eventing/pkg/apis/messaging/v1beta1"
	sourcesv1alpha1 "knative.dev/eventing/pkg/apis/sources/v1alpha1"
	sourcesv1alpha2 "knative.dev/eventing/pkg/apis/sources/v1alpha2"
	autoscalingv1alpha1 "knative.dev/serving/pkg/apis/autoscaling/v1alpha1"
	net "knative.dev/serving/pkg/apis/networking/v1alpha1"
	v1 "knative.dev/serving/pkg/apis/serving/v1"
	"knative.dev/serving/pkg/apis/serving/v1alpha1"
	"knative.dev/serving/pkg/apis/serving/v1beta1"

	tknv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	tknv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
)

var ourTypes = map[schema.GroupVersionKind]resourcesemantics.GenericCRD{
	v1alpha1.SchemeGroupVersion.WithKind("Revision"):      &v1alpha1.Revision{},
	v1alpha1.SchemeGroupVersion.WithKind("Configuration"): &v1alpha1.Configuration{},
	v1alpha1.SchemeGroupVersion.WithKind("Route"):         &v1alpha1.Route{},
	v1alpha1.SchemeGroupVersion.WithKind("Service"):       &v1alpha1.Service{},
	v1beta1.SchemeGroupVersion.WithKind("Revision"):       &v1beta1.Revision{},
	v1beta1.SchemeGroupVersion.WithKind("Configuration"):  &v1beta1.Configuration{},
	v1beta1.SchemeGroupVersion.WithKind("Route"):          &v1beta1.Route{},
	v1beta1.SchemeGroupVersion.WithKind("Service"):        &v1beta1.Service{},
	v1.SchemeGroupVersion.WithKind("Revision"):            &v1.Revision{},
	v1.SchemeGroupVersion.WithKind("Configuration"):       &v1.Configuration{},
	v1.SchemeGroupVersion.WithKind("Route"):               &v1.Route{},
	v1.SchemeGroupVersion.WithKind("Service"):             &v1.Service{},

	autoscalingv1alpha1.SchemeGroupVersion.WithKind("PodAutoscaler"): &autoscalingv1alpha1.PodAutoscaler{},
	autoscalingv1alpha1.SchemeGroupVersion.WithKind("Metric"):        &autoscalingv1alpha1.Metric{},

	net.SchemeGroupVersion.WithKind("Certificate"):       &net.Certificate{},
	net.SchemeGroupVersion.WithKind("Ingress"):           &net.Ingress{},
	net.SchemeGroupVersion.WithKind("ServerlessService"): &net.ServerlessService{},

	// For group messaging.knative.dev.
	// v1alpha1
	messagingv1alpha1.SchemeGroupVersion.WithKind("InMemoryChannel"): &messagingv1alpha1.InMemoryChannel{},
	messagingv1alpha1.SchemeGroupVersion.WithKind("Channel"):         &messagingv1alpha1.Channel{},
	messagingv1alpha1.SchemeGroupVersion.WithKind("Subscription"):    &messagingv1alpha1.Subscription{},
	// v1beta1
	messagingv1beta1.SchemeGroupVersion.WithKind("InMemoryChannel"): &messagingv1beta1.InMemoryChannel{},
	messagingv1beta1.SchemeGroupVersion.WithKind("Channel"):         &messagingv1beta1.Channel{},
	messagingv1beta1.SchemeGroupVersion.WithKind("Subscription"):    &messagingv1beta1.Subscription{},

	// For group sources.knative.dev.
	// v1alpha1
	sourcesv1alpha1.SchemeGroupVersion.WithKind("ApiServerSource"): &sourcesv1alpha1.ApiServerSource{},
	sourcesv1alpha1.SchemeGroupVersion.WithKind("PingSource"):      &sourcesv1alpha1.PingSource{},
	sourcesv1alpha1.SchemeGroupVersion.WithKind("SinkBinding"):     &sourcesv1alpha1.SinkBinding{},
	// v1alpha2
	sourcesv1alpha2.SchemeGroupVersion.WithKind("ApiServerSource"): &sourcesv1alpha2.ApiServerSource{},
	sourcesv1alpha2.SchemeGroupVersion.WithKind("PingSource"):      &sourcesv1alpha2.PingSource{},
	sourcesv1alpha2.SchemeGroupVersion.WithKind("SinkBinding"):     &sourcesv1alpha2.SinkBinding{},

	tknv1alpha1.SchemeGroupVersion.WithKind("Pipeline"):         &tknv1alpha1.Pipeline{},
	tknv1alpha1.SchemeGroupVersion.WithKind("Task"):             &tknv1alpha1.Task{},
	tknv1alpha1.SchemeGroupVersion.WithKind("ClusterTask"):      &tknv1alpha1.ClusterTask{},
	tknv1alpha1.SchemeGroupVersion.WithKind("TaskRun"):          &tknv1alpha1.TaskRun{},
	tknv1alpha1.SchemeGroupVersion.WithKind("PipelineRun"):      &tknv1alpha1.PipelineRun{},
	tknv1alpha1.SchemeGroupVersion.WithKind("Condition"):        &tknv1alpha1.Condition{},
	tknv1alpha1.SchemeGroupVersion.WithKind("PipelineResource"): &tknv1alpha1.PipelineResource{},
	tknv1beta1.SchemeGroupVersion.WithKind("Pipeline"):          &tknv1beta1.Pipeline{},
	tknv1beta1.SchemeGroupVersion.WithKind("Task"):              &tknv1beta1.Task{},
	tknv1beta1.SchemeGroupVersion.WithKind("ClusterTask"):       &tknv1beta1.ClusterTask{},
	tknv1beta1.SchemeGroupVersion.WithKind("TaskRun"):           &tknv1beta1.TaskRun{},
	tknv1beta1.SchemeGroupVersion.WithKind("PipelineRun"):       &tknv1beta1.PipelineRun{},
}
