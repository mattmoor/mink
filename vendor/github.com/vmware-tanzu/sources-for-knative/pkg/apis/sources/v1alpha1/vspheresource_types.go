/*
Copyright 2020 VMware, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"
)

// +genclient
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VSphereSource is a Knative abstraction that encapsulates the interface by which Knative
// components express a desire to have a particular image cached.
type VSphereSource struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the desired state of the VSphereSource (from the client).
	// +optional
	Spec VSphereSourceSpec `json:"spec,omitempty"`

	// Status communicates the observed state of the VSphereSource (from the controller).
	// +optional
	Status VSphereSourceStatus `json:"status,omitempty"`
}

// Check that VSphereSource can be validated and defaulted.
var _ apis.Validatable = (*VSphereSource)(nil)
var _ apis.Defaultable = (*VSphereSource)(nil)
var _ kmeta.OwnerRefable = (*VSphereSource)(nil)

// VSphereSourceSpec holds the desired state of the VSphereSource (from the client).
type VSphereSourceSpec struct {
	duckv1.SourceSpec `json:",inline"`

	VAuthSpec `json:",inline"`
}

const (
	// VSphereSourceConditionReady is set to reflect the overall state of the resource.
	VSphereSourceConditionReady = apis.ConditionReady

	// VSphereSourceConditionSourceReady is set to reflect the state of the source part of the VSphereSource.
	VSphereSourceConditionSourceReady = "SourceReady"

	// VSphereSourceConditionAuthReady is set to reflect the state of the auth part of the VSphereSource.
	VSphereSourceConditionAuthReady = "AuthReady"

	// VSphereSourceConditionAdapterReady is set to reflect the state of the adapter part of the VSphereSource.
	VSphereSourceConditionAdapterReady = "AdapterReady"
)

// VSphereSourceStatus communicates the observed state of the VSphereSource (from the controller).
type VSphereSourceStatus struct {
	duckv1.SourceStatus `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VSphereSourceList is a list of VSphereSource resources
type VSphereSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []VSphereSource `json:"items"`
}

// GetStatus retrieves the status of the VSphereSource. Implements the KRShaped interface.
func (t *VSphereSource) GetStatus() *duckv1.Status {
	return &t.Status.Status
}
