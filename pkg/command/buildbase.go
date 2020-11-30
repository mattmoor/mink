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
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"knative.dev/pkg/apis"
)

// BaseBuildOptions implements Interface for the `kn im build` command.
type BaseBuildOptions struct {
	// Inherit all of the bundle options.
	BundleOptions

	// ImageName is the string name of the bundle image to which we should publish things.
	ImageName string

	// ServiceAccount is the name of the service account *as* which to run the build.
	ServiceAccount string
}

// BaseBuildOptions implements Interface
var _ Interface = (*BaseBuildOptions)(nil)

// AddFlags implements Interface
func (opts *BaseBuildOptions) AddFlags(cmd *cobra.Command) {
	// Add the bundle flags to our surface.
	opts.BundleOptions.AddFlags(cmd)

	cmd.Flags().String("image", "", "Where to publish the final image.")
	cmd.Flags().String("as", "default",
		"The name of the ServiceAccount as which to run the build, pass --as=me to "+
			"temporarily create a new ServiceAccount to push with your local credentials.")
}

// Validate implements Interface
func (opts *BaseBuildOptions) Validate(cmd *cobra.Command, args []string) error {
	// Validate the bundle arguments.
	if err := opts.BundleOptions.Validate(cmd, args); err != nil {
		return err
	}

	opts.ImageName = viper.GetString("image")
	if opts.ImageName == "" {
		return apis.ErrMissingField("image")
	} else if _, err := opts.Tag(); err != nil {
		return apis.ErrInvalidValue(err.Error(), "image")
	}

	opts.ServiceAccount = viper.GetString("as")
	if opts.ServiceAccount == "" {
		return apis.ErrMissingField("as")
	}

	return nil
}

func (opts *BaseBuildOptions) Tag() (name.Tag, error) {
	return name.NewTag(opts.ImageName, name.WeakValidation)
}
