/*
Copyright 2020 VMware, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package vspheresource

import (
	"context"
	"fmt"

	sourcesv1alpha1 "github.com/vmware-tanzu/sources-for-knative/pkg/apis/sources/v1alpha1"
	clientset "github.com/vmware-tanzu/sources-for-knative/pkg/client/clientset/versioned"
	vspherereconciler "github.com/vmware-tanzu/sources-for-knative/pkg/client/injection/reconciler/sources/v1alpha1/vspheresource"
	v1alpha1lister "github.com/vmware-tanzu/sources-for-knative/pkg/client/listers/sources/v1alpha1"
	"github.com/vmware-tanzu/sources-for-knative/pkg/reconciler/vspheresource/resources"
	resourcenames "github.com/vmware-tanzu/sources-for-knative/pkg/reconciler/vspheresource/resources/names"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	appsv1listers "k8s.io/client-go/listers/apps/v1"
	corev1Listers "k8s.io/client-go/listers/core/v1"
	rbacv1listers "k8s.io/client-go/listers/rbac/v1"
	eventingclientset "knative.dev/eventing/pkg/client/clientset/versioned"
	sourcesv1alpha1lister "knative.dev/eventing/pkg/client/listers/sources/v1alpha1"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/reconciler"
)

// Reconciler implements vspherereconciler.Interface for
// VSphereSource resources.
type Reconciler struct {
	adapterImage string

	kubeclient     kubernetes.Interface
	eventingclient eventingclientset.Interface
	client         clientset.Interface

	deploymentLister     appsv1listers.DeploymentLister
	vspherebindingLister v1alpha1lister.VSphereBindingLister
	sinkbindingLister    sourcesv1alpha1lister.SinkBindingLister
	rbacLister           rbacv1listers.RoleBindingLister
	cmLister             corev1Listers.ConfigMapLister
	saLister             corev1Listers.ServiceAccountLister
}

// Check that our Reconciler implements Interface
var _ vspherereconciler.Interface = (*Reconciler)(nil)

// ReconcileKind implements Interface.ReconcileKind.
func (r *Reconciler) ReconcileKind(ctx context.Context, vms *sourcesv1alpha1.VSphereSource) reconciler.Event {
	vms.Status.InitializeConditions()

	if err := r.reconcileSinkBinding(ctx, vms); err != nil {
		return err
	}
	if err := r.reconcileVSphereBinding(ctx, vms); err != nil {
		return err
	}

	// Make sure the ConfigMap for storing state exists before we
	// create the deployment so that it gets created as owned
	// by the source and hence won't be leaked.
	if err := r.reconcileConfigMap(ctx, vms); err != nil {
		return err
	}
	if err := r.reconcileServiceAccount(ctx, vms); err != nil {
		return err
	}
	if err := r.reconcileRoleBinding(ctx, vms); err != nil {
		return err
	}
	if err := r.reconcileDeployment(ctx, vms); err != nil {
		return err
	}

	vms.Status.ObservedGeneration = vms.Generation
	return nil
}

func (r *Reconciler) reconcileSinkBinding(ctx context.Context, vms *sourcesv1alpha1.VSphereSource) error {
	ns := vms.Namespace
	sinkbindingName := resourcenames.SinkBinding(vms)

	sinkbinding, err := r.sinkbindingLister.SinkBindings(ns).Get(sinkbindingName)
	if apierrs.IsNotFound(err) {
		sinkbinding = resources.MakeSinkBinding(ctx, vms)
		sinkbinding, err = r.eventingclient.SourcesV1alpha1().SinkBindings(ns).Create(sinkbinding)
		if err != nil {
			return fmt.Errorf("failed to create sinkbinding %q: %w", sinkbindingName, err)
		}
		logging.FromContext(ctx).Infof("Created sinkbinding %q", sinkbindingName)
	} else if err != nil {
		return fmt.Errorf("failed to get sinkbinding %q: %w", sinkbindingName, err)
	} else {
		// The sinkbinding exists, but make sure that it has the shape that we expect.
		desiredSinkBinding := resources.MakeSinkBinding(ctx, vms)
		sinkbinding = sinkbinding.DeepCopy()
		sinkbinding.Spec = desiredSinkBinding.Spec
		sinkbinding, err = r.eventingclient.SourcesV1alpha1().SinkBindings(ns).Update(sinkbinding)
		if err != nil {
			return fmt.Errorf("failed to create sinkbinding %q: %w", sinkbindingName, err)
		}
	}

	// Reflect the state of the SinkBinding in the VSphereSource
	vms.Status.PropagateSourceStatus(sinkbinding.Status.SourceStatus)

	return nil
}

func (r *Reconciler) reconcileVSphereBinding(ctx context.Context, vms *sourcesv1alpha1.VSphereSource) error {
	ns := vms.Namespace
	vspherebindingName := resourcenames.VSphereBinding(vms)

	vspherebinding, err := r.vspherebindingLister.VSphereBindings(ns).Get(vspherebindingName)
	if apierrs.IsNotFound(err) {
		vspherebinding = resources.MakeVSphereBinding(ctx, vms)
		vspherebinding, err = r.client.SourcesV1alpha1().VSphereBindings(ns).Create(vspherebinding)
		if err != nil {
			return fmt.Errorf("failed to create vspherebinding %q: %w", vspherebindingName, err)
		}
		logging.FromContext(ctx).Infof("Created vspherebinding %q", vspherebindingName)
	} else if err != nil {
		return fmt.Errorf("failed to get vspherebinding %q: %w", vspherebindingName, err)
	} else {
		// The vspherebinding exists, but make sure that it has the shape that we expect.
		desiredVSphereBinding := resources.MakeVSphereBinding(ctx, vms)
		vspherebinding = vspherebinding.DeepCopy()
		vspherebinding.Spec = desiredVSphereBinding.Spec
		vspherebinding, err = r.client.SourcesV1alpha1().VSphereBindings(ns).Update(vspherebinding)
		if err != nil {
			return fmt.Errorf("failed to create vspherebinding %q: %w", vspherebindingName, err)
		}
	}

	// Reflect the state of the VSphereBinding in the VSphereSource
	vms.Status.PropagateAuthStatus(vspherebinding.Status.Status)

	return nil
}

func (r *Reconciler) reconcileConfigMap(ctx context.Context, vms *sourcesv1alpha1.VSphereSource) error {
	ns := vms.Namespace
	name := resourcenames.ConfigMap(vms)

	cm, err := r.cmLister.ConfigMaps(ns).Get(name)
	// Note that we only create the configmap if it does not exist so that we get the
	// OwnerRefs set up properly so it gets Garbage Collected.
	if apierrs.IsNotFound(err) {
		cm = resources.MakeConfigMap(ctx, vms)
		cm, err = r.kubeclient.CoreV1().ConfigMaps(ns).Create(cm)
		if err != nil {
			return fmt.Errorf("failed to create configmap %q: %w", name, err)
		}
		logging.FromContext(ctx).Infof("Created configmap %q", name)
	} else if err != nil {
		return fmt.Errorf("failed to get configmap %q: %w", name, err)
	}

	return nil
}

func (r *Reconciler) reconcileServiceAccount(ctx context.Context, vms *sourcesv1alpha1.VSphereSource) error {
	ns := vms.Namespace
	name := resourcenames.ServiceAccount(vms)

	sa, err := r.saLister.ServiceAccounts(ns).Get(name)
	if apierrs.IsNotFound(err) {
		sa = resources.MakeServiceAccount(ctx, vms)
		sa, err = r.kubeclient.CoreV1().ServiceAccounts(ns).Create(sa)
		if err != nil {
			return fmt.Errorf("failed to create serviceaccount %q: %w", name, err)
		}
		logging.FromContext(ctx).Infof("Created serviceaccount %q", name)
	} else if err != nil {
		return fmt.Errorf("failed to get serviceaccount %q: %w", name, err)
	}

	return nil
}

func (r *Reconciler) reconcileRoleBinding(ctx context.Context, vms *sourcesv1alpha1.VSphereSource) error {
	ns := vms.Namespace
	name := resourcenames.RoleBinding(vms)
	roleBinding, err := r.rbacLister.RoleBindings(ns).Get(name)
	if apierrs.IsNotFound(err) {
		roleBinding = resources.MakeRoleBinding(ctx, vms)
		roleBinding, err = r.kubeclient.RbacV1().RoleBindings(ns).Create(roleBinding)
		if err != nil {
			return fmt.Errorf("failed to create rolebinding %q: %w", name, err)
		}
		logging.FromContext(ctx).Infof("Created rolebinding %q", name)
	} else if err != nil {
		return fmt.Errorf("failed to get rolebinding %q: %w", name, err)
	}
	// TODO: diff the roleref / subjects and update as necessary.
	return nil
}

func (r *Reconciler) reconcileDeployment(ctx context.Context, vms *sourcesv1alpha1.VSphereSource) error {
	ns := vms.Namespace
	deploymentName := resourcenames.Deployment(vms)

	deployment, err := r.deploymentLister.Deployments(ns).Get(deploymentName)
	if apierrs.IsNotFound(err) {
		deployment = resources.MakeDeployment(ctx, vms, r.adapterImage)
		deployment, err = r.kubeclient.AppsV1().Deployments(ns).Create(deployment)
		if err != nil {
			return fmt.Errorf("failed to create deployment %q: %w", deploymentName, err)
		}
		logging.FromContext(ctx).Infof("Created deployment %q", deploymentName)
	} else if err != nil {
		return fmt.Errorf("failed to get deployment %q: %w", deploymentName, err)
	} else {
		// The deployment exists, but make sure that it has the shape that we expect.
		desiredDeployment := resources.MakeDeployment(ctx, vms, r.adapterImage)
		deployment = deployment.DeepCopy()
		deployment.Spec = desiredDeployment.Spec
		deployment, err = r.kubeclient.AppsV1().Deployments(ns).Update(deployment)
		if err != nil {
			return fmt.Errorf("failed to create deployment %q: %w", deploymentName, err)
		}
	}

	// Reflect the state of the Adapter Deployment in the VSphereSource
	vms.Status.PropagateAdapterStatus(deployment.Status)

	return nil
}
