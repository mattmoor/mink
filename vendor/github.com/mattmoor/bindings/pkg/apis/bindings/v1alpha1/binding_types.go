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

	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/tracker"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GithubBinding is a Knative-style Binding for injecting Github credentials
// compatible with ./pkg/github into any Kubernetes resource with a Pod Spec.
type GithubBinding struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the desired state of the GithubBinding (from the client).
	// +optional
	Spec GithubBindingSpec `json:"spec,omitempty"`

	// Status communicates the observed state of the GithubBinding (from the controller).
	// +optional
	Status GithubBindingStatus `json:"status,omitempty"`
}

var (
	// Check that GithubBinding can be validated and defaulted.
	_ apis.Validatable   = (*GithubBinding)(nil)
	_ apis.Defaultable   = (*GithubBinding)(nil)
	_ kmeta.OwnerRefable = (*GithubBinding)(nil)
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SlackBinding is a Knative-style Binding for injecting Slack credentials
// compatible with ./pkg/slack into any Kubernetes resource with a Pod Spec.
type SlackBinding struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the desired state of the SlackBinding (from the client).
	// +optional
	Spec SlackBindingSpec `json:"spec,omitempty"`

	// Status communicates the observed state of the SlackBinding (from the controller).
	// +optional
	Status SlackBindingStatus `json:"status,omitempty"`
}

var (
	// Check that SlackBinding can be validated and defaulted.
	_ apis.Validatable   = (*SlackBinding)(nil)
	_ apis.Defaultable   = (*SlackBinding)(nil)
	_ kmeta.OwnerRefable = (*SlackBinding)(nil)
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TwitterBinding is a Knative-style Binding for injecting Twitter credentials
// compatible with ./pkg/slack into any Kubernetes resource with a Pod Spec.
type TwitterBinding struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the desired state of the TwitterBinding (from the client).
	// +optional
	Spec TwitterBindingSpec `json:"spec,omitempty"`

	// Status communicates the observed state of the TwitterBinding (from the controller).
	// +optional
	Status TwitterBindingStatus `json:"status,omitempty"`
}

var (
	// Check that TwitterBinding can be validated and defaulted.
	_ apis.Validatable   = (*TwitterBinding)(nil)
	_ apis.Defaultable   = (*TwitterBinding)(nil)
	_ kmeta.OwnerRefable = (*TwitterBinding)(nil)
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GoogleCloudSQLBinding is a Knative-style Binding for injecting GoogleCloudSQL credentials
// compatible with ./pkg/github into any Kubernetes resource with a Pod Spec.
type GoogleCloudSQLBinding struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the desired state of the GoogleCloudSQLBinding (from the client).
	// +optional
	Spec GoogleCloudSQLBindingSpec `json:"spec,omitempty"`

	// Status communicates the observed state of the GoogleCloudSQLBinding (from the controller).
	// +optional
	Status GoogleCloudSQLBindingStatus `json:"status,omitempty"`
}

var (
	// Check that GoogleCloudSQLBinding can be validated and defaulted.
	_ apis.Validatable   = (*GoogleCloudSQLBinding)(nil)
	_ apis.Defaultable   = (*GoogleCloudSQLBinding)(nil)
	_ kmeta.OwnerRefable = (*GoogleCloudSQLBinding)(nil)
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SQLBinding is a Knative-style Binding for injecting GoogleCloudSQL credentials
// compatible with ./pkg/github into any Kubernetes resource with a Pod Spec.
type SQLBinding struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the desired state of the SQLBinding (from the client).
	// +optional
	Spec SQLBindingSpec `json:"spec,omitempty"`

	// Status communicates the observed state of the SQLBinding (from the controller).
	// +optional
	Status SQLBindingStatus `json:"status,omitempty"`
}

