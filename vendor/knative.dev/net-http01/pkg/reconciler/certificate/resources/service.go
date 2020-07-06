/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License"); you
may not use this file except in compliance with the License.  You may
obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
implied.  See the License for the specific language governing
permissions and limitations under the License.
*/

package resources

import (
	"os"

	"knative.dev/networking/pkg/apis/networking/v1alpha1"
	"knative.dev/pkg/kmeta"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// TODO(mattmoor): Make the use of 8080 configurable as it's a
// fairly common port.
var (
	serviceSpec = corev1.ServiceSpec{
		Ports: []corev1.ServicePort{{
			Name:       "http-challenge",
			Port:       80,
			TargetPort: intstr.FromInt(8080),
		}},
	}

	endpointSubsets = []corev1.EndpointSubset{{
		Addresses: []corev1.EndpointAddress{{
			IP: os.Getenv("POD_IP"),
		}},
		Ports: []corev1.EndpointPort{{
			Name:     "http-challenge",
			Port:     8080,
			Protocol: corev1.ProtocolTCP,
		}},
	}}
)

// MakeService creates a Service, which we will point at ourselves.
// This service does not have a selector because it is created alongside
// the Certificate, but we will point it at our Pod running in the system
// namespace by directly manipulating Endpoints (see below).
func MakeService(o *v1alpha1.Certificate, opts ...func(*corev1.Service)) *corev1.Service {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            o.Name,
			Namespace:       o.Namespace,
			OwnerReferences: []metav1.OwnerReference{*kmeta.NewControllerRef(o)},
		},
		Spec: serviceSpec,
	}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
}

// MakeEndpoints creates an Endpoints, which we will point at ourselves.
// We populate the endpoints with out own Pod's IP address, which we get
// via the downward API.
func MakeEndpoints(o *v1alpha1.Certificate, opts ...func(*corev1.Endpoints)) *corev1.Endpoints {
	ep := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:            o.Name,
			Namespace:       o.Namespace,
			OwnerReferences: []metav1.OwnerReference{*kmeta.NewControllerRef(o)},
		},
		Subsets: endpointSubsets,
	}
	for _, opt := range opts {
		opt(ep)
	}
	return ep
}
