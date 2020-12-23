// Copyright 2018 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package commands

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/google/ko/pkg/commands/options"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/sync/errgroup"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// addApply augments our CLI surface with apply.
func addApply(topLevel *cobra.Command) {
	koApplyFlags := []string{}
	po := &options.PublishOptions{}
	fo := &options.FilenameOptions{}
	so := &options.SelectorOptions{}
	sto := &options.StrictOptions{}
	bo := &options.BuildOptions{}
	apply := &cobra.Command{
		Use:   "apply -f FILENAME",
		Short: "Apply the input files with image references resolved to built/pushed image digests.",
		Long:  `This sub-command finds import path references within the provided files, builds them into Go binaries, containerizes them, publishes them, and then feeds the resulting yaml into "kubectl apply".`,
		Example: `
  # Build and publish import path references to a Docker
  # Registry as:
  #   ${KO_DOCKER_REPO}/<package name>-<hash of import path>
  # Then, feed the resulting yaml into "kubectl apply".
  # When KO_DOCKER_REPO is ko.local, it is the same as if
  # --local was passed.
  ko apply -f config/

  # Build and publish import path references to a Docker
  # Registry preserving import path names as:
  #   ${KO_DOCKER_REPO}/<import path>
  # Then, feed the resulting yaml into "kubectl apply".
  ko apply --preserve-import-paths -f config/

  # Build and publish import path references to a Docker
  # daemon as:
  #   ko.local/<import path>
  # Then, feed the resulting yaml into "kubectl apply".
  ko apply --local -f config/

  # Apply from stdin:
  cat config.yaml | ko apply -f -`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if !isKubectlAvailable() {
				log.Print("error: kubectl is not available. kubectl must be installed to use ko apply.")
				return
			}

			// Cancel on signals.
			ctx := createCancellableContext()

			builder, err := makeBuilder(ctx, bo)
			if err != nil {
				log.Fatalf("error creating builder: %v", err)
			}
			publisher, err := makePublisher(po)
			if err != nil {
				log.Fatalf("error creating publisher: %v", err)
			}
			defer publisher.Close()
			// Create a set of ko-specific flags to ignore when passing through
			// kubectl global flags.
			ignoreSet := make(map[string]struct{})
			for _, s := range koApplyFlags {
				ignoreSet[s] = struct{}{}
			}

			// Filter out ko flags from what we will pass through to kubectl.
			kubectlFlags := []string{}
			cmd.Flags().Visit(func(flag *pflag.Flag) {
				if _, ok := ignoreSet[flag.Name]; !ok {
					kubectlFlags = append(kubectlFlags, "--"+flag.Name, flag.Value.String())
				}
			})

			// Issue a "kubectl apply" command reading from stdin,
			// to which we will pipe the resolved files.
			argv := []string{"apply", "-f", "-"}
			argv = append(argv, kubectlFlags...)
			kubectlCmd := exec.CommandContext(ctx, "kubectl", argv...)

			// Pass through our environment
			kubectlCmd.Env = os.Environ()
			// Pass through our std{out,err} and make our resolved buffer stdin.
			kubectlCmd.Stderr = os.Stderr
			kubectlCmd.Stdout = os.Stdout

			// Wire up kubectl stdin to resolveFilesToWriter.
			stdin, err := kubectlCmd.StdinPipe()
			if err != nil {
				log.Fatalf("error piping to 'kubectl apply': %v", err)
			}

			// Make sure builds are cancelled if kubectl apply fails.
			g, ctx := errgroup.WithContext(ctx)
			g.Go(func() error {
				// kubectl buffers data before starting to apply it, which
				// can lead to resources being created more slowly than desired.
				// In the case of --watch, it can lead to resources not being
				// applied at all until enough iteration has occurred.  To work
				// around this, we prime the stream with a bunch of empty objects
				// which kubectl will discard.
				// See https://github.com/google/go-containerregistry/pull/348
				for i := 0; i < 1000; i++ {
					stdin.Write([]byte("---\n"))
				}
				// Once primed kick things off.
				return resolveFilesToWriter(ctx, builder, publisher, fo, so, sto, stdin)
			})

			g.Go(func() error {
				// Run it.
				if err := kubectlCmd.Run(); err != nil {
					return fmt.Errorf("error executing 'kubectl apply': %v", err)
				}
				return nil
			})

			if err := g.Wait(); err != nil {
				log.Fatal(err)
			}
		},
	}
	options.AddPublishArg(apply, po)
	options.AddFileArg(apply, fo)
	options.AddSelectorArg(apply, so)
	options.AddBuildOptions(apply, bo)

	// Collect the ko-specific apply flags before registering the kubectl global
	// flags so that we can ignore them when passing kubectl global flags through
	// to kubectl.
	apply.Flags().VisitAll(func(flag *pflag.Flag) {
		koApplyFlags = append(koApplyFlags, flag.Name)
	})

	// Register the kubectl global flags.
	kubeConfigFlags := genericclioptions.NewConfigFlags(false)
	kubeConfigFlags.AddFlags(apply.Flags())

	topLevel.AddCommand(apply)
}
