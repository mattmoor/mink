/*
Copyright 2020 The Knative Authors

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

// Code generated by injection-gen. DO NOT EDIT.

package metric

import (
	context "context"
	"encoding/json"
	"reflect"

	zap "go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	sets "k8s.io/apimachinery/pkg/util/sets"
	cache "k8s.io/client-go/tools/cache"
	record "k8s.io/client-go/tools/record"
	controller "knative.dev/pkg/controller"
	logging "knative.dev/pkg/logging"
	reconciler "knative.dev/pkg/reconciler"
	v1alpha1 "knative.dev/serving/pkg/apis/autoscaling/v1alpha1"
	versioned "knative.dev/serving/pkg/client/clientset/versioned"
	autoscalingv1alpha1 "knative.dev/serving/pkg/client/listers/autoscaling/v1alpha1"
)

// Interface defines the strongly typed interfaces to be implemented by a
// controller reconciling v1alpha1.Metric.
type Interface interface {
	// ReconcileKind implements custom logic to reconcile v1alpha1.Metric. Any changes
	// to the objects .Status or .Finalizers will be propagated to the stored
	// object. It is recommended that implementors do not call any update calls
	// for the Kind inside of ReconcileKind, it is the responsibility of the calling
	// controller to propagate those properties. The resource passed to ReconcileKind
	// will always have an empty deletion timestamp.
	ReconcileKind(ctx context.Context, o *v1alpha1.Metric) reconciler.Event
}

// Finalizer defines the strongly typed interfaces to be implemented by a
// controller finalizing v1alpha1.Metric.
type Finalizer interface {
	// FinalizeKind implements custom logic to finalize v1alpha1.Metric. Any changes
	// to the objects .Status or .Finalizers will be ignored. Returning a nil or
	// Normal type reconciler.Event will allow the finalizer to be deleted on
	// the resource. The resource passed to FinalizeKind will always have a set
	// deletion timestamp.
	FinalizeKind(ctx context.Context, o *v1alpha1.Metric) reconciler.Event
}

// reconcilerImpl implements controller.Reconciler for v1alpha1.Metric resources.
type reconcilerImpl struct {
	// Client is used to write back status updates.
	Client versioned.Interface

	// Listers index properties about resources
	Lister autoscalingv1alpha1.MetricLister

	// Recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	Recorder record.EventRecorder

	// configStore allows for decorating a context with config maps.
	// +optional
	configStore reconciler.ConfigStore

	// reconciler is the implementation of the business logic of the resource.
	reconciler Interface
}

// Check that our Reconciler implements controller.Reconciler
var _ controller.Reconciler = (*reconcilerImpl)(nil)

func NewReconciler(ctx context.Context, logger *zap.SugaredLogger, client versioned.Interface, lister autoscalingv1alpha1.MetricLister, recorder record.EventRecorder, r Interface, options ...controller.Options) controller.Reconciler {
	// Check the options function input. It should be 0 or 1.
	if len(options) > 1 {
		logger.Fatalf("up to one options struct is supported, found %d", len(options))
	}

	rec := &reconcilerImpl{
		Client:     client,
		Lister:     lister,
		Recorder:   recorder,
		reconciler: r,
	}

	for _, opts := range options {
		if opts.ConfigStore != nil {
			rec.configStore = opts.ConfigStore
		}
	}

	return rec
}

// Reconcile implements controller.Reconciler
func (r *reconcilerImpl) Reconcile(ctx context.Context, key string) error {
	logger := logging.FromContext(ctx)

	// If configStore is set, attach the frozen configuration to the context.
	if r.configStore != nil {
		ctx = r.configStore.ToContext(ctx)
	}

	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logger.Errorf("invalid resource key: %s", key)
		return nil
	}

	// Get the resource with this namespace/name.
	original, err := r.Lister.Metrics(namespace).Get(name)
	if errors.IsNotFound(err) {
		// The resource may no longer exist, in which case we stop processing.
		logger.Errorf("resource %q no longer exists", key)
		return nil
	} else if err != nil {
		return err
	}
	// Don't modify the informers copy.
	resource := original.DeepCopy()

	var reconcileEvent reconciler.Event
	if resource.GetDeletionTimestamp().IsZero() {
		// Append the target method to the logger.
		logger = logger.With(zap.String("targetMethod", "ReconcileKind"))

		// Set and update the finalizer on resource if r.reconciler
		// implements Finalizer.
		if err := r.setFinalizerIfFinalizer(ctx, resource); err != nil {
			logger.Warnw("Failed to set finalizers", zap.Error(err))
		}

		// Reconcile this copy of the resource and then write back any status
		// updates regardless of whether the reconciliation errored out.
		reconcileEvent = r.reconciler.ReconcileKind(ctx, resource)
	} else if fin, ok := r.reconciler.(Finalizer); ok {
		// Append the target method to the logger.
		logger = logger.With(zap.String("targetMethod", "FinalizeKind"))

		// For finalizing reconcilers, if this resource being marked for deletion
		// and reconciled cleanly (nil or normal event), remove the finalizer.
		reconcileEvent = fin.FinalizeKind(ctx, resource)
		if err := r.clearFinalizer(ctx, resource, reconcileEvent); err != nil {
			logger.Warnw("Failed to clear finalizers", zap.Error(err))
		}
	}

	// Synchronize the status.
	if equality.Semantic.DeepEqual(original.Status, resource.Status) {
		// If we didn't change anything then don't call updateStatus.
		// This is important because the copy we loaded from the injectionInformer's
		// cache may be stale and we don't want to overwrite a prior update
		// to status with this stale state.
	} else if err = r.updateStatus(original, resource); err != nil {
		logger.Warnw("Failed to update resource status", zap.Error(err))
		r.Recorder.Eventf(resource, v1.EventTypeWarning, "UpdateFailed",
			"Failed to update status for %q: %v", resource.Name, err)
		return err
	}

	// Report the reconciler event, if any.
	if reconcileEvent != nil {
		var event *reconciler.ReconcilerEvent
		if reconciler.EventAs(reconcileEvent, &event) {
			logger.Infow("returned an event", zap.Any("event", reconcileEvent))
			r.Recorder.Eventf(resource, event.EventType, event.Reason, event.Format, event.Args...)
			return nil
		} else {
			logger.Errorw("returned an error", zap.Error(reconcileEvent))
			r.Recorder.Event(resource, v1.EventTypeWarning, "InternalError", reconcileEvent.Error())
			return reconcileEvent
		}
	}
	return nil
}

func (r *reconcilerImpl) updateStatus(existing *v1alpha1.Metric, desired *v1alpha1.Metric) error {
	existing = existing.DeepCopy()
	return reconciler.RetryUpdateConflicts(func(attempts int) (err error) {
		// The first iteration tries to use the injectionInformer's state, subsequent attempts fetch the latest state via API.
		if attempts > 0 {
			existing, err = r.Client.AutoscalingV1alpha1().Metrics(desired.Namespace).Get(desired.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
		}

		// If there's nothing to update, just return.
		if reflect.DeepEqual(existing.Status, desired.Status) {
			return nil
		}

		existing.Status = desired.Status
		_, err = r.Client.AutoscalingV1alpha1().Metrics(existing.Namespace).UpdateStatus(existing)
		return err
	})
}

// updateFinalizersFiltered will update the Finalizers of the resource.
// TODO: this method could be generic and sync all finalizers. For now it only
// updates defaultFinalizerName.
func (r *reconcilerImpl) updateFinalizersFiltered(ctx context.Context, resource *v1alpha1.Metric) error {
	finalizerName := defaultFinalizerName

	actual, err := r.Lister.Metrics(resource.Namespace).Get(resource.Name)
	if err != nil {
		return err
	}

	// Don't modify the informers copy.
	existing := actual.DeepCopy()

	var finalizers []string

	// If there's nothing to update, just return.
	existingFinalizers := sets.NewString(existing.Finalizers...)
	desiredFinalizers := sets.NewString(resource.Finalizers...)

	if desiredFinalizers.Has(finalizerName) {
		if existingFinalizers.Has(finalizerName) {
			// Nothing to do.
			return nil
		}
		// Add the finalizer.
		finalizers = append(existing.Finalizers, finalizerName)
	} else {
		if !existingFinalizers.Has(finalizerName) {
			// Nothing to do.
			return nil
		}
		// Remove the finalizer.
		existingFinalizers.Delete(finalizerName)
		finalizers = existingFinalizers.List()
	}

	mergePatch := map[string]interface{}{
		"metadata": map[string]interface{}{
			"finalizers":      finalizers,
			"resourceVersion": existing.ResourceVersion,
		},
	}

	patch, err := json.Marshal(mergePatch)
	if err != nil {
		return err
	}

	_, err = r.Client.AutoscalingV1alpha1().Metrics(resource.Namespace).Patch(resource.Name, types.MergePatchType, patch)
	if err != nil {
		r.Recorder.Eventf(resource, v1.EventTypeWarning, "FinalizerUpdateFailed",
			"Failed to update finalizers for %q: %v", resource.Name, err)
	} else {
		r.Recorder.Eventf(resource, v1.EventTypeNormal, "FinalizerUpdate",
			"Updated %q finalizers", resource.GetName())
	}
	return err
}

func (r *reconcilerImpl) setFinalizerIfFinalizer(ctx context.Context, resource *v1alpha1.Metric) error {
	if _, ok := r.reconciler.(Finalizer); !ok {
		return nil
	}

	finalizers := sets.NewString(resource.Finalizers...)

	// If this resource is not being deleted, mark the finalizer.
	if resource.GetDeletionTimestamp().IsZero() {
		finalizers.Insert(defaultFinalizerName)
	}

	resource.Finalizers = finalizers.List()

	// Synchronize the finalizers filtered by defaultFinalizerName.
	return r.updateFinalizersFiltered(ctx, resource)
}

func (r *reconcilerImpl) clearFinalizer(ctx context.Context, resource *v1alpha1.Metric, reconcileEvent reconciler.Event) error {
	if _, ok := r.reconciler.(Finalizer); !ok {
		return nil
	}
	if resource.GetDeletionTimestamp().IsZero() {
		return nil
	}

	finalizers := sets.NewString(resource.Finalizers...)

	if reconcileEvent != nil {
		var event *reconciler.ReconcilerEvent
		if reconciler.EventAs(reconcileEvent, &event) {
			if event.EventType == v1.EventTypeNormal {
				finalizers.Delete(defaultFinalizerName)
			}
		}
	} else {
		finalizers.Delete(defaultFinalizerName)
	}

	resource.Finalizers = finalizers.List()

	// Synchronize the finalizers filtered by defaultFinalizerName.
	return r.updateFinalizersFiltered(ctx, resource)
}
