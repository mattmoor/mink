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

package contour

import (
	"context"
	"fmt"
	"reflect"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	contourclientset "knative.dev/net-contour/pkg/client/clientset/versioned"
	contourlisters "knative.dev/net-contour/pkg/client/listers/projectcontour/v1"
	"knative.dev/net-contour/pkg/reconciler/contour/config"
	"knative.dev/net-contour/pkg/reconciler/contour/resources"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/network"
	"knative.dev/pkg/reconciler"
	"knative.dev/pkg/tracker"
	"knative.dev/serving/pkg/apis/networking"
	"knative.dev/serving/pkg/apis/networking/v1alpha1"
	clientset "knative.dev/serving/pkg/client/clientset/versioned"
	listers "knative.dev/serving/pkg/client/listers/networking/v1alpha1"
	"knative.dev/serving/pkg/network/status"
)

const (
	// ContourIngressClassName value for specifying knative's Contour
	// Ingress reconciler.
	ContourIngressClassName = "contour.ingress.networking.knative.dev"
)

// Reconciler implements controller.Reconciler for Ingress resources.
type Reconciler struct {
	// Client is used to write back status updates.
	client        clientset.Interface
	contourClient contourclientset.Interface

	// Listers index properties about resources
	lister          listers.IngressLister
	contourLister   contourlisters.HTTPProxyLister
	serviceLister   corev1listers.ServiceLister
	endpointsLister corev1listers.EndpointsLister

	// Recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder

	statusManager status.Manager
	configStore   reconciler.ConfigStore
	tracker       tracker.Interface
}

// Check that our Reconciler implements controller.Reconciler
var _ controller.Reconciler = (*Reconciler)(nil)

// Reconcile implements controller.Reconciler
func (r *Reconciler) Reconcile(ctx context.Context, key string) error {
	logger := logging.FromContext(ctx)
	ctx = r.configStore.ToContext(ctx)

	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logger.Errorf("Invalid resource key: %q.", key)
		return nil
	}

	// If our controller has configuration state, we'd "freeze" it and
	// attach the frozen configuration to the context.
	//    ctx = r.configStore.ToContext(ctx)

	// Get the resource with this namespace/name.
	original, err := r.lister.Ingresses(namespace).Get(name)
	if apierrs.IsNotFound(err) {
		// The resource may no longer exist, in which case we stop processing.
		logger.Errorf("Resource %q no longer exists.", key)
		return nil
	} else if err != nil {
		return err
	} else if original.Annotations != nil {
		class := original.Annotations[networking.IngressClassAnnotationKey]
		if class != ContourIngressClassName {
			logger.Debugf("Resource %q is not our class.", key)
			return nil
		}
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
		logger.Warnw("Failed to update resource status.", zap.Error(err))
		r.recorder.Eventf(resource, corev1.EventTypeWarning, "UpdateFailed",
			"Failed to update status for %q: %v", resource.Name, err)
		return err
	}
	if reconcileErr != nil {
		r.recorder.Event(resource, corev1.EventTypeWarning, "InternalError", reconcileErr.Error())
	}
	return reconcileErr
}

func (r *Reconciler) reconcile(ctx context.Context, ing *v1alpha1.Ingress) error {
	if ing.GetDeletionTimestamp() != nil {
		// Check for a DeletionTimestamp.  If present, elide the normal reconcile logic.
		// When a controller needs finalizer handling, it would go here.
		return nil
	}
	ing.Status.InitializeConditions()

	if err := r.reconcileProxies(ctx, ing); err != nil {
		return err
	}

	ing.Status.ObservedGeneration = ing.Generation
	return nil
}

