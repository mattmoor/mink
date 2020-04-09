/*
Copyright 2020 VMware, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package resources

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"

	"github.com/vmware-tanzu/sources-for-knative/pkg/apis/sources/v1alpha1"
	"github.com/vmware-tanzu/sources-for-knative/pkg/reconciler/vspheresource/resources/names"
)

// MakeConfigMap creates a ConfigMap owned by the VSphereSource
func MakeConfigMap(ctx context.Context, vms *v1alpha1.VSphereSource) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            names.ConfigMap(vms),
			Namespace:       vms.Namespace,
			OwnerReferences: []metav1.OwnerReference{*kmeta.NewControllerRef(vms)},
		},
	}
}
