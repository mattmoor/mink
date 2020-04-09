/*
Copyright 2020 VMware, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package resources

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	sourcesv1alpha1 "knative.dev/eventing/pkg/apis/sources/v1alpha1"
	duckv1alpha1 "knative.dev/pkg/apis/duck/v1alpha1"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/tracker"

	"github.com/vmware-tanzu/sources-for-knative/pkg/apis/sources/v1alpha1"
	"github.com/vmware-tanzu/sources-for-knative/pkg/reconciler/vspheresource/resources/names"
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
