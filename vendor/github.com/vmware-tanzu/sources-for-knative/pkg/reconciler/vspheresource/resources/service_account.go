/*
Copyright 2020 VMware, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package resources

import (
	"context"

	"github.com/vmware-tanzu/sources-for-knative/pkg/apis/sources/v1alpha1"
	"github.com/vmware-tanzu/sources-for-knative/pkg/reconciler/vspheresource/resources/names"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"
)

// MakeServiceAccount creates a ServiceAccount object for the Namespace 'ns'.
func MakeServiceAccount(ctx context.Context, vms *v1alpha1.VSphereSource) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			OwnerReferences: []metav1.OwnerReference{*kmeta.NewControllerRef(vms)},
			Namespace:       vms.Namespace,
			Name:            names.ServiceAccount(vms),
		},
	}
}
