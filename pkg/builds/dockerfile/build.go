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

	"github.com/ghodss/yaml"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/mattmoor/mink/pkg/constants"
	tknv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/ptr"
)

var (
	// KanikoTaskString holds the raw definition of the Kaniko task.
	// We export this into ./examples/kaniko.yaml
	KanikoTaskString = `
apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: kaniko
spec:
  description: "An example kaniko task illustrating some of the parameter processing."
  params:
    - name: mink-source-bundle
      description: A self-extracting container image of source
    - name: mink-image-target
      description: Where to publish an image.
    - name: path
      description: The path to the dockerfile.
      default: .
    - name: dockerfile
      description: The name of the dockerfile.
      default: Dockerfile
    - name: kaniko-args
      description: Extra arguments to supply to kaniko
      type: array
      default: []

  results:
    - name: mink-image-digest
      description: The digest of the resulting image.

  steps:
    - name: extract-bundle
      image: $(params.mink-source-bundle)
      workingDir: /workspace

    - name: build-and-push
      image: ghcr.io/mattmoor/kaniko:latest
      env:
      - name: DOCKER_CONFIG
        value: /tekton/home/.docker
      args:
      - --dockerfile=/workspace/$(params.path)/$(params.dockerfile)
      - --context=/workspace
      - --destination=$(params.mink-image-target)
      - --digest-file=/tekton/results/mink-image-digest
      - --cache=true
      - --cache-ttl=24h
      - $(params.kaniko-args)
`
	// KanikoTask is the parsed form of KanikoTaskString.
	KanikoTask tknv1beta1.Task
)

func init() {
	if err := yaml.Unmarshal([]byte(KanikoTaskString), &KanikoTask); err != nil {
		panic(err)
	}
}

// Options holds configuration options specific to Dockerfile builds
type Options struct {
	// Dockerfile is the path to the Dockerfile within the build context.
	Dockerfile string

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
			TaskSpec: KanikoTask.Spec.DeepCopy(),
			Params: []tknv1beta1.Param{{
				Name:  constants.SourceBundleParam,
				Value: *tknv1beta1.NewArrayOrString(source.String()),
			}, {
				Name:  constants.ImageTargetParam,
				Value: *tknv1beta1.NewArrayOrString(target.String()),
			}, {
				Name:  "dockerfile",
				Value: *tknv1beta1.NewArrayOrString(opt.Dockerfile),
			}, {
				Name: "kaniko-args",
				Value: tknv1beta1.ArrayOrString{
					Type:     tknv1beta1.ParamTypeArray,
					ArrayVal: opt.KanikoArgs,
				},
			}},
		},
	}
}
