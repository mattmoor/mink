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

package addressableservice

import (
	"context"
	"reflect"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/network"
	"knative.dev/pkg/tracker"
	"knative.dev/sample-controller/pkg/apis/samples/v1alpha1"
	clientset "knative.dev/sample-controller/pkg/client/clientset/versioned"
	listers "knative.dev/sample-controller/pkg/client/listers/samples/v1alpha1"
)

// Reconciler implements controller.Reconciler for AddressableService resources.
type Reconciler struct {
	// Client is used to write back status updates.
	Client clientset.Interface

	// Listers index properties about resources
	Lister        listers.AddressableServiceLister
	ServiceLister corev1listers.ServiceLister

	// The tracker builds an index of what resources are watching other
	// resources so that we can immediately react to changes to changes in
	// tracked resources.
	Tracker tracker.Interface

	// Recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	Recorder record.EventRecorder
}

// Check that our Reconciler implements controller.Reconciler
var _ controller.Reconciler = (*Reconciler)(nil)

// Reconcile implements controller.Reconciler
func (r *Reconciler) Reconcile(ctx context.Context, key string) error {
	logger := logging.FromContext(ctx)

	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logger.Errorf("invalid resource key: %s", key)
		return nil
	}

	// If our controller has configuration state, we'd "freeze" it and
	// attach the frozen configuration to the context.
	//    ctx = r.configStore.ToContext(ctx)

	// Get the resource with this namespace/name.
	original, err := r.Lister.AddressableServices(namespace).Get(name)
	if apierrs.IsNotFound(err) {
		// The resource may no longer exist, in which case we stop processing.
		logger.Errorf("resource %q no longer exists", key)
		return nil
	} else if err != nil {
		return err
	}
	// Don't modify the informers copy.
	resource := original.DeepCopy()

	// Reconcile this copy of the resource and then write back any status
	// updates regardless of whether the reconciliation errored out.
	reconcileErr := r.reconcile(ctx, resource)
	if equality.Semantic.DeepEqual(original.Status, resource.Status) {
		// If we didn't change anything then don't call updateStatus.
		// This is important because the copy we loaded from the informer's
		// cache may be stale and we don't want to overwrite a prior update
		// to status with this stale state.
	} else if _, err = r.updateStatus(resource); err != nil {
		logger.Warnw("Failed to update resource status", zap.Error(err))
		r.Recorder.Eventf(resource, corev1.EventTypeWarning, "UpdateFailed",
			"Failed to update status for %q: %v", resource.Name, err)
		return err
	}
	if reconcileErr != nil {
		r.Recorder.Event(resource, corev1.EventTypeWarning, "InternalError", reconcileErr.Error())
	}
	return reconcileErr
}

func (r *Reconciler) reconcile(ctx context.Context, asvc *v1alpha1.AddressableService) error {
	if asvc.GetDeletionTimestamp() != nil {
		// Check for a DeletionTimestamp.  If present, elide the normal reconcile logic.
		// When a controller needs finalizer handling, it would go here.
		return nil
	}
	asvc.Status.InitializeConditions()

	if err := r.reconcileService(ctx, asvc); err != nil {
		return err
	}

	asvc.Status.ObservedGeneration = asvc.Generation
	return nil
}

func (r *Reconciler) reconcileService(ctx context.Context, asvc *v1alpha1.AddressableService) error {
	logger := logging.FromContext(ctx)

	if err := r.Tracker.TrackReference(tracker.Reference{
		APIVersion: "v1",
		Kind:       "Service",
		Name:       asvc.Spec.ServiceName,
		Namespace:  asvc.Namespace,
	}, asvc); err != nil {
		logger.Errorf("Error tracking service %s: %v", asvc.Spec.ServiceName, err)
		return err
	}

	_, err := r.ServiceLister.Services(asvc.Namespace).Get(asvc.Spec.ServiceName)
	if apierrs.IsNotFound(err) {
		logger.Info("Service does not yet exist:", asvc.Spec.ServiceName)
		asvc.Status.MarkServiceUnavailable(asvc.Spec.ServiceName)
		return nil
	} else if err != nil {
		logger.Errorf("Error reconciling service %s: %v", asvc.Spec.ServiceName, err)
		return err
	}

	asvc.Status.MarkServiceAvailable()
	asvc.Status.Address = &duckv1.Addressable{
		URL: &apis.URL{
			Scheme: "http",
			Host:   network.GetServiceHostname(asvc.Spec.ServiceName, asvc.Namespace),
		},
	}
	return nil
}

// Update the Status of the resource.  Caller is responsible for checking
// for semantic differences before calling.
func (r *Reconciler) updateStatus(desired *v1alpha1.AddressableService) (*v1alpha1.AddressableService, error) {
	actual, err := r.Lister.AddressableServices(desired.Namespace).Get(desired.Name)
	if err != nil {
		return nil, err
	}
	// If there's nothing to update, just return.
	if reflect.DeepEqual(actual.Status, desired.Status) {
		return actual, nil
	}
	// Don't modify the informers copy
	existing := actual.DeepCopy()
	existing.Status = desired.Status
	return r.Client.SamplesV1alpha1().AddressableServices(desired.Namespace).UpdateStatus(existing)
}
