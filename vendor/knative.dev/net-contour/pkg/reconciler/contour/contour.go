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

	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"

	contourclientset "knative.dev/net-contour/pkg/client/clientset/versioned"
	contourlisters "knative.dev/net-contour/pkg/client/listers/projectcontour/v1"
	ingressreconciler "knative.dev/networking/pkg/client/injection/reconciler/networking/v1alpha1/ingress"

	"knative.dev/net-contour/pkg/reconciler/contour/config"
	"knative.dev/net-contour/pkg/reconciler/contour/resources"
	"knative.dev/networking/pkg/apis/networking"
	"knative.dev/networking/pkg/apis/networking/v1alpha1"
	"knative.dev/pkg/network"
	"knative.dev/pkg/reconciler"
	"knative.dev/pkg/tracker"
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
	contourClient contourclientset.Interface

	// Listers index properties about resources
	contourLister   contourlisters.HTTPProxyLister
	serviceLister   corev1listers.ServiceLister
	endpointsLister corev1listers.EndpointsLister

	statusManager status.Manager
	tracker       tracker.Interface
}

var _ ingressreconciler.Interface = (*Reconciler)(nil)

// ReconcileKind reconciles ingress resource.
func (r *Reconciler) ReconcileKind(ctx context.Context, ing *v1alpha1.Ingress) reconciler.Event {
	if ann := ing.Annotations[networking.IngressClassAnnotationKey]; ann != ContourIngressClassName {
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
