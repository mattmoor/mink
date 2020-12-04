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
	"bytes"
	"net/url"
	"path"
	"strings"
	"text/template"

	"github.com/google/go-containerregistry/pkg/name"
	minkcli "github.com/mattmoor/mink/pkg/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// BaseBuildOptions implements Interface for the `kn im build` command.
type BaseBuildOptions struct {
	// Inherit all of the bundle options.
	BundleOptions

	// ImageName is the string name of the bundle image to which we should publish things.
	ImageName string

	// ServiceAccount is the name of the service account *as* which to run the build.
	ServiceAccount string

	// tmpl is the template used to instantiate image names.
	tmpl *template.Template
}

// BaseBuildOptions implements Interface
var _ Interface = (*BaseBuildOptions)(nil)

// AddFlags implements Interface
func (opts *BaseBuildOptions) AddFlags(cmd *cobra.Command) {
	// Add the bundle flags to our surface.
	opts.BundleOptions.AddFlags(cmd)

	cmd.Flags().String("image", "", "Where to publish the final image.  This can be a go template "+
		"and has access to the url.URL fields (e.g. Scheme, Host, Path) that would represent this "+
		"build with the resolve command.  Functions are also provided for: basename, dirname, join, "+
		"lower, and split.")
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
		return minkcli.ErrMissingFlag("image")
	} else if tmpl, err := template.New("image").Funcs(imageNameFunctions).Parse(opts.ImageName); err != nil {
		return minkcli.ErrInvalidValue("image", err.Error())
	} else {
		opts.tmpl = tmpl
		if _, err := opts.tag(imageNameContext{
			URL: url.URL{
				// Arbitrary choice, but we should always have at least scheme.
				Scheme: "dockerfile",
			},
		}); err != nil {
			return minkcli.ErrInvalidValue("image", err.Error())
		}
	}

	opts.ServiceAccount = viper.GetString("as")
	if opts.ServiceAccount == "" {
		return minkcli.ErrMissingFlag("as")
	}

	return nil
}

type imageNameContext struct {
	url.URL
}

var imageNameFunctions = template.FuncMap{
	"basename": path.Base,
	"dirname":  path.Dir,
	"join":     path.Join,
	"split":    strings.Split,
	"lower":    strings.ToLower,
}

func (opts *BaseBuildOptions) tag(inc imageNameContext) (name.Tag, error) {
	buf := bytes.NewBuffer(nil)
	if err := opts.tmpl.Execute(buf, inc); err != nil {
		return name.Tag{}, err
	}
	return name.NewTag(strings.TrimSpace(buf.String()), name.WeakValidation)
}
