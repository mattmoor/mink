/*
Copyright 2020 VMware, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	duckv1alpha1 "knative.dev/pkg/apis/duck/v1alpha1"
	"knative.dev/pkg/kmeta"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:defaulter-gen=true

// VSphereBinding describes a Binding that makes authenticating against
// a vSphere API simple.
type VSphereBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VSphereBindingSpec   `json:"spec"`
	Status VSphereBindingStatus `json:"status"`
}

// Check the interfaces that VSphereBinding should be implementing.
var (
	_ runtime.Object     = (*VSphereBinding)(nil)
	_ kmeta.OwnerRefable = (*VSphereBinding)(nil)
	_ apis.Validatable   = (*VSphereBinding)(nil)
	_ apis.Defaultable   = (*VSphereBinding)(nil)
	_ apis.HasSpec       = (*VSphereBinding)(nil)
)

// VSphereBindingSpec holds the desired state of the VSphereBinding (from the client).
type VSphereBindingSpec struct {
	duckv1alpha1.BindingSpec `json:",inline"`

	VAuthSpec `json:",inline"`
}

// VAuthSpec is the information used to authenticate with a vSphere API
type VAuthSpec struct {
	// Address contains the URL of the vSphere API.
	Address apis.URL `json:"address"`

	// SkipTLSVerify specifies whether the client should skip TLS verification when
	// talking to the vsphere address.
	SkipTLSVerify bool `json:"skipTLSVerify,omitempty"`

	// SecretRef is a reference to a Kubernetes secret of type kubernetes.io/basic-auth
	// which contains keys for "username" and "password", which will be used to authenticate
	//  with the vSphere API at "address".
	SecretRef corev1.LocalObjectReference `json:"secretRef"`
}

const (
	// VSphereBindingConditionReady is configured to indicate whether the Binding
	// has been configured for resources subject to its runtime contract.
	VSphereBindingConditionReady = apis.ConditionReady
)

// VSphereBindingStatus communicates the observed state of the VSphereBinding (from the controller).
type VSphereBindingStatus struct {
	duckv1.Status `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VSphereBindingList contains a list of VSphereBinding
type VSphereBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VSphereBinding `json:"items"`
}
