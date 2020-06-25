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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"
)

// +genclient
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
type PostgresSource struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the desired state of the PostgresSource (from the client).
	Spec PostgresSourceSpec `json:"spec"`

	// Status communicates the observed state of the PostgresSource (from the controller).
	// +optional
	Status PostgresSourceStatus `json:"status,omitempty"`
}

// GetGroupVersionKind returns the GroupVersionKind.
func (s *PostgresSource) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("PostgresSource")
}

// Check that PostgresSource is a runtime.Object.
var _ runtime.Object = (*PostgresSource)(nil)

// Check that we can create OwnerReferences to a PostgresSource.
var _ kmeta.OwnerRefable = (*PostgresSource)(nil)

// Check that PostgresSource implements the Conditions duck type.
var _ = duck.VerifyType(&PostgresSource{}, &duckv1.Conditions{})

const (
	// PostgresSourceEventType is the PostgresSource CloudEvent type.
	PostgresSourceEventType = "dev.vaikas.sources.postgressource"
)

// PostgresSourceSpec holds the desired state of the PostgresSource (from the client).
type PostgresSourceSpec struct {
	// inherits duck/v1 SourceSpec, which currently provides:
	// * Sink - a reference to an object that will resolve to a domain name or
	//   a URI directly to use as the sink.
	// * CloudEventOverrides - defines overrides to control the output format
	//   and modifications of the event sent to the sink.
	duckv1.SourceSpec `json:",inline"`

	// Tables to subscribe to
	// TODO: more granularity?
	Tables []TableSpec `json:"tables,omitempty"`

	// ServiceAccountName holds the name of the Kubernetes service account
	// as which the underlying K8s resources should be run. If unspecified
	// the controller will create a service account owned by this Source.
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// Secret to use when creating functions/triggers on the database.
	// You can use this to override the secret that's typically a
	// a binding for the entire controller.
	// The secret must have a field 'connectionstr' in the data
	// section, for example:
	// apiVersion: v1
	// kind: Secret
	// data:
	//  connectionstr: <base64 encoded connection string>
	// type: Opaque
	//
	Secret corev1.LocalObjectReference `json:"secret"`
}

type TableSpec struct {
	// TODO: make this more granular, like which operations, etc.
	Name string `json:"name"`
}

const (
	// PostgresSourceConditionReady is set when the revision is starting to materialize
	// runtime resources, and becomes true when those resources are ready.
	PostgresSourceConditionReady = apis.ConditionReady
)

// PostgresSourceStatus communicates the observed state of the PostgresSource (from the controller).
type PostgresSourceStatus struct {
	// inherits duck/v1 SourceStatus, which currently provides:
	// * ObservedGeneration - the 'Generation' of the Service that was last
	//   processed by the controller.
	// * Conditions - the latest available observations of a resource's current
	//   state.
	// * SinkURI - the current active sink URI that has been configured for the
	//   Source.
	duckv1.SourceStatus `json:",inline"`

	// Holds 1:1 status for the ones specified in the Spec.Tables
	TableStatuses []TableStatus `json:"tableStatuses,omitempty"`
}

type TableStatus struct {
	Subscribed bool `json:"subscribed"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PostgresSourceList is a list of PostgresSource resources
type PostgresSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []PostgresSource `json:"items"`
}

// GetStatus retrieves the status of the PostgresSource. Implements the KRShaped interface.
func (t *PostgresSource) GetStatus() *duckv1.Status {
	return &t.Status.Status
}
