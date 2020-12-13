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
	"math/rand"
	"path"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/mattmoor/mink/pkg/constants"
	tknv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/ptr"
)

const (
	// BuildpackImage defines the name of the default buildpack builder image
	// with which to execute the buildpack lifecycle.  We default to Paketo,
	// see other options below.
	//
	// For GCP see: http://github.com/GoogleCloudPlatform/buildpacks
	//  - gcr.io/buildpacks/builder
	//
	// For Boson see: https://github.com/boson-project/faas/blob/main/buildpacks/builder.go#L25
	//  - Quarkus: quay.io/boson/faas-quarkus-builder
	//  - Node.js: quay.io/boson/faas-nodejs-builder
	//  - Go:      quay.io/boson/faas-go-builder
	BuildpackImage = "docker.io/paketobuildpacks/builder:full"
)

var (
	// PlatformSetupImageString holds a reference to a built image of ./cmd/platform-setup
	// See ./hack/build-flags.sh for how this is replaced at link-time.
	PlatformSetupImageString = "docker.io/mattmoor/platform-setup:latest"
	// PlatformSetupImage is where we publish ./cmd/platform-setup
	PlatformSetupImage, _ = name.ParseReference(PlatformSetupImageString)

	// ExtractDigestImageString holds a reference to a built image of ./cmd/extract-digest
	// See ./hack/build-flags.sh for how this is replaced at link-time.
	ExtractDigestImageString = "docker.io/mattmoor/extract-digest:latest"
	// ExtractDigestImage is where we publish ./cmd/extract-digest
	ExtractDigestImage, _ = name.ParseReference(ExtractDigestImageString)
)

// Options are the options for executing a buildpack build.
type Options struct {
	// Builder is the name of the builder image for which to apply the buildpack lifecycle.
	Builder string

	// OverrideFile is the name of the override.toml file (under Path)
	OverrideFile string
}

// Build synthesizes a TaskRun definition that evaluates the buildpack lifecycle with the
// given options over the provided source.
func Build(ctx context.Context, source name.Reference, target name.Tag, opt Options) *tknv1beta1.TaskRun {
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

	//nolint:gosec // crypto rand is not needed.
	workspaceDirectory := fmt.Sprint("/workspace/", rand.Uint64())
	user, group := int64(1000), int64(1000)

	return &tknv1beta1.TaskRun{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "buildpack-",
		},
		Spec: tknv1beta1.TaskRunSpec{
			PodTemplate: &tknv1beta1.PodTemplate{
				EnableServiceLinks: ptr.Bool(false),
			},

			Params: []tknv1beta1.Param{{
				Name:  constants.SourceBundleParam,
				Value: *tknv1beta1.NewArrayOrString(source.String()),
			}, {
				Name:  constants.ImageTargetParam,
				Value: *tknv1beta1.NewArrayOrString(target.String()),
			}},

			TaskSpec: &tknv1beta1.TaskSpec{
				Params: []tknv1beta1.ParamSpec{{
					Name:        constants.SourceBundleParam,
					Description: "A self-extracting container image of source",
				}, {
					Name:        constants.ImageTargetParam,
					Description: "Where to publish an image.",
				}},

				Results: []tknv1beta1.TaskResult{{
					Name: constants.ImageDigestResult,
				}},

				Volumes: []corev1.Volume{{
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
				}},

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
						Image:      "$(params." + constants.SourceBundleParam + ")",
						WorkingDir: workspaceDirectory,
						SecurityContext: &corev1.SecurityContext{
							RunAsUser:  &user,
							RunAsGroup: &group,
						},
					},
				}, {
					Container: corev1.Container{
						Name:  "platform-setup",
						Image: PlatformSetupImage.String(),
						Args: []string{
							"--overrides", filepath.Join(workspaceDirectory, opt.OverrideFile),
						},
						WorkingDir:   workspaceDirectory,
						VolumeMounts: volumeMounts,
					},
				}, {
					Container: corev1.Container{
						Name:       "create",
						Image:      opt.Builder,
						WorkingDir: workspaceDirectory,
						Command:    []string{"/cnb/lifecycle/creator"},
						Args: []string{
							"-layers=/layers",
							"-app=" + workspaceDirectory,
							"-cache-dir=/cache",
							"-platform=/platform",
							fmt.Sprint("-gid=", group),
							fmt.Sprint("-uid=", user),
							"$(params." + constants.ImageTargetParam + ")",
						},
						Env: []corev1.EnvVar{{
							Name:  "DOCKER_CONFIG",
							Value: "/tekton/home/.docker",
						}},
						VolumeMounts: volumeMounts,
						SecurityContext: &corev1.SecurityContext{
							RunAsUser:  &user,
							RunAsGroup: &group,
						},
					},
				}, {
					Container: corev1.Container{
						Name:       "extract-digest",
						Image:      ExtractDigestImage.String(),
						WorkingDir: workspaceDirectory,
						Args: []string{
							"-output", path.Join("/tekton/results", constants.ImageDigestResult),
						},
					},
				}},
			},
		},
	}
}
