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
	"github.com/mattmoor/mink/pkg/bundles/git"
	"github.com/mattmoor/mink/pkg/bundles/kontext"
	minkcli "github.com/mattmoor/mink/pkg/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"k8s.io/apimachinery/pkg/util/sets"
	"knative.dev/pkg/signals"
)

// BundleMode indicates the type of bundling mode that this bundle command is performing.
type BundleMode int

// These constants define the different bundle "modes".
const (
	UnknownMode BundleMode = iota
	KontextMode
	GitMode
)

// String implements fmt.Stringer
func (m BundleMode) String() string {
	return [...]string{"Unknown", "Kontext", "Git"}[m]
}

// BundleOptions implements Interface for the `kn im bundle` command.
type BundleOptions struct {
	// ImageName is the string name of the bundle image to which we should publish things.
	ImageName string

	// tag is the processed version of ImageName that is populated while validating it.
	tag name.Tag

	// Directory is the string containing the directory to bundle.
	// This option signals "kontext mode".
	Directory string

	// GitURL is the URL of the git repository to clone.
	// This option signals "git mode".
	GitURL string
	// GitRef is the ref to check out within the above repository.
	GitRef plumbing.ReferenceName

	// mode holds the type of bundling we are performing.
	mode BundleMode
}

// BundleOptions implements Interface
var _ Interface = (*BundleOptions)(nil)

// AddFlags implements Interface
func (opts *BundleOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().String("bundle", "", "Where to publish the bundle.")

	// KontextMode options
	cmd.Flags().String("directory", "", "The directory to bundle up.")

	// GitMode options
	cmd.Flags().String("git-url", "", "The git repository to bundle.")
	cmd.Flags().String("git-ref", "", "The ref within the git repository to bundle.")
}

// Validate implements Interface
func (opts *BundleOptions) Validate(cmd *cobra.Command, args []string) error {
	viper.BindPFlags(cmd.Flags())
	opts.ImageName = viper.GetString("bundle")
	opts.mode = UnknownMode

	// Check for multiple mode-determining options.
	seen := sets.NewString()
	for _, key := range []string{"directory", "get-url"} {
		if viper.GetString(key) != "" {
			seen.Insert(key)
		}
	}
	if len(seen) > 1 {
		return fmt.Errorf("saw multiple configuration options %v, expected one", seen.List())
	}

	// See if we're in "kontext mode"
	opts.Directory = viper.GetString("directory")
	if opts.Directory != "" {
		opts.mode = KontextMode
	}

	// See if we're in "git mode"
	opts.GitURL = viper.GetString("git-url")
	if opts.GitURL != "" {
		opts.mode = GitMode
		if ref := viper.GetString("git-ref"); ref != "" {
			opts.GitRef = plumbing.ReferenceName(ref)
		}
	}

	// When all else fails, fallback on the current directory and context mode.
	if opts.mode == UnknownMode {
		opts.mode = KontextMode
		opts.Directory = "."
	}

	if opts.ImageName == "" {
		return minkcli.ErrMissingFlag("bundle")
	} else if tag, err := name.NewTag(opts.ImageName, name.WeakValidation); err != nil {
		return minkcli.ErrInvalidValue("bundle", err.Error())
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

	digest, err := opts.bundle(signals.NewContext())
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%s\n", digest.String())
	return nil
}

func (opts *BundleOptions) bundle(ctx context.Context) (name.Digest, error) {
	switch opts.mode {
	case KontextMode:
		return kontext.Bundle(ctx, opts.Directory, opts.tag)
	case GitMode:
		return git.Bundle(ctx, git.Options{
			URL: opts.GitURL,
			Ref: opts.GitRef,
		}, opts.tag)
	default:
		return name.Digest{}, fmt.Errorf("unsupported mode %v", opts.mode)
	}
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
