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

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cranecmd "github.com/google/go-containerregistry/cmd/crane/cmd"
	"github.com/mattmoor/mink/pkg/command"

	// Support GCP auth
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var rootCmd *cobra.Command

func init() {
	binaryName := command.BinaryName()
	if binaryName == "kn-im" {
		rootCmd = &cobra.Command{
			// Unclear how we make `use` show up as `kn im build` instead
			// of `kn build`, so using `kn-im` here, so it at least isn't
			// wrong.
			Use:   "kn-im",
			Short: "kn plugin for mink",
			Long:  "'kn im' is mink backwards!",
		}

		// TODO(mattmoor): Is there a way to set up the same shared flags as `kn`
		// gives us through kncore.NewKnCommand()?

	} else {
		// TODO(mattmoor): Use the base `kn` root command when we are invoked via `mink`
		// rootCmd = kncore.NewKnCommand()
		// rootCmd.Use = "mink"
		// rootCmd.Short = "The command-line for mink."
		// rootCmd.Long = "TODO"
		rootCmd = &cobra.Command{
			Use:   "mink",
			Short: "The command-line for mink",
			Long:  "TODO",
		}

		// Add the auth group from crane to facilitate logging into a registry.
		rootCmd.AddCommand(cranecmd.NewCmdAuth())
	}

	// TODO(mattmoor): Have these take a commands.KnParams
	rootCmd.AddCommand(command.NewVersionCommand())
	rootCmd.AddCommand(command.NewBundleCommand())
	rootCmd.AddCommand(command.NewBuildCommand())
	rootCmd.AddCommand(command.NewBuildpackCommand())
	rootCmd.AddCommand(command.NewInstallCommand())
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}
}
