/*
Copyright 2019 The Knative Authors.

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

package v1alpha1

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

var condSet = apis.NewLivingConditionSet(
	VSphereSourceConditionSourceReady,
	VSphereSourceConditionAuthReady,
	VSphereSourceConditionAdapterReady,
)

// GetGroupVersionKind implements kmeta.OwnerRefable
func (as *VSphereSource) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("VSphereSource")
}

func (ass *VSphereSourceStatus) InitializeConditions() {
	condSet.Manage(ass).InitializeConditions()
}

func (ass *VSphereSourceStatus) PropagateSourceStatus(status duckv1.SourceStatus) {
	// Copy over the whole of SourceStatus *except* Conditions.
	conds := ass.Conditions
	ass.SourceStatus = status
	ass.Conditions = conds

	cond := status.GetCondition(apis.ConditionReady)
	switch {
	case cond == nil:
		condSet.Manage(ass).MarkUnknown(VSphereSourceConditionSourceReady, "", "")
	case cond.Status == corev1.ConditionUnknown:
		condSet.Manage(ass).MarkUnknown(VSphereSourceConditionSourceReady, cond.Reason, cond.Message)
	case cond.Status == corev1.ConditionFalse:
		condSet.Manage(ass).MarkFalse(VSphereSourceConditionSourceReady, cond.Reason, cond.Message)
	case cond.Status == corev1.ConditionTrue:
		condSet.Manage(ass).MarkTrue(VSphereSourceConditionSourceReady)
	}
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
