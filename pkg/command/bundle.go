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
	"errors"
	"fmt"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/mattmoor/mink/pkg/kontext"
	"github.com/mattmoor/mink/pkg/source"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/signals"
)

// BundleOptions implements Interface for the `kn im bundle` command.
type BundleOptions struct {
	// ImageName is the string name of the bundle image to which we should publish things.
	ImageName string

	// tag is the processed version of ImageName that is populated while validating it.
	tag name.Tag

	// Director is the string containing the directory to bundle.
	Directory string

	// GitLocation the git location used to git clone the source if not using a bundle
	GitLocation *source.GitLocation
}

// BundleOptions implements Interface
var _ Interface = (*BundleOptions)(nil)

// AddFlags implements Interface
func (opts *BundleOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().String("bundle", "", "Where to publish the bundle.")
	cmd.Flags().String("directory", ".", "The directory to bundle up.")

	cmd.Flags().String("git-url", "", "The git URL to clone the source from if using git clone rather than a bundle image (e.g. if using mink inside a CI/CD pipeline).")
	cmd.Flags().String("git-rev", "", "The git revision (branch, tag, SHA) to clone the source from if using git clone rather than a bundle image (e.g. if using mink inside a CI/CD pipeline).")
	cmd.Flags().Bool("git-verbose", false, "If using git to clone the source enable verbose logging")
}

// Validate implements Interface
func (opts *BundleOptions) Validate(cmd *cobra.Command, args []string) error {
	viper.BindPFlags(cmd.Flags())

	opts.ImageName = viper.GetString("bundle")
	opts.Directory = viper.GetString("directory")

	gitURL := viper.GetString("git-url")
	if gitURL != "" {
		if opts.GitLocation == nil {
			opts.GitLocation = &source.GitLocation{}
		}
		opts.GitLocation.URL = gitURL
		opts.GitLocation.Revision = viper.GetString("git-rev")
		opts.GitLocation.Verbose = viper.GetBool("git-verbose")

		// lets create a sample source bundle image...
		var err error
		opts.tag, err = name.NewTag("gcr.io/sample/source-bundle:latest", name.WeakValidation)
		if err != nil {
			return err
		}
	}

	if opts.GitLocation != nil {
		return nil
	}
	if opts.ImageName == "" {
		return apis.ErrMissingField("bundle")
	} else if tag, err := name.NewTag(opts.ImageName, name.WeakValidation); err != nil {
		return apis.ErrInvalidValue(err.Error(), "bundle")
	} else {
		opts.tag = tag
	}
	return nil
}

// Execute implements Interface
func (opts *BundleOptions) Execute(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return errors.New("'im bundle' does not take any arguments")
	}

	digest, err := kontext.Bundle(signals.NewContext(), opts.Directory, opts.tag)
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%s\n", digest.String())
	return nil
}

var bundleExample = fmt.Sprintf(`
  # Create a self-extracting bundle of the current directory.
  %[1]s bundle --bundle docker.io/mattmoor/bundle:latest

  # Create a self-extracting bundle of a sub-directory.
  %[1]s bundle --bundle docker.io/mattmoor/bundle:latest --directory subdir/`, ExamplePrefix())

// NewBundleCommand implements 'kn-im bundle' command
func NewBundleCommand() *cobra.Command {
	opts := &BundleOptions{}

	cmd := &cobra.Command{
		Use:     "bundle --bundle IMAGE",
		Short:   "Bundles the context into a self-extracting container image",
		Example: bundleExample,
		PreRunE: opts.Validate,
		RunE:    opts.Execute,
	}

	opts.AddFlags(cmd)

	return cmd
}
