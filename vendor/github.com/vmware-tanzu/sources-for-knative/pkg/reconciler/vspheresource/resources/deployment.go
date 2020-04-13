/*
Copyright 2020 VMware, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package resources

import (
	"context"
	"encoding/json"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/ptr"

	"github.com/vmware-tanzu/sources-for-knative/pkg/apis/sources/v1alpha1"
	"github.com/vmware-tanzu/sources-for-knative/pkg/reconciler/vspheresource/resources/names"
)

func MakeDeployment(ctx context.Context, vms *v1alpha1.VSphereSource, adapterImage string) *appsv1.Deployment {
	labels := map[string]string{
		"vspheresources.sources.tanzu.vmware.com/name": vms.Name,
	}

	var ceOverrides string
	if vms.Spec.CloudEventOverrides != nil {
		if co, err := json.Marshal(vms.Spec.SourceSpec.CloudEventOverrides); err != nil {
			logging.FromContext(ctx).Errorf(
				"Failed to marshal CloudEventOverrides into JSON for %+v, %v", vms, err)
		} else if len(co) > 0 {
			ceOverrides = string(co)
		}
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            names.Deployment(vms),
			Namespace:       vms.Namespace,
			OwnerReferences: []metav1.OwnerReference{*kmeta.NewControllerRef(vms)},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.Int32(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: names.ServiceAccount(vms),
					Containers: []corev1.Container{{
						Name:  "adapter",
						Image: adapterImage,
						Env: []corev1.EnvVar{{
							Name: "NAMESPACE",
							ValueFrom: &corev1.EnvVarSource{
								FieldRef: &corev1.ObjectFieldSelector{
									FieldPath: "metadata.namespace",
								},
							},
						}, {
							Name: "NAME",
							ValueFrom: &corev1.EnvVarSource{
								FieldRef: &corev1.ObjectFieldSelector{
									FieldPath: "metadata.name",
								},
							},
						}, {
							Name:  "K_METRICS_CONFIG",
							Value: `{"Domain":"tanzu.vmware.com/sources","Component":"source"}`,
						}, {
							Name:  "K_LOGGING_CONFIG",
							Value: "{}",
						}, {
							Name:  "VSPHERE_KVSTORE_CONFIGMAP",
							Value: names.ConfigMap(vms),
						}, {
							Name:  "K_CE_OVERRIDES",
							Value: ceOverrides,
						}, {
							Name:  "K_SINK",
							Value: vms.Status.SinkURI.String(),
						}},
					}},
				},
			},
		},
	}
}
