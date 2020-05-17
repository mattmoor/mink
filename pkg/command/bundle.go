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
	"github.com/mattmoor/mink/pkg/kontext"
	"github.com/spf13/cobra"
)

var bundle_example = fmt.Sprintf(`
  # Create a self-extracting bundle of the current directory.
  %[1]s bundle --image docker.io/mattmoor/bundle:latest

  # Create a self-extracting bundle of a sub-directory.
  %[1]s bundle --image docker.io/mattmoor/bundle:latest --directory subdir/`, ExamplePrefix())

// NewBundleCommand implements 'kn-im bundle' command
func NewBundleCommand() *cobra.Command {
	var directory string
	var image string

	cmd := &cobra.Command{
		Use:     "bundle --image IMAGE",
		Short:   "Bundles the context into a container image",
		Example: bundle_example,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return errors.New("'im bundle' does not take any arguments.")
			}

			tag, err := name.NewTag(image, name.WeakValidation)
			if err != nil {
				return err
			}
			digest, err := kontext.Bundle(context.Background(), directory, tag)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s\n", digest.String())
			return nil
		},
	}

	cmd.Flags().StringVarP(&image, "image", "i", "", "Where to publish the bundle.")
	cmd.MarkFlagRequired("image")

	cmd.Flags().StringVarP(&directory, "directory", "d", ".", "The directory to bundle up.")

	return cmd
}
