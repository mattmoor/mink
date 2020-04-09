/*
Copyright 2020 VMware, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package resources

import (
	"context"

	"github.com/vmware-tanzu/sources-for-knative/pkg/apis/sources/v1alpha1"
	"github.com/vmware-tanzu/sources-for-knative/pkg/reconciler/vspheresource/resources/names"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"
)

// MakeRoleBinding creates a RoleBinding object for the receive adapter
// service account 'sa' in the Namespace 'ns'. This is necessary for
// the receive adapter to be able to store state in configmaps.
func MakeRoleBinding(ctx context.Context, vms *v1alpha1.VSphereSource) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			OwnerReferences: []metav1.OwnerReference{*kmeta.NewControllerRef(vms)},
			Name:            names.RoleBinding(vms),
			Namespace:       vms.Namespace,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "receive-adapter-cm",
		},
		Subjects: []rbacv1.Subject{{
			Kind:      "ServiceAccount",
			Namespace: vms.Namespace,
			Name:      names.ServiceAccount(vms),
		}},
	}
}
