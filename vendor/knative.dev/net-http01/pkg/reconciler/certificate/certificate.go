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

package certificate

import (
	context "context"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"knative.dev/net-http01/pkg/ordermanager"
	"knative.dev/net-http01/pkg/reconciler/certificate/resources"
	v1alpha1 "knative.dev/networking/pkg/apis/networking/v1alpha1"
	certificate "knative.dev/networking/pkg/client/injection/reconciler/networking/v1alpha1/certificate"
	logging "knative.dev/pkg/logging"
	reconciler "knative.dev/pkg/reconciler"
)

// Reconciler implements controller.Reconciler for Certificate resources.
type Reconciler struct {
	kubeClient kubernetes.Interface

	challengePort int

	secretLister    corev1listers.SecretLister
	serviceLister   corev1listers.ServiceLister
	endpointsLister corev1listers.EndpointsLister

	orderManager ordermanager.Interface
}

// Check that our Reconciler implements Interface
var _ certificate.Interface = (*Reconciler)(nil)

// ReconcileKind implements Interface.ReconcileKind.
func (r *Reconciler) ReconcileKind(ctx context.Context, o *v1alpha1.Certificate) reconciler.Event {
	o.Status.InitializeConditions()

	svc, err := r.reconcileService(ctx, o)
	if err != nil {
		return err
	}
	if err := r.reconcileEndpoints(ctx, o); err != nil {
		return err
	}

	// Lookup the secret, and ensure that it's contents are still valid.
	secret, err := r.secretLister.Secrets(o.Namespace).Get(o.Spec.SecretName)
	if apierrs.IsNotFound(err) {
		// We have to create it!
		logging.FromContext(ctx).Info("Secret doesn't exist, we must provision a new Certificate.")
	} else if err != nil {
		return err
	} else if valid, err := resources.IsValidCertificate(secret, o.Spec.DNSNames, 30*24*time.Hour); err == nil && valid {
		o.Status.MarkReady()
		o.Status.ObservedGeneration = o.Generation
		logging.FromContext(ctx).Info("Existing Certificate is valid.")
		return nil
	} else {
		logging.FromContext(ctx).Info("Certificate is not (or no longer) valid.")
	}

	// Don't let the OrderManager hang on client calls.
	// We don't "cancel" this context, because it is passed
	// to Go routines that extend pass this function's return.
	// TODO(mattmoor): 5 minutes is too long for this.
	// nolint
	ctx, _ = context.WithTimeout(ctx, 5*time.Minute)

	chall, cert, err := r.orderManager.Order(ctx, o.Spec.DNSNames, o)
	switch {
	case err != nil:
		return err

	case len(chall) != 0:
		o.Status.HTTP01Challenges = nil
		for _, url := range chall {
			o.Status.HTTP01Challenges = append(o.Status.HTTP01Challenges, v1alpha1.HTTP01Challenge{
				URL:              url,
				ServiceName:      svc.Name,
				ServiceNamespace: svc.Namespace, // Must be same namespace for KIngress
				ServicePort:      intstr.FromInt(80),
			})
		}
		o.Status.MarkNotReady("OrderCert", "Provisioning Certificate through HTTP01 challenges.")

	case cert != nil:
		wantSecret, err := resources.MakeSecret(o, cert)
		if err != nil {
			return err
		}
		if secret == nil {
			if _, err := r.kubeClient.CoreV1().Secrets(wantSecret.Namespace).Create(ctx, wantSecret, metav1.CreateOptions{}); err != nil {
				return err
			}
		} else {
			secret := secret.DeepCopy()
			secret.Data = wantSecret.Data
			if _, err := r.kubeClient.CoreV1().Secrets(secret.Namespace).Update(ctx, secret, metav1.UpdateOptions{}); err != nil {
				return err
			}
		}
		o.Status.MarkReady()
	}

	o.Status.ObservedGeneration = o.Generation
	return nil
}

func (r *Reconciler) reconcileService(ctx context.Context, o *v1alpha1.Certificate) (*corev1.Service, error) {
	svc, err := r.serviceLister.Services(o.Namespace).Get(resources.ServiceName(o))
	if apierrs.IsNotFound(err) {
		svc = resources.MakeService(o, resources.WithServicePort(r.challengePort))
		if _, err := r.kubeClient.CoreV1().Services(o.Namespace).Create(ctx, svc, metav1.CreateOptions{}); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else {
		desired := resources.MakeService(o, resources.WithServicePort(r.challengePort))
		if !equality.Semantic.DeepEqual(svc.Spec, desired.Spec) {
			updated := svc.DeepCopy()
			updated.Spec = desired.Spec
			updated.Spec.ClusterIP = svc.Spec.ClusterIP
			if svc, err = r.kubeClient.CoreV1().Services(o.Namespace).Update(ctx, updated, metav1.UpdateOptions{}); err != nil {
				return nil, err
			}
		}
	}
	return svc, nil
}

func (r *Reconciler) reconcileEndpoints(ctx context.Context, o *v1alpha1.Certificate) error {
	if ep, err := r.endpointsLister.Endpoints(o.Namespace).Get(resources.ServiceName(o)); apierrs.IsNotFound(err) {
		ep = resources.MakeEndpoints(o, resources.WithEndpointsPort(r.challengePort))
		if _, err := r.kubeClient.CoreV1().Endpoints(o.Namespace).Create(ctx, ep, metav1.CreateOptions{}); err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		desired := resources.MakeEndpoints(o, resources.WithEndpointsPort(r.challengePort))
		if !equality.Semantic.DeepEqual(ep.Subsets, desired.Subsets) {
			ep = ep.DeepCopy()
			ep.Subsets = desired.Subsets
			if _, err := r.kubeClient.CoreV1().Endpoints(o.Namespace).Update(ctx, ep, metav1.UpdateOptions{}); err != nil {
				return err
			}
		}
	}
	return nil
}
