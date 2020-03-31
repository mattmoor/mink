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

package resources

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	sourcesv1alpha1 "knative.dev/eventing/pkg/apis/sources/v1alpha1"
	duckv1alpha1 "knative.dev/pkg/apis/duck/v1alpha1"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/tracker"

	"github.com/mattmoor/vmware-sources/pkg/apis/sources/v1alpha1"
	"github.com/mattmoor/vmware-sources/pkg/reconciler/vsphere/resources/names"
)

func MakeSinkBinding(ctx context.Context, vms *v1alpha1.VSphereSource) *sourcesv1alpha1.SinkBinding {
	return &sourcesv1alpha1.SinkBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:            names.SinkBinding(vms),
			Namespace:       vms.Namespace,
			OwnerReferences: []metav1.OwnerReference{*kmeta.NewControllerRef(vms)},
		},
		Spec: sourcesv1alpha1.SinkBindingSpec{
			// Copy the SourceSpec wholesale.
			SourceSpec: vms.Spec.SourceSpec,
			// Bind to the Deployment for the receive adapter.
			BindingSpec: duckv1alpha1.BindingSpec{
				Subject: tracker.Reference{
					APIVersion: "apps/v1",
					Kind:       "Deployment",
					Namespace:  vms.Namespace,
					Name:       names.Deployment(vms),
				},
			},
		},
	}
}
