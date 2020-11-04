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
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"knative.dev/pkg/signals"
)

var applyExample = fmt.Sprintf(`
  # Build and publish references within foo.yaml, and pipe the resolved output to kubectl apply.
  %[1]s apply -f config/foo.yaml

  # Build and publish references within yaml files recursively under config/, and
  # pipe the resolved output to kubectl apply
  %[1]s apply -Rf config/

  # Customize the builder used for buildpack:/// builds
  %[1]s apply -f config/ --builder foo:latest

  # Customize the name of the overrides file used for buildpack:/// builds
  %[1]s apply -f config/ --overrides another-name.toml

  # Customize the name of Dockerfiles to use for dockerfile:/// builds
  %[1]s apply -f config/ --dockerfile Dockerfile.production`, ExamplePrefix())

// NewApplyCommand implements 'kn-im apply' command
func NewApplyCommand() *cobra.Command {
	opts := &ApplyOptions{}

	cmd := &cobra.Command{
		Use:     "apply -f FILE",
		Short:   "Build and publish image references within a collection of yaml files, and apply them to the cluster.",
		Example: applyExample,
		PreRunE: opts.Validate,
		RunE:    opts.Execute,
	}

	opts.AddFlags(cmd)

	return cmd
}

// ApplyOptions implements Interface for the `kn im apply` command.
type ApplyOptions struct {
	// Inherit all of the base build options.
	ResolveOptions
}

// ApplyOptions implements Interface
var _ Interface = (*ApplyOptions)(nil)

// Execute implements Interface
func (opts *ApplyOptions) Execute(cmd *cobra.Command, args []string) error {
	// Issue a "kubectl apply" command reading from stdin,
	// to which we will pipe the resolved files.
	// TODO(mattmoor): mimic kubectl flags like ko does.
	argv := []string{"apply", "-f", "-"}
	kubectlCmd := exec.Command("kubectl", argv...)

	// Pass through our environment
	kubectlCmd.Env = os.Environ()
	// Pass through our std{out,err} and make our resolved buffer stdin.
	kubectlCmd.Stderr = os.Stderr
	kubectlCmd.Stdout = os.Stdout

	// Wire up kubectl stdin to resolveFilesToWriter.
	stdin, err := kubectlCmd.StdinPipe()
	if err != nil {
		return err
	}

	cmd.SetOutput(stdin)

	// Make sure builds are cancelled if kubectl apply fails.
	g, ctx := errgroup.WithContext(signals.NewContext())
	g.Go(func() error {
		defer stdin.Close()
		return opts.ResolveOptions.execute(ctx, cmd)
	})
	g.Go(kubectlCmd.Run)
	return g.Wait()
}