func (r *Reconciler) reconcileProxies(ctx context.Context, ing *v1alpha1.Ingress) error {
	serviceNames := resources.ServiceNames(ctx, ing)
	serviceToProtocol := make(map[string]string, len(serviceNames))

	// Establish the protocol for each Service, and ensure that their Endpoints are
	// populated with Ready addresses before we reprogram Contour.
	for _, name := range serviceNames.List() {
		if err := r.tracker.TrackReference(tracker.Reference{
			APIVersion: "v1",
			Kind:       "Service",
			Namespace:  ing.Namespace,
			Name:       name,
		}, ing); err != nil {
			return err
		}
		svc, err := r.serviceLister.Services(ing.Namespace).Get(name)
		if err != nil {
			return err
		}
		for _, port := range svc.Spec.Ports {
			if port.Name == networking.ServicePortNameH2C {
				serviceToProtocol[name] = "h2c"
				break
			}
		}

		if err := r.tracker.TrackReference(tracker.Reference{
			APIVersion: "v1",
			Kind:       "Endpoints",
			Namespace:  ing.Namespace,
			Name:       name,
		}, ing); err != nil {
			return err
		}
		ep, err := r.endpointsLister.Endpoints(ing.Namespace).Get(name)
		if err != nil {
			return err
		}
		for _, subset := range ep.Subsets {
			if len(subset.Addresses) == 0 {
				ing.Status.MarkIngressNotReady("EndpointsNotReady",
					fmt.Sprintf("Waiting for Endpoints %q to have ready addresses.", name))
				return nil
			}
		}
	}

	for _, proxy := range resources.MakeHTTPProxies(ctx, ing, serviceToProtocol) {
		selector := labels.Set(map[string]string{
			resources.ParentKey:     proxy.Labels[resources.ParentKey],
			resources.DomainHashKey: proxy.Labels[resources.DomainHashKey],
		}).AsSelector()
		elts, err := r.contourLister.HTTPProxies(ing.Namespace).List(selector)
		if err != nil {
			return err
		}
		if len(elts) == 0 {
			if _, err := r.contourClient.ProjectcontourV1().HTTPProxies(proxy.Namespace).Create(proxy); err != nil {
				return err
			}
			continue
		}
		update := elts[0].DeepCopy()
		update.Annotations = proxy.Annotations
		update.Labels = proxy.Labels
		update.Spec = proxy.Spec
		if equality.Semantic.DeepEqual(elts[0], update) {
			// Avoid updates that don't change anything.
			continue
		}
		if _, err = r.contourClient.ProjectcontourV1().HTTPProxies(proxy.Namespace).Update(update); err != nil {
			return err
		}
	}

	if err := r.contourClient.ProjectcontourV1().HTTPProxies(ing.Namespace).DeleteCollection(
		&metav1.DeleteOptions{},
		metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s,%s!=%d",
				resources.ParentKey, ing.Name,
				resources.GenerationKey, ing.Generation),
		}); err != nil {
		return err
	}
	ing.Status.MarkNetworkConfigured()

	ready, err := r.statusManager.IsReady(ctx, ing)
	if err != nil {
		return fmt.Errorf("failed to probe Ingress %s/%s: %w", ing.GetNamespace(), ing.GetName(), err)
	}
	if ready {
		ing.Status.MarkLoadBalancerReady(
			[]v1alpha1.LoadBalancerIngressStatus{},
			lbStatus(ctx, v1alpha1.IngressVisibilityExternalIP),
			lbStatus(ctx, v1alpha1.IngressVisibilityClusterLocal))
	} else {
		ing.Status.MarkLoadBalancerNotReady()
	}
	return nil
}

func lbStatus(ctx context.Context, vis v1alpha1.IngressVisibility) (lbs []v1alpha1.LoadBalancerIngressStatus) {
	if keys, ok := config.FromContext(ctx).Contour.VisibilityKeys[vis]; ok {
		for _, key := range keys.List() {
			namespace, name, _ := cache.SplitMetaNamespaceKey(key)
			lbs = append(lbs, v1alpha1.LoadBalancerIngressStatus{
				DomainInternal: network.GetServiceHostname(name, namespace),
			})
		}
	}
	return
}

// Update the Status of the resource.  Caller is responsible for checking
// for semantic differences before calling.
func (r *Reconciler) updateStatus(desired *v1alpha1.Ingress) (*v1alpha1.Ingress, error) {
	actual, err := r.lister.Ingresses(desired.Namespace).Get(desired.Name)
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
	return r.client.NetworkingV1alpha1().Ingresses(desired.Namespace).UpdateStatus(existing)
}
