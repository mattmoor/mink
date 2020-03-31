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

package v1alpha1

import (
	"context"
	"fmt"

	"github.com/mattmoor/vmware-sources/pkg/vsphere"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/tracker"
)

var vsbCondSet = apis.NewLivingConditionSet()

// GetGroupVersionKind returns the GroupVersionKind.
func (s *VSphereBinding) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("VSphereBinding")
}

// GetUntypedSpec implements apis.HasSpec
func (s *VSphereBinding) GetUntypedSpec() interface{} {
	return s.Spec
}

// GetSubject implements psbinding.Bindable
func (sb *VSphereBinding) GetSubject() tracker.Reference {
	return sb.Spec.Subject
}

// GetBindingStatus implements psbinding.Bindable
func (sb *VSphereBinding) GetBindingStatus() duck.BindableStatus {
	return &sb.Status
}

// SetObservedGeneration implements psbinding.BindableStatus
func (sbs *VSphereBindingStatus) SetObservedGeneration(gen int64) {
	sbs.ObservedGeneration = gen
}

// InitializeConditions populates the VSphereBindingStatus's conditions field
// with all of its conditions configured to Unknown.
func (sbs *VSphereBindingStatus) InitializeConditions() {
	vsbCondSet.Manage(sbs).InitializeConditions()
}

// MarkBindingUnavailable marks the VSphereBinding's Ready condition to False with
// the provided reason and message.
func (sbs *VSphereBindingStatus) MarkBindingUnavailable(reason, message string) {
	vsbCondSet.Manage(sbs).MarkFalse(VSphereBindingConditionReady, reason, message)
}

// MarkBindingAvailable marks the VSphereBinding's Ready condition to True.
func (sbs *VSphereBindingStatus) MarkBindingAvailable() {
	vsbCondSet.Manage(sbs).MarkTrue(VSphereBindingConditionReady)
}

// Do implements psbinding.Bindable
func (vsb *VSphereBinding) Do(ctx context.Context, ps *duckv1.WithPod) {
	// First undo so that we can just unconditionally append below.
	vsb.Undo(ctx, ps)

	// Make sure the PodSpec has a Volume like this:
	volume := corev1.Volume{
		Name: vsphere.VolumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: vsb.Spec.SecretRef.Name,
			},
		},
	}
	ps.Spec.Template.Spec.Volumes = append(ps.Spec.Template.Spec.Volumes, volume)

	// Make sure that each [init]container in the PodSpec has a VolumeMount like this:
	volumeMount := corev1.VolumeMount{
		Name:      vsphere.VolumeName,
		ReadOnly:  true,
		MountPath: vsphere.MountPath,
	}

	spec := ps.Spec.Template.Spec
	for i := range spec.InitContainers {
		spec.InitContainers[i].VolumeMounts = append(spec.InitContainers[i].VolumeMounts, volumeMount)
		spec.InitContainers[i].Env = append(spec.InitContainers[i].Env, corev1.EnvVar{
			Name:  "GOVC_URL",
			Value: vsb.Spec.Address.String(),
		}, corev1.EnvVar{
			Name:  "GOVC_INSECURE",
			Value: fmt.Sprintf("%v", vsb.Spec.SkipTLSVerify),
		}, corev1.EnvVar{
			Name: "GOVC_USERNAME",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: vsb.Spec.SecretRef.Name,
					},
					Key: corev1.BasicAuthUsernameKey,
				},
			},
		}, corev1.EnvVar{
			Name: "GOVC_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: vsb.Spec.SecretRef.Name,
					},
					Key: corev1.BasicAuthPasswordKey,
				},
			},
		})
	}
	for i := range spec.Containers {
		spec.Containers[i].VolumeMounts = append(spec.Containers[i].VolumeMounts, volumeMount)
		spec.Containers[i].Env = append(spec.Containers[i].Env, corev1.EnvVar{
			Name:  "GOVC_URL",
			Value: vsb.Spec.Address.String(),
		}, corev1.EnvVar{
			Name:  "GOVC_INSECURE",
			Value: fmt.Sprintf("%v", vsb.Spec.SkipTLSVerify),
		}, corev1.EnvVar{
			Name: "GOVC_USERNAME",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: vsb.Spec.SecretRef.Name,
					},
					Key: corev1.BasicAuthUsernameKey,
				},
			},
		}, corev1.EnvVar{
			Name: "GOVC_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: vsb.Spec.SecretRef.Name,
					},
					Key: corev1.BasicAuthPasswordKey,
				},
			},
		})
	}
}

func (vsb *VSphereBinding) Undo(ctx context.Context, ps *duckv1.WithPod) {
	spec := ps.Spec.Template.Spec

	for i, v := range spec.Volumes {
		if v.Name == vsphere.VolumeName {
			ps.Spec.Template.Spec.Volumes = append(spec.Volumes[:i], spec.Volumes[i+1:]...)
			break
		}
	}

	for i, c := range spec.InitContainers {
		for j, vm := range c.VolumeMounts {
			if vm.Name == vsphere.VolumeName {
				spec.InitContainers[i].VolumeMounts = append(c.VolumeMounts[:j], c.VolumeMounts[j+1:]...)
				break
			}
		}

		if len(c.Env) == 0 {
			continue
		}
		env := make([]corev1.EnvVar, 0, len(spec.InitContainers[i].Env))
		for j, ev := range c.Env {
			switch ev.Name {
			case "GOVC_URL", "GOVC_INSECURE", "GOVC_USERNAME", "GOVC_PASSWORD":
				continue
			default:
				env = append(env, spec.InitContainers[i].Env[j])
			}
		}
		spec.InitContainers[i].Env = env
	}
	for i, c := range spec.Containers {
		for j, vm := range c.VolumeMounts {
			if vm.Name == vsphere.VolumeName {
				spec.Containers[i].VolumeMounts = append(c.VolumeMounts[:j], c.VolumeMounts[j+1:]...)
				break
			}
		}

		if len(c.Env) == 0 {
			continue
		}
		env := make([]corev1.EnvVar, 0, len(spec.Containers[i].Env))
		for j, ev := range c.Env {
			switch ev.Name {
			case "GOVC_URL", "GOVC_INSECURE", "GOVC_USERNAME", "GOVC_PASSWORD":
				continue
			default:
				env = append(env, spec.Containers[i].Env[j])
			}
		}
		spec.Containers[i].Env = env
	}
}
