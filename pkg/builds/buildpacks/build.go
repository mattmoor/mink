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

	"github.com/ghodss/yaml"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/mattmoor/mink/pkg/constants"
	tknv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
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
	PlatformSetupImageString = "ghcr.io/mattmoor/github.com/mattmoor/mink/cmd/platform-setup:latest"
	// PlatformSetupImage is where we publish ./cmd/platform-setup
	PlatformSetupImage, _ = name.ParseReference(PlatformSetupImageString)

	// ExtractDigestImageString holds a reference to a built image of ./cmd/extract-digest
	// See ./hack/build-flags.sh for how this is replaced at link-time.
	ExtractDigestImageString = "ghcr.io/mattmoor/github.com/mattmoor/mink/cmd/extract-digest:latest"
	// ExtractDigestImage is where we publish ./cmd/extract-digest
	ExtractDigestImage, _ = name.ParseReference(ExtractDigestImageString)

	// BuildpackTaskString holds the raw definition of the Buildpack task.
	// We export this into ./examples/buildpack.yaml
	BuildpackTaskString = `
apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: buildpack
spec:
  description: "An example buildpack task illustrating some of the parameter processing."
  params:
    - name: mink-source-bundle
      description: A self-extracting container image of source
    - name: mink-image-target
      description: Where to publish an image.

    - name: descriptor
      default: "./project.toml"
      description: |
        The path to the project descriptor relative to the source bundle.
        For more information: https://buildpacks.io/docs/app-developer-guide/using-project-descriptor/
    - name: builder
      default: "docker.io/paketobuildpacks/builder:full"
      description: |
        The image uri for the builder to execute.  Some example builders:

        Paketo (default): docker.io/paketobuildpacks/builder:full

        GCP: gcr.io/buildpacks/builder

        Boson (Quarkus): quay.io/boson/faas-quarkus-builder
        Boson (Node.js): quay.io/boson/faas-nodejs-builder
        Boson (Go): quay.io/boson/faas-go-builder

        For more information on builders, see: https://buildpacks.io/docs/concepts/components/builder/

    # TODO(mattmoor): There is not a good way to support integer substitutions in tekton,
    # so we cannot practically make user-id and group-id parameters.

  results:
    - name: mink-image-digest
      description: The digest of the resulting image.

  steps:
    - name: prepare
      image: gcr.io/distroless/base:debug
      workingDir: /workspace
      command: ["sh"]
      args:
        - "-c"
        - |-
          chown -R "1000:1000" "/tekton/home" &&
          chown -R "1000:1000" "/layers" &&
          chown -R "1000:1000" "/cache" &&
          chown -R "1000:1000" "/workspace"
      volumeMounts: &mounts
        - name: layers-dir
          mountPath: /layers
        - name: empty-dir
          mountPath: /cache
        - name: platform-dir
          mountPath: /platform

    - name: extract-bundle
      image: $(params.mink-source-bundle)
      workingDir: /workspace
      securityContext: &run-as
        runAsUser: 1000
        runAsGroup: 1000

    - name: platform-setup
      image: ko://github.com/mattmoor/mink/cmd/platform-setup
      workingDir: /workspace
      args: ["--descriptor=/workspace/$(params.descriptor)"]
      volumeMounts: *mounts

    - name: create
      image: $(params.builder)
      workingDir: /workspace
      imagePullPolicy: Always
      command: ["/cnb/lifecycle/creator"]
      args:
        - "-layers=/layers"
        - "-app=/workspace"
        - "-cache-dir=/cache"
        - "-platform=/platform"
        - "-uid=1000"
        - "-gid=1000"
        - "$(params.mink-image-target)"
      env:
      - name: DOCKER_CONFIG
        value: /tekton/home/.docker
      volumeMounts: *mounts
      securityContext: *run-as

    - name: extract-digest
      image: ko://github.com/mattmoor/mink/cmd/extract-digest
      workingDir: /workspace
      args: ["-output=/tekton/results/mink-image-digest"]
      volumeMounts: *mounts

  volumes:
    - name: platform-dir
      emptyDir: {}
    - name: layers-dir
      emptyDir: {}
    - name: empty-dir
      emptyDir: {}
`

	// BuildpackTask is the parsed form of BuildpackTaskString.
	BuildpackTask tknv1beta1.Task
)

func init() {
	// Replace the ko strings we use for the sample with the values the build process has injected.
	substitutions := map[string]string{
		"ko://github.com/mattmoor/mink/cmd/platform-setup": PlatformSetupImageString,
		"ko://github.com/mattmoor/mink/cmd/extract-digest": ExtractDigestImageString,
	}

	raw := BuildpackTaskString
	for k, v := range substitutions {
		raw = strings.ReplaceAll(raw, k, v)
	}

	if err := yaml.Unmarshal([]byte(raw), &BuildpackTask); err != nil {
		panic(err)
	}
}

// Options are the options for executing a buildpack build.
type Options struct {
	// Builder is the name of the builder image for which to apply the buildpack lifecycle.
	Builder string

	// DescriptorFile holds the name of the project descriptor file (aka project.toml).
	DescriptorFile string
}

// Build synthesizes a TaskRun definition that evaluates the buildpack lifecycle with the
// given options over the provided source.
func Build(ctx context.Context, source name.Reference, target name.Tag, opt Options) *tknv1beta1.TaskRun {
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
			}, {
				Name:  "builder",
				Value: *tknv1beta1.NewArrayOrString(opt.Builder),
			}, {
				Name:  "descriptor",
				Value: *tknv1beta1.NewArrayOrString(opt.DescriptorFile),
			}},

			TaskSpec: BuildpackTask.Spec.DeepCopy(),
		},
	}
}
