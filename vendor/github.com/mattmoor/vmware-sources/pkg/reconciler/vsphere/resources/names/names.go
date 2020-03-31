/*
Copyright 2019 The Knative Authors

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

package names

import (
	"knative.dev/pkg/kmeta"

	"github.com/mattmoor/vmware-sources/pkg/apis/sources/v1alpha1"
)

func Deployment(vms *v1alpha1.VSphereSource) string {
	return kmeta.ChildName(vms.Name, "-deployment")
}

func SinkBinding(vms *v1alpha1.VSphereSource) string {
	return kmeta.ChildName(vms.Name, "-sinkbinding")
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
