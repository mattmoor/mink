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

package ko

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	tknv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/ptr"
)

// Options holds configuration options specific to Dockerfile builds
type Options struct {
	// ImportPath is the path to ko publish
	ImportPath string
}

var (
	// KoImageString holds a reference to a built image of github.com/google/ko
	// See ./hack/build-flags.sh for how this is replaced at link-time.
	KoImageString = "docker.io/mattmoor/ko:latest"
)

// Build returns a TaskRun suitable for performing a "ko publish" build over the
// provided kontext and publishing to the target tag.
func Build(ctx context.Context, kontext name.Reference, target name.Tag, opt Options) *tknv1beta1.TaskRun {
	return &tknv1beta1.TaskRun{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "ko-publish-",
		},
		Spec: tknv1beta1.TaskRunSpec{
			PodTemplate: &tknv1beta1.PodTemplate{
				EnableServiceLinks: ptr.Bool(false),
			},

			TaskSpec: &tknv1beta1.TaskSpec{
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
						Name:  "ko-publish",
						Image: KoImageString,
						Env: []corev1.EnvVar{{
							Name:  "DOCKER_CONFIG",
							Value: "/tekton/home/.docker",
						}, {
							Name:  "KO_DOCKER_REPO",
							Value: target.Repository.String(),
						}},
						Command: []string{
							"/bin/bash", "-c",
						},
						Args: []string{
							strings.Join([]string{
								// Good for debugging.
								"go env",
								// Not set for some reason :rolls_eyes:
								"export GOARCH=$(go env GOARCH)",
								"export GOOS=$(go env GOOS)",
								"export GOARM=$(go env GOARM)",
								"export GOROOT=$(go env GOROOT)",
								// Where the magic happens.
								fmt.Sprintf("ko publish --bare %s | cut -d'@' -f 2 > /tekton/results/IMAGE-DIGEST", opt.ImportPath),
							}, " && "),
						},
						Resources: corev1.ResourceRequirements{
							// Set requests based on a typical ko task,
							// but do not set limits because it could
							// go well beyond this (in theory).
							Requests: corev1.ResourceList{
								// 1 is typical, but I've seen up to 2.
								"cpu": resource.MustParse("1"),
								// 500-700Mi was typical.
								"memory": resource.MustParse("1Gi"),
							},
						},
					},
				}},
			},
		},
	}
}
