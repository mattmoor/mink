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
	"io"
	"net/url"
	"path/filepath"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/mattmoor/mink/pkg/builds"
	"github.com/mattmoor/mink/pkg/builds/dockerfile"
	minkcli "github.com/mattmoor/mink/pkg/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

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
func NewBuildCommand(ctx context.Context) *cobra.Command {
	opts := &BuildOptions{
		BaseBuildOptions: BaseBuildOptions{BundleOptions: BundleOptions{ctx: ctx}},
	}

	cmd := &cobra.Command{
		Use:     "build --image IMAGE",
		Short:   "Build an image from a Dockerfile.",
		Example: dockerfileExample,
		PreRunE: opts.Validate,
		RunE:    opts.Execute,
	}

	opts.AddFlags(cmd)

	return cmd
}

type dockerfileOptions struct {
	// Dockerfile is the relative path to the Dockerfile within the build context.
	Dockerfile string

	// The extra kaniko arguments for handling things like insecure registries
	KanikoArgs []string
}

// AddFlags implements Interface
func (opts *dockerfileOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().String("dockerfile", "Dockerfile", "The path to the Dockerfile within the build context.")
	cmd.Flags().StringSlice("kaniko-args", nil, "Optional arguments to pass to kaniko for dealing with insecure registries. For details see: https://github.com/GoogleContainerTools/kaniko/blob/master/README.md#additional-flags")
}

// Validate implements Interface
func (opts *dockerfileOptions) Validate(cmd *cobra.Command, args []string) error {
	opts.Dockerfile = viper.GetString("dockerfile")
	if opts.Dockerfile == "" {
		return minkcli.ErrMissingFlag("dockerfile")
	}

	opts.KanikoArgs = viper.GetStringSlice("kaniko-args")
	return nil
}

// BuildOptions implements Interface for the `kn im build` command.
type BuildOptions struct {
	// Inherit all of the base build options.
	BaseBuildOptions

	// Inherit the dockerfile options.
	dockerfileOptions
}

// BuildOptions implements Interface
var _ Interface = (*BuildOptions)(nil)

// AddFlags implements Interface
func (opts *BuildOptions) AddFlags(cmd *cobra.Command) {
	// Add the bundle flags to our surface.
	opts.BaseBuildOptions.AddFlags(cmd)

	opts.dockerfileOptions.AddFlags(cmd)
}

// Validate implements Interface
func (opts *BuildOptions) Validate(cmd *cobra.Command, args []string) error {
	// Validate the bundle arguments.
	if err := opts.BaseBuildOptions.Validate(cmd, args); err != nil {
		return err
	}
	return opts.dockerfileOptions.Validate(cmd, args)
}

// Execute implements Interface
func (opts *BuildOptions) Execute(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return errors.New("'im bundle' does not take any arguments")
	}

	// Handle ctrl+C
	ctx := opts.GetContext(cmd)

	// Bundle up the source context in an image.
	sourceDigest, err := opts.bundle(ctx)
	if err != nil {
		return err
	}

	// Run the produced Build definition to completion, streaming logs to stdout, and
	// returning the digest of the produced image.
	digest, err := opts.build(ctx, sourceDigest, cmd.OutOrStderr())
	if err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "%s\n", digest.String())
	return nil
}

func (opts *BuildOptions) build(ctx context.Context, sourceDigest name.Digest, w io.Writer) (name.Digest, error) {
	tag, err := opts.tag(imageNameContext{
		URL: url.URL{
			Scheme: "dockerfile",
			Path:   filepath.Clean(filepath.Dir(opts.Dockerfile)),
		},
	})
	if err != nil {
		return name.Digest{}, err
	}

	// Create a Build definition for turning the source into an image by Dockerfile build.
	tr := dockerfile.Build(ctx, sourceDigest, tag, dockerfile.Options{
		Dockerfile: opts.Dockerfile,
		KanikoArgs: opts.KanikoArgs,
	})
	tr.Namespace = Namespace()

	// Run the produced Build definition to completion, streaming logs to stdout, and
	// returning the digest of the produced image.
	return builds.Run(ctx, tag.String(), tr, &options.LogOptions{
		Params: &cli.TektonParams{},
		Stream: &cli.Stream{
			// Send Out to stderr so we can capture the digest for composition.
			Out: w,
			Err: w,
		},
		Follow: true,
	}, builds.WithTaskServiceAccount(ctx, opts.ServiceAccount, tag, sourceDigest))
}
