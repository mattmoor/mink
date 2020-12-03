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

package dockerfile

import (
	"context"
	"path"
	"path/filepath"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/mattmoor/mink/pkg/constants"
	tknv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/ptr"
)

const (
	// KanikoImage is the path to the kaniko image we use for Dockerfile builds.
	KanikoImage = "gcr.io/kaniko-project/executor:multi-arch"
)

// Options holds configuration options specific to Dockerfile builds
type Options struct {
	// Dockerfile is the path to the Dockerfile within the build context.
	Dockerfile string

	// The path within the build context in which to execute the build.
	Path string

	// The extra kaniko arguments for handling things like insecure registries
	KanikoArgs []string
}

// Build returns a TaskRun suitable for performing a Dockerfile build over the
// provided source and publishing to the target tag.
func Build(ctx context.Context, source name.Reference, target name.Tag, opt Options) *tknv1beta1.TaskRun {
	return &tknv1beta1.TaskRun{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "dockerfile-",
		},
		Spec: tknv1beta1.TaskRunSpec{
			PodTemplate: &tknv1beta1.PodTemplate{
				EnableServiceLinks: ptr.Bool(false),
			},

			TaskSpec: &tknv1beta1.TaskSpec{
				Results: []tknv1beta1.TaskResult{{
					Name: constants.ImageDigestResult,
				}},

				Steps: []tknv1beta1.Step{{
					Container: corev1.Container{
						Name:       "extract-bundle",
						Image:      source.String(),
						WorkingDir: "/workspace",
					},
				}, {
					Container: corev1.Container{
						Name:  "build-and-push",
						Image: KanikoImage,
						Env: []corev1.EnvVar{{
							Name:  "DOCKER_CONFIG",
							Value: "/tekton/home/.docker",
						}},
						Args: append([]string{
							"--dockerfile=" + filepath.Join("/workspace", opt.Path, opt.Dockerfile),

							// We expand into /workspace, and publish to the specified
							// output resource image.
							"--context=" + filepath.Join("/workspace", opt.Path),
							"--destination=" + target.Name(),

							// Write out the digest to the appropriate result file.
							"--digest-file", path.Join("/tekton/results", constants.ImageDigestResult),

							// Enable kanikache to get incremental builds
							"--cache=true",
							"--cache-ttl=24h",
						}, opt.KanikoArgs...),
					},
				}},
			},
		},
	}
}
