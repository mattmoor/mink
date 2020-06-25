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
	"path/filepath"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	tknv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	resources "github.com/tektoncd/pipeline/pkg/apis/resource/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/ptr"
)

const (
	KanikoImage = "gcr.io/kaniko-project/executor:latest"
)

type Options struct {
	Dockerfile string
}

func Build(ctx context.Context, kontext name.Reference, target name.Tag, opt Options) *tknv1beta1.TaskRun {
	return &tknv1beta1.TaskRun{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "dockerfile-",
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

				Steps: []tknv1beta1.Step{{
					Container: corev1.Container{
						Name:       "extract-bundle",
						Image:      kontext.String(),
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
						Command: []string{
							"/kaniko/executor",

							"--dockerfile=" + filepath.Join("/workspace", opt.Dockerfile),

							// We expand into /workspace, and publish to the specified
							// output resource image.
							"--context=/workspace",
							"--destination=$(resources.outputs.image.url)",

							// Write out the digest to the appropriate result file.
							"--digest-file=/tekton/results/IMAGE-DIGEST",

							// Enable kanikache to get incremental builds
							"--cache=true",
							"--cache-ttl=24h",
						},
					},
				}},
			},
		},
	}
}
