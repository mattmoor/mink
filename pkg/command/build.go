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
	"errors"
	"fmt"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/mattmoor/mink/pkg/builds"
	"github.com/mattmoor/mink/pkg/builds/dockerfile"
	"github.com/mattmoor/mink/pkg/kontext"
	"github.com/spf13/cobra"
	tknv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"knative.dev/pkg/signals"

	"github.com/tektoncd/cli/pkg/cli"
	"github.com/tektoncd/cli/pkg/options"
)

var dockerfileExample = fmt.Sprintf(`
  # Create a build from the context in the current directory and publish
  # it as the provided image name.
  %[1]s build --image docker.io/mattmoor/bundle:latest

  # As the first, but executes the build as the service account "builder"
  %[1]s build --as=builder --image docker.io/mattmoor/bundle:latest

  # As the first, but uses the context from the provided directory
  %[1]s build --directory=subdir/ --image docker.io/mattmoor/bundle:latest

  # As the first, but builds ./app/Dockerfile.production.
  %[1]s build --dockerfile=./app/Dockerfile.production --image docker.io/mattmoor/bundle:latest

  # As the first, but executes the build as a temporary ServiceAccount
  # that is configured with the user's local credentials.
  # WARNING: This temporarily places your registry credentials in a Secret
  # on your cluster, so use this option with caution in shared environments.
  %[1]s build --as=me --image docker.io/mattmoor/bundle:latest`, ExamplePrefix())

// NewBuildCommand implements 'kn-im build' command
func NewBuildCommand() *cobra.Command {
	var opt dockerfile.Options

	cmd := makeBuildCommand(properties{
		cmd:     "build",
		short:   "Build an image from a Dockerfile.",
		example: dockerfileExample,
	}, func(ctx context.Context, kontext name.Reference, target name.Tag) *tknv1beta1.TaskRun {
		return dockerfile.Build(ctx, kontext, target, opt)
	})

	// Allow the user to override the path to the Dockerfile?
	cmd.Flags().StringVarP(&opt.Dockerfile, "dockerfile", "", "Dockerfile",
		"The path to the Dockerfile within the build context.")

	return cmd
}

type properties struct {
	cmd     string
	short   string
	example string
}

func makeBuildCommand(props properties, fn func(context.Context, name.Reference, name.Tag) *tknv1beta1.TaskRun) *cobra.Command {
	var directory string
	var image string
	var serviceaccount string

	cmd := &cobra.Command{
		Use:     fmt.Sprintf("%s --image IMAGE", props.cmd),
		Short:   props.short,
		Example: props.example,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return errors.New("'im bundle' does not take any arguments")
			}

			// Handle ctrl+C
			ctx := signals.NewContext()

			buildTag, err := name.NewTag(image, name.WeakValidation)
			if err != nil {
				return err
			}
			sourceTag, err := name.NewTag(buildTag.Name()+"-source", name.WeakValidation)
			if err != nil {
				return err
			}
			// Bundle up the source context in an image.
			sourceDigest, err := kontext.Bundle(ctx, directory, sourceTag)
			if err != nil {
				return err
			}

			// Create a Build definition for turning the source into an image by Dockerfile build.
			tr := fn(ctx, sourceDigest, buildTag)
			tr.Namespace = Namespace()

			// Run the produced Build definition to completion, streaming logs to stdout, and
			// returning the digest of the produced image.
			digest, err := builds.Run(ctx, image, tr, &options.LogOptions{
				Params: &cli.TektonParams{},
				Stream: &cli.Stream{
					// Send Out to stderr so we can capture the digest for composition.
					Out: cmd.OutOrStderr(),
					Err: cmd.OutOrStderr(),
				},
				Follow: true,
			}, builds.WithServiceAccount(serviceaccount, buildTag))
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%s\n", digest.String())
			return nil
		},
	}

	cmd.Flags().StringVarP(&image, "image", "i", "", "Where to publish the build.")
	cmd.MarkFlagRequired("image")

	cmd.Flags().StringVarP(&directory, "directory", "d", ".", "The directory to bundle up.")

	cmd.Flags().StringVarP(&serviceaccount, "as", "", "default",
		"The name of the ServiceAccount as which to run the build, pass --as=me to "+
			"temporarily create a new ServiceAccount to push with your local credentials.")

	return cmd
}
