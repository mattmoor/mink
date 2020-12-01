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
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
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

	// ExtractDigestImageString holds a reference to a built image of ./cmd/platform-setup
	// See ./hack/build-flags.sh for how this is replaced at link-time.
	ExtractDigestImageString = "docker.io/mattmoor/extract-digest:latest"
	// ExtractDigestImage is where we publish ./cmd/platform-setup
	ExtractDigestImage, _ = name.ParseReference(ExtractDigestImageString)
)

// Options are the options for executing a buildpack build.
type Options struct {
	// Builder is the name of the builder image for which to apply the buildpack lifecycle.
	Builder string

	// OverrideFile is the name of the override.toml file (under Path)
	OverrideFile string

	// Env is additional environment variables to pass to the build.
	Env []corev1.EnvVar
}

// Build synthesizes a TaskRun definition that evaluates the buildpack lifecycle with the
// given options over the provided source.
func Build(ctx context.Context, source name.Reference, target name.Tag, opt Options) *tknv1beta1.TaskRun {
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

	//nolint:gosec // crypto rand is not needed.
	workspaceDirectory := fmt.Sprint("/workspace/", rand.Uint64())
	user, group := determineUserAndGroup(opt.Builder)

	pfSetupArgs := make([]string, 0, 2*(len(opt.Env)+1))
	pfSetupArgs = append(pfSetupArgs,
		"--overrides", filepath.Join(workspaceDirectory, opt.OverrideFile),
	)
	for _, ev := range opt.Env {
		pfSetupArgs = append(pfSetupArgs,
			"--env", ev.Name+"="+ev.Value,
		)
	}

	return &tknv1beta1.TaskRun{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "buildpack-",
		},
		Spec: tknv1beta1.TaskRunSpec{
			PodTemplate: &tknv1beta1.PodTemplate{
				EnableServiceLinks: ptr.Bool(false),
			},

			TaskSpec: &tknv1beta1.TaskSpec{
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
						Image:      source.String(),
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
						Args:         pfSetupArgs,
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
							"-analyzed=/layers/analyze.toml",
							"-cache-dir=/cache",
							target.Name(),
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
							"-analyzed=/layers/analyze.toml",
							"-cache-dir=/cache",
							target.Name(),
						},
						Env: []corev1.EnvVar{{
							Name:  "DOCKER_CONFIG",
							Value: "/tekton/home/.docker",
						}},
						VolumeMounts: volumeMounts,
					},
				}, {
					Container: corev1.Container{
						Name:       "extract-digest",
						Image:      ExtractDigestImage.String(),
						WorkingDir: workspaceDirectory,
						Args:       []string{"-output=/tekton/results/IMAGE-DIGEST"},
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
	}
	uid = int64(user)
	group, err := strconv.Atoi(parts[1])
	if err != nil {
		log.Print("Error parsing gid: ", parts[1])
		return
	}
	gid = int64(group)
	return
}
