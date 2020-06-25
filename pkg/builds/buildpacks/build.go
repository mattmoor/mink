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

package buildpacks

import (
	"context"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	tknv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	resources "github.com/tektoncd/pipeline/pkg/apis/resource/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/ptr"
)

const (
	// Paketo
	BuildpackImage = "gcr.io/paketo-buildpacks/builder:base"

	// Riff (doesn't work?)
	// BuildpackImage = "projectriff/builder"

	// GCP
	// BuildpackImage = "gcr.io/buildpacks/builder"
)

type Options struct {
	Builder string
}

func Build(ctx context.Context, kontext name.Reference, target name.Tag, opt Options) *tknv1beta1.TaskRun {
	volumes := []corev1.Volume{{
		Name: "empty-dir",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}, {
		Name: "layers-dir",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}}

	volumeMounts := []corev1.VolumeMount{{
		Name:      "layers-dir",
		MountPath: "/layers",
	}, {
		Name:      "empty-dir",
		MountPath: "/cache",
	}}

	return &tknv1beta1.TaskRun{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "buildpack-",
		},
		Spec: tknv1beta1.TaskRunSpec{
			PodTemplate: &v1beta1.PodTemplate{
				EnableServiceLinks: ptr.Bool(false),
			},

			// Out only resource is the output image.
			Resources: &tknv1beta1.TaskRunResources{
				Outputs: []tknv1beta1.TaskResourceBinding{{
					PipelineResourceBinding: tknv1beta1.PipelineResourceBinding{
						Name: "image",
						ResourceSpec: &resources.PipelineResourceSpec{
							Type: "image",
							Params: []resources.ResourceParam{{
								Name:  "url",
								Value: target.Name(),
							}},
						},
					},
				}},
			},

			TaskSpec: &tknv1beta1.TaskSpec{
				Resources: &tknv1beta1.TaskResources{
					Outputs: []tknv1beta1.TaskResource{{
						ResourceDeclaration: tknv1beta1.ResourceDeclaration{
							Name: "image",
							Type: "image",
						},
					}},
				},

				Results: []tknv1beta1.TaskResult{{
					Name: "IMAGE-DIGEST",
				}},

				Volumes: volumes,

				Steps: []tknv1beta1.Step{{
					Container: corev1.Container{
						Name:       "extract-bundle",
						Image:      kontext.String(),
						WorkingDir: "/workspace",
					},
				}, {
					Container: corev1.Container{
						Name:    "prepare",
						Image:   "alpine",
						Command: []string{"/bin/sh"},
						Args: []string{
							"-c",
							strings.Join([]string{
								`chown -R "1000:1000" "/tekton/home"`,
								`chown -R "1000:1000" "/layers"`,
								`chown -R "1000:1000" "/cache"`,
								`chown -R "1000:1000" "/workspace"`,
							}, " && "),
						},
						VolumeMounts: volumeMounts,
					},
				}, {
					Container: corev1.Container{
						Name:    "detect",
						Image:   opt.Builder,
						Command: []string{"/cnb/lifecycle/detector"},
						Args: []string{
							"-app=/workspace",
							"-group=/layers/group.toml",
							"-plan=/layers/plan.toml",
						},
						VolumeMounts: volumeMounts,
					},
				}, {
					Container: corev1.Container{
						Name:    "analyze",
						Image:   opt.Builder,
						Command: []string{"/cnb/lifecycle/analyzer"},
						Args: []string{
							"-layers=/layers",
							"-group=/layers/group.toml",
							"-cache-dir=/cache",
							"$(resources.outputs.image.url)",
						},
						VolumeMounts: volumeMounts,
					},
				}, {
					Container: corev1.Container{
						Name:    "restore",
						Image:   opt.Builder,
						Command: []string{"/cnb/lifecycle/restorer"},
						Args: []string{
							"-group=/layers/group.toml",
							"-layers=/layers",
							"-cache-dir=/cache",
						},
						VolumeMounts: volumeMounts,
					},
				}, {
					Container: corev1.Container{
						Name:    "build",
						Image:   opt.Builder,
						Command: []string{"/cnb/lifecycle/builder"},
						Args: []string{
							"-app=/workspace",
							"-layers=/layers",
							"-group=/layers/group.toml",
							"-plan=/layers/plan.toml",
						},
						VolumeMounts: volumeMounts,
					},
				}, {
					Container: corev1.Container{
						Name:    "export",
						Image:   opt.Builder,
						Command: []string{"/cnb/lifecycle/exporter"},
						Args: []string{
							"-app=/workspace",
							"-layers=/layers",
							"-group=/layers/group.toml",
							"-cache-dir=/cache",
							"$(resources.outputs.image.url)",
						},
						VolumeMounts: volumeMounts,
					},
				}, {
					// TODO(mattmoor): Replace with https://github.com/buildpacks/rfcs/pull/70
					Container: corev1.Container{
						Name:    "emit-digest",
						Image:   "gcr.io/go-containerregistry/crane:debug",
						Command: []string{"sh"},
						Args: []string{
							"-c",
							"crane digest $(resources.outputs.image.url) > /tekton/results/IMAGE-DIGEST",
						},
					},
				}},
			},
		},
	}
}
