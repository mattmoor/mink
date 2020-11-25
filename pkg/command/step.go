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
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jenkins-x/jx-helpers/v3/pkg/files"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var stepExample = fmt.Sprintf(`
  # Generate a .mink.yaml file if it does not exist and a dockerfile or build pack can be detected
  # then build and publish references within .mink.yaml either outputs the YAML or saves it in place
  %[1]s step
`, ExamplePrefix())

// NewStepCommand implements 'kn-im resolve' command
func NewStepCommand() *cobra.Command {
	opts := &StepOptions{}

	cmd := &cobra.Command{
		Use:     "step",
		Short:   "Runs the pipeline step to build publish and resolve image references within a collection of yaml files if required.",
		Example: stepExample,
		PreRunE: opts.Validate,
		RunE:    opts.Execute,
	}

	opts.AddFlags(cmd)

	return cmd
}

// StepOptions implements Interface for the `kn im resolve` command.
type StepOptions struct {
	// Inherit all of the resolve options.
	ResolveOptions

	InitOptions InitOptions

	// Out the output destination
	Out io.Writer
}

// StepOptions implements Interface
var _ Interface = (*StepOptions)(nil)

// AddFlags implements Interface
func (opts *StepOptions) AddFlags(cmd *cobra.Command) {
	opts.InitOptions.InStepCommand = true

	// Add the bundle flags to our surface.
	opts.ResolveOptions.AddFlags(cmd)

	opts.InitOptions.AddFlags(cmd)
}

// Validate implements Interface
func (opts *StepOptions) Validate(cmd *cobra.Command, args []string) error {
	// Validate the bundle arguments.
	if err := opts.ResolveOptions.Validate(cmd, args); err != nil {
		return err
	}

	// InitOptions flags
	if err := opts.InitOptions.Validate(cmd, args); err != nil {
		return err
	}
	return nil
}

// Execute implements Interface
func (opts *StepOptions) Execute(cmd *cobra.Command, args []string) error {
	if opts.Out == nil {
		opts.Out = cmd.OutOrStdout()
	}
	err := opts.InitOptions.Execute(cmd, nil)
	if err != nil {
		return errors.Wrapf(err, "failed to detect .mink.yaml file")
	}

	if opts.ResolveOptions.LocalKaniko {
		err = opts.copyKanikoDockerSecrets()
		if err != nil {
			return errors.Wrapf(err, "failed to copy kaniko docker secrets")
		}
	}

	if !opts.InitOptions.MinkEnabled {
		return nil
	}
	return opts.ResolveOptions.Execute(cmd, args)
}

func (opts *StepOptions) copyKanikoDockerSecrets() error {
	glob := filepath.Join("/tekton", "creds-secrets", "*", ".dockerconfigjson")
	fs, err := filepath.Glob(glob)
	if err != nil {
		return errors.Wrapf(err, "failed to find tekton secrets")
	}
	if len(fs) == 0 {
		fmt.Fprintf(opts.Out, "failed to find docker secrets %s\n", glob)
		return nil
	}
	srcFile := fs[0]

	outDir := filepath.Join("/kaniko", ".docker")
	err = os.MkdirAll(outDir, files.DefaultDirWritePermissions)
	if err != nil {
		return errors.Wrapf(err, "failed to create dir %s", outDir)
	}
	outFile := filepath.Join(outDir, "config.json")
	err = files.CopyFile(srcFile, outFile)
	if err != nil {
		return errors.Wrapf(err, "failed to copy file %s to %s", srcFile, outFile)
	}

	fmt.Fprintf(opts.Out, "copied secret %s to %s\n", srcFile, outFile)
	return nil
}