var (
	// Check that SQLBinding can be validated and defaulted.
	_ apis.Validatable   = (*SQLBinding)(nil)
	_ apis.Defaultable   = (*SQLBinding)(nil)
	_ kmeta.OwnerRefable = (*SQLBinding)(nil)
)

// GithubBindingSpec holds the desired state of the GithubBinding (from the client).
type GithubBindingSpec struct {
	// Subject holds a reference to the "pod speccable" Kubernetes resource which will
	// be bound with Github secret data.
	Subject tracker.Reference `json:"subject"`

	// Secret holds a reference to a secret containing the Github auth data.
	Secret corev1.LocalObjectReference `json:"secret"`
}

// GithubBindingStatus communicates the observed state of the GithubBinding (from the controller).
type GithubBindingStatus struct {
	duckv1.Status `json:",inline"`
}

// SlackBindingSpec holds the desired state of the SlackBinding (from the client).
type SlackBindingSpec struct {
	// Subject holds a reference to the "pod speccable" Kubernetes resource which will
	// be bound with Slack secret data.
	Subject tracker.Reference `json:"subject"`

	// Secret holds a reference to a secret containing the Slack auth data.
	Secret corev1.LocalObjectReference `json:"secret"`
}

// SlackBindingStatus communicates the observed state of the SlackBinding (from the controller).
type SlackBindingStatus struct {
	duckv1.Status `json:",inline"`
}

// TwitterBindingSpec holds the desired state of the TwitterBinding (from the client).
type TwitterBindingSpec struct {
	// Subject holds a reference to the "pod speccable" Kubernetes resource which will
	// be bound with Twitter secret data.
	Subject tracker.Reference `json:"subject"`

	// Secret holds a reference to a secret containing the Twitter auth data.
	Secret corev1.LocalObjectReference `json:"secret"`
}

// TwitterBindingStatus communicates the observed state of the TwitterBinding (from the controller).
type TwitterBindingStatus struct {
	duckv1.Status `json:",inline"`
}

// GoogleCloudSQLBindingSpec holds the desired state of the GoogleCloudSQLBinding (from the client).
type GoogleCloudSQLBindingSpec struct {
	// Subject holds a reference to the "pod speccable" Kubernetes resource which will
	// be bound with GoogleCloudSQL secret data.
	Subject tracker.Reference `json:"subject"`

	// Secret holds a reference to a secret containing the CloudSQL auth data.
	Secret corev1.LocalObjectReference `json:"secret"`

	// Instance holds the name of the Cloud SQL instance to which the sidecar can connect.
	Instance string `json:"instance"`
}

// SQLBindingSpec holds the desired state of the SQLBinding (from the client).
type SQLBindingSpec struct {
	// Subject holds a reference to the "pod speccable" Kubernetes resource which will
	// be bound with SQL secret data.
	Subject tracker.Reference `json:"subject"`

	// Secret holds a reference to a secret containing the SQL auth data.
	Secret corev1.LocalObjectReference `json:"secret"`

	// Instance holds the name of the Cloud SQL instance to which the sidecar can connect.
	Instance string `json:"instance"`
}

// GoogleCloudSQLBindingStatus communicates the observed state of the GoogleCloudSQLBinding (from the controller).
type GoogleCloudSQLBindingStatus struct {
	duckv1.Status `json:",inline"`
}

// SQLBindingStatus communicates the observed state of the SQLBinding (from the controller).
type SQLBindingStatus struct {
	duckv1.Status `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GithubBindingList is a list of GithubBinding resources
type GithubBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []GithubBinding `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SlackBindingList is a list of SlackBinding resources
type SlackBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []SlackBinding `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TwitterBindingList is a list of TwitterBinding resources
type TwitterBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []TwitterBinding `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GoogleCloudSQLBindingList is a list of GoogleCloudSQLBinding resources
type GoogleCloudSQLBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []GoogleCloudSQLBinding `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SQLBindingList is a list of SQLBinding resources
type SQLBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []SQLBinding `json:"items"`
}
