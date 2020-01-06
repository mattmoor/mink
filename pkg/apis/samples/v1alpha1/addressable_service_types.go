/*
Copyright 2019 The Knative Authors.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AddressableService is a Knative abstraction that encapsulates the interface by which Knative
// components express a desire to have a particular image cached.
type AddressableService struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the desired state of the AddressableService (from the client).
	// +optional
	Spec AddressableServiceSpec `json:"spec,omitempty"`

	// Status communicates the observed state of the AddressableService (from the controller).
	// +optional
	Status AddressableServiceStatus `json:"status,omitempty"`
}

// Check that AddressableService can be validated and defaulted.
var _ apis.Validatable = (*AddressableService)(nil)
var _ apis.Defaultable = (*AddressableService)(nil)
var _ kmeta.OwnerRefable = (*AddressableService)(nil)

// AddressableServiceSpec holds the desired state of the AddressableService (from the client).
type AddressableServiceSpec struct {
	// ServiceName holds the name of the Kubernetes Service to expose as an "addressable".
	ServiceName string `json:"serviceName"`
}

const (
	// AddressableServiceConditionReady is set when the revision is starting to materialize
	// runtime resources, and becomes true when those resources are ready.
	AddressableServiceConditionReady = apis.ConditionReady
)

// AddressableServiceStatus communicates the observed state of the AddressableService (from the controller).
type AddressableServiceStatus struct {
	duckv1.Status `json:",inline"`

	// Address holds the information needed to connect this Addressable up to receive events.
	// +optional
	Address *duckv1.Addressable `json:"address,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AddressableServiceList is a list of AddressableService resources
type AddressableServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []AddressableService `json:"items"`
}
