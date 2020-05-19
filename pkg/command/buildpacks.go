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

package command

import (
	"context"
	"fmt"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/mattmoor/mink/pkg/builds/buildpacks"
	"github.com/spf13/cobra"
	tknv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
)

var buildpacks_example = fmt.Sprintf(`
  # Create a build from the context in the current directory and publish
  # it as the provided image name.
  %[1]s buildpack --image docker.io/mattmoor/bundle:latest

  # As the first, but executes the build as the service account "builder".
  %[1]s buildpack --as=builder --image docker.io/mattmoor/bundle:latest

  # As the first, but uses the context from the provided directory.
  %[1]s buildpack --directory=subdir/ --image docker.io/mattmoor/bundle:latest

  # As the first, but uses a different builder image.
  %[1]s buildpack --builder=cloudfoundry/cnb:bionic --image docker.io/mattmoor/bundle:latest

  # As the first, but executes the build as a temporary ServiceAccount
  # that is configured with the user's local credentials.
  # WARNING: This temporarily places your registry credentials in a Secret
  # on your cluster, so use this option with caution in shared environments.
  %[1]s buildpack --as=me --image docker.io/mattmoor/bundle:latest`, ExamplePrefix())

// NewBuildpackCommand implements 'kn-im buildpack' command
func NewBuildpackCommand() *cobra.Command {
	var opt buildpacks.Options

	cmd := makeBuildCommand(properties{
		cmd:     "buildpack",
		short:   "Build an image using a Cloud Native Buildpack.",
		example: buildpacks_example,
	}, func(ctx context.Context, kontext name.Reference, target name.Tag) *tknv1beta1.TaskRun {
		return buildpacks.Build(ctx, kontext, target, opt)
	})

	// Allow the user to override the builder image
	cmd.Flags().StringVarP(&opt.Builder, "builder", "b", buildpacks.BuildpackImage,
		"The name of the builder container image to execute.")

	return cmd
}
