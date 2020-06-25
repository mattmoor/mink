/*
Copyright 2020 VMware, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package v1alpha1

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

var condSet = apis.NewLivingConditionSet(
	VSphereSourceConditionAuthReady,
	VSphereSourceConditionAdapterReady,
)

// GetConditionSet retrieves the condition set for this resource.
// Implements the KRShaped interface.
func (*VSphereSource) GetConditionSet() apis.ConditionSet {
	return condSet
}

// GetGroupVersionKind implements kmeta.OwnerRefable
func (as *VSphereSource) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("VSphereSource")
}

func (ass *VSphereSourceStatus) InitializeConditions() {
	condSet.Manage(ass).InitializeConditions()
}

func (ass *VSphereSourceStatus) PropagateAuthStatus(status duckv1.Status) {
	cond := status.GetCondition(apis.ConditionReady)
	switch {
	case cond == nil:
		condSet.Manage(ass).MarkUnknown(VSphereSourceConditionAuthReady, "", "")
	case cond.Status == corev1.ConditionUnknown:
		condSet.Manage(ass).MarkUnknown(VSphereSourceConditionAuthReady, cond.Reason, cond.Message)
	case cond.Status == corev1.ConditionFalse:
		condSet.Manage(ass).MarkFalse(VSphereSourceConditionAuthReady, cond.Reason, cond.Message)
	case cond.Status == corev1.ConditionTrue:
		condSet.Manage(ass).MarkTrue(VSphereSourceConditionAuthReady)
	}
}

func (ass *VSphereSourceStatus) PropagateAdapterStatus(d appsv1.DeploymentStatus) {
	// Check if the Deployment is available.
	for _, cond := range d.Conditions {
		if cond.Type == appsv1.DeploymentAvailable {
			switch {
			case cond.Status == corev1.ConditionUnknown:
				condSet.Manage(ass).MarkUnknown(VSphereSourceConditionAdapterReady, cond.Reason, cond.Message)
			case cond.Status == corev1.ConditionFalse:
				condSet.Manage(ass).MarkFalse(VSphereSourceConditionAdapterReady, cond.Reason, cond.Message)
			case cond.Status == corev1.ConditionTrue:
				condSet.Manage(ass).MarkTrue(VSphereSourceConditionAdapterReady)
			}
			return
		}
	}

	condSet.Manage(ass).MarkUnknown(VSphereSourceConditionAdapterReady, "", "")
}
