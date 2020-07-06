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
	"knative.dev/eventing/pkg/apis/duck"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

const (
	// PostgresConditionReady has status True when the PostgresSource is ready to send events.
	PostgresConditionReady = apis.ConditionReady

	// PostgresConditionSinkProvided has status True when the PostgresSource has been configured with a sink target.
	PostgresConditionSinkProvided apis.ConditionType = "SinkProvided"

	// PostgresConditionDeployed has status True when the PostgresSource has had it's deployment created.
	PostgresConditionDeployed apis.ConditionType = "Deployed"

	// PostgresFunctionCreated has status True when Function has been created
	PostgresFunctionCreated apis.ConditionType = "FunctionCreated"

	// PostgresTriggersCreated has status True when triggers have been created
	PostgresTriggersCreated apis.ConditionType = "TriggersCreated"

	// PostgresAdapterBindingReady has status True when the proper SQL binding for RA has been created
	PostgresAdapterBindingReady apis.ConditionType = "SQLBindingCreated"
)

var PostgresCondSet = apis.NewLivingConditionSet(
	PostgresConditionSinkProvided,
	PostgresConditionDeployed,
	PostgresFunctionCreated,
	PostgresTriggersCreated,
	PostgresAdapterBindingReady,
)

// GetConditionSet retrieves the condition set for this resource.
// Implements the KRShaped interface.
func (*PostgresSource) GetConditionSet() apis.ConditionSet {
	return PostgresCondSet
}

// GetCondition returns the condition currently associated with the given type, or nil.
func (s *PostgresSourceStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return PostgresCondSet.Manage(s).GetCondition(t)
}

// InitializeConditions sets relevant unset conditions to Unknown state.
func (s *PostgresSourceStatus) InitializeConditions() {
	PostgresCondSet.Manage(s).InitializeConditions()
}

// MarkSink sets the condition that the source has a sink configured.
func (s *PostgresSourceStatus) MarkSink(uri *apis.URL) {
	s.SinkURI = uri
	if len(uri.String()) > 0 {
		PostgresCondSet.Manage(s).MarkTrue(PostgresConditionSinkProvided)
	} else {
		PostgresCondSet.Manage(s).MarkUnknown(PostgresConditionSinkProvided, "SinkEmpty", "Sink has resolved to empty.")
	}
}

// MarkNoSink sets the condition that the source does not have a sink configured.
func (s *PostgresSourceStatus) MarkNoSink(reason, messageFormat string, messageA ...interface{}) {
	PostgresCondSet.Manage(s).MarkFalse(PostgresConditionSinkProvided, reason, messageFormat, messageA...)
}

// PropagateDeploymentAvailability uses the availability of the provided Deployment to determine if
// PostgresConditionDeployed should be marked as true or false.
func (s *PostgresSourceStatus) PropagateDeploymentAvailability(d *appsv1.Deployment) {
	if duck.DeploymentIsAvailable(&d.Status, false) {
		PostgresCondSet.Manage(s).MarkTrue(PostgresConditionDeployed)
	} else {
		// I don't know how to propagate the status well, so just give the name of the Deployment
		// for now.
		PostgresCondSet.Manage(s).MarkFalse(PostgresConditionDeployed, "DeploymentUnavailable", "The Deployment '%s' is unavailable.", d.Name)
	}
}

// PropagateFunctionCreated propagates the status of the postgres function
func (s *PostgresSourceStatus) PropagateFunctionCreated(exists bool, err error) {
	if exists {
		PostgresCondSet.Manage(s).MarkTrue(PostgresFunctionCreated)
	} else {
		// I don't know how to propagate the status well, so just set the error
		PostgresCondSet.Manage(s).MarkFalse(PostgresFunctionCreated, "FunctionDoesNotExist", "The function does not exist: %s", err)
	}
}

// PropagateTriggersCreated propagates the status of the postgres triggers
func (s *PostgresSourceStatus) PropagateTriggersCreated(exists bool, err error) {
	if exists {
		PostgresCondSet.Manage(s).MarkTrue(PostgresTriggersCreated)
	} else {
		// I don't know how to propagate the status well, so just set the error
		PostgresCondSet.Manage(s).MarkFalse(PostgresTriggersCreated, "TriggersDoesNotExist", "The function does not exist: %s", err)
	}
}

func (s *PostgresSourceStatus) PropagateAuthStatus(status duckv1.Status) {
	cond := status.GetCondition(apis.ConditionReady)
	switch {
	case cond == nil:
		PostgresCondSet.Manage(s).MarkUnknown(PostgresAdapterBindingReady, "", "")
	case cond.Status == corev1.ConditionUnknown:
		PostgresCondSet.Manage(s).MarkUnknown(PostgresAdapterBindingReady, cond.Reason, cond.Message)
	case cond.Status == corev1.ConditionFalse:
		PostgresCondSet.Manage(s).MarkFalse(PostgresAdapterBindingReady, cond.Reason, cond.Message)
	case cond.Status == corev1.ConditionTrue:
		PostgresCondSet.Manage(s).MarkTrue(PostgresAdapterBindingReady)
	}
}

// IsReady returns true if the resource is ready overall.
func (s *PostgresSourceStatus) IsReady() bool {
	return PostgresCondSet.Manage(s).IsHappy()
}
