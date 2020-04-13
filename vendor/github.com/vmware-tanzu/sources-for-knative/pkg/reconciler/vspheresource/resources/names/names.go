/*
Copyright 2020 VMware, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package names

import (
	"knative.dev/pkg/kmeta"

	"github.com/vmware-tanzu/sources-for-knative/pkg/apis/sources/v1alpha1"
)

func Deployment(vms *v1alpha1.VSphereSource) string {
	return kmeta.ChildName(vms.Name, "-deployment")
}

func VSphereBinding(vms *v1alpha1.VSphereSource) string {
	return kmeta.ChildName(vms.Name, "-vspherebinding")
}

func ConfigMap(vms *v1alpha1.VSphereSource) string {
	return kmeta.ChildName(vms.Name, "-configmap")
}

func RoleBinding(vms *v1alpha1.VSphereSource) string {
	return kmeta.ChildName(vms.Name, "-rolebinding")
}

func ServiceAccount(vms *v1alpha1.VSphereSource) string {
	return kmeta.ChildName(vms.Name, "-serviceaccount")
}
