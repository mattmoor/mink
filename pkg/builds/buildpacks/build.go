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
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	tknv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	resources "github.com/tektoncd/pipeline/pkg/apis/resource/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/ptr"
)

const (
	// Paketo
	BuildpackImage = "docker.io/paketobuildpacks/builder:full"

	// GCP
	// BuildpackImage = "gcr.io/buildpacks/builder"

	// Boson
	// see: https://github.com/boson-project/faas/blob/main/buildpacks/builder.go#L25
	// BuildpackImage = "quay.io/boson/faas-quarkus-builder"
	// BuildpackImage = "quay.io/boson/faas-nodejs-builder"
	// BuildpackImage = "quay.io/boson/faas-go-builder"
)

var (
	PlatformSetupImageString = "docker.io/mattmoor/platform-setup:latest"
	// BaseImage is where we publish ./cmd/platform-setup
	PlatformSetupImage, _ = name.ParseReference(PlatformSetupImageString)
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
	}, {
		Name: "platform-dir",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}}

	volumeMounts := []corev1.VolumeMount{{
		Name:      "platform-dir",
		MountPath: "/platform",
	}, {
		Name:      "layers-dir",
		MountPath: "/layers",
	}, {
		Name:      "empty-dir",
		MountPath: "/cache",
	}}

	workspaceDirectory := fmt.Sprint("/workspace/", rand.Uint64())
	user, group := determineUserAndGroup(opt.Builder)

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
						Name:       "prepare",
						Image:      "alpine",
						WorkingDir: workspaceDirectory,
						Command:    []string{"/bin/sh"},
						Args: []string{
							"-c",
							fmt.Sprintf(strings.Join([]string{
								`chown -R "%[1]d:%[2]d" "/tekton/home"`,
								`chown -R "%[1]d:%[2]d" "/layers"`,
								`chown -R "%[1]d:%[2]d" "/cache"`,
								`chown -R "%[1]d:%[2]d" "/workspace"`,
							}, " && "), user, group),
						},
						VolumeMounts: volumeMounts,
					},
				}, {
					Container: corev1.Container{
						Name:       "extract-bundle",
						Image:      kontext.String(),
						WorkingDir: workspaceDirectory,
						SecurityContext: &corev1.SecurityContext{
							RunAsUser:  &user,
							RunAsGroup: &group,
						},
					},
				}, {
					Container: corev1.Container{
						Name:         "platform-setup",
						Image:        PlatformSetupImage.String(),
						WorkingDir:   workspaceDirectory,
						VolumeMounts: volumeMounts,
					},
				}, {
					Container: corev1.Container{
						Name:       "detect",
						Image:      opt.Builder,
						WorkingDir: workspaceDirectory,
						Command:    []string{"/cnb/lifecycle/detector"},
						Args: []string{
							"-app=" + workspaceDirectory,
							"-group=/layers/group.toml",
							"-plan=/layers/plan.toml",
						},
						VolumeMounts: volumeMounts,
					},
				}, {
					Container: corev1.Container{
						Name:       "analyze",
						Image:      opt.Builder,
						WorkingDir: workspaceDirectory,
						Command:    []string{"/cnb/lifecycle/analyzer"},
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
						Name:       "restore",
						Image:      opt.Builder,
						WorkingDir: workspaceDirectory,
						Command:    []string{"/cnb/lifecycle/restorer"},
						Args: []string{
							"-group=/layers/group.toml",
							"-layers=/layers",
							"-cache-dir=/cache",
						},
						VolumeMounts: volumeMounts,
					},
				}, {
					Container: corev1.Container{
						Name:       "build",
						Image:      opt.Builder,
						WorkingDir: workspaceDirectory,
						Command:    []string{"/cnb/lifecycle/builder"},
						Args: []string{
							"-app=" + workspaceDirectory,
							"-layers=/layers",
							"-group=/layers/group.toml",
							"-plan=/layers/plan.toml",
						},
						VolumeMounts: volumeMounts,
					},
				}, {
					Container: corev1.Container{
						Name:       "export",
						Image:      opt.Builder,
						WorkingDir: workspaceDirectory,
						Command:    []string{"/cnb/lifecycle/exporter"},
						Args: []string{
							"-app=" + workspaceDirectory,
							"-layers=/layers",
							"-group=/layers/group.toml",
							"-cache-dir=/cache",
							"$(resources.outputs.image.url)",
						},
						Env: []corev1.EnvVar{{
							Name:  "DOCKER_CONFIG",
							Value: "/tekton/home/.docker",
						}},
						VolumeMounts: volumeMounts,
					},
				}, {
					// TODO(mattmoor): Replace with https://github.com/buildpacks/rfcs/pull/70
					Container: corev1.Container{
						Name:       "emit-digest",
						Image:      "gcr.io/go-containerregistry/crane:debug",
						WorkingDir: workspaceDirectory,
						Command:    []string{"sh"},
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

func determineUserAndGroup(builder string) (uid int64, gid int64) {
	uid, gid = 1000, 1000
	ref, err := name.ParseReference(builder)
	if err != nil {
		log.Print("Error parsing reference: ", err)
		return
	}
	img, err := remote.Image(ref)
	if err != nil {
		log.Print("Error parsing reference: ", err)
		return
	}
	cfg, err := img.ConfigFile()
	if err != nil {
		log.Print("Error reading config file: ", err)
		return
	}
	parts := strings.SplitN(cfg.Config.User, ":", 2)
	if len(parts) < 2 {
		parts = append(parts, parts[0])
	}
	user, err := strconv.Atoi(parts[0])
	if err != nil {
		log.Print("Error parsing uid: ", parts[0])
		return
	} else {
		uid = int64(user)
	}
	group, err := strconv.Atoi(parts[1])
	if err != nil {
		log.Print("Error parsing gid: ", parts[1])
		return
	} else {
		gid = int64(group)
	}
	return
}
