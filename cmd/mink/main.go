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
	"path/filepath"
	"strings"
	"sync"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"knative.dev/pkg/injection"
	"knative.dev/pkg/signals"

	cranecmd "github.com/google/go-containerregistry/cmd/crane/cmd"
	"github.com/mattmoor/mink/pkg/command"

	// Support GCP auth
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var rootCmd *cobra.Command
var once sync.Once

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
		// TODO(https://github.com/mattmoor/mink/issues/59): Use the base `kn` root command when we are invoked via `mink`
		// rootCmd = kncore.NewKnCommand()
		// rootCmd.Use = "mink"
		// rootCmd.Short = "The command-line for mink."
		// rootCmd.Long = "TODO"
		rootCmd = &cobra.Command{
			Use:   "mink",
			Short: "The command-line for mink",
			Long:  "mink is a collection of commands to stitch together a cohesive experience across Knative Serving & Eventing and Tekton Pipelines.",
		}

		// Add the auth group from crane to facilitate logging into a registry.
		rootCmd.AddCommand(cranecmd.NewCmdAuth())
	}

	// We do not start informers.
	ctx, _ := injection.EnableInjectionOrDie(signals.NewContext(), nil)

	// TODO(mattmoor): Have these take a commands.KnParams
	rootCmd.AddCommand(command.NewVersionCommand())

	rootCmd.AddCommand(command.NewInstallCommand(ctx))

	rootCmd.AddCommand(command.NewBundleCommand(ctx))
	rootCmd.AddCommand(command.NewBuildCommand(ctx))
	rootCmd.AddCommand(command.NewBuildpackCommand(ctx))
	rootCmd.AddCommand(command.NewRunCommand(ctx))

	rootCmd.AddCommand(command.NewResolveCommand(ctx))
	rootCmd.AddCommand(command.NewApplyCommand(ctx))

	cobra.OnInitialize(func() {
		// In the context of mink run we might run this multiple times,
		// and mink resolve (w/ run executors) might run those in parallel.
		// viper does not use threadsafe maps, which can lead to problems.
		// See: https://github.com/mattmoor/mink/issues/363
		once.Do(initViperConfig)
	})
}

// initViperConfig reads in config file and ENV variables if set.
func initViperConfig() {
	searchpath := make([]string, 0, 10)
	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		// avoid color since we don't know if it should be enabled yet
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// We don't use this handling, but we need to call
	// these for things to work properly.
	viper.AddConfigPath(home)
	viper.SetConfigName(".mink")
	viper.SetConfigType("yaml")

	filename := ".mink.yaml"
	if nearest := nearestConfig(filename); nearest != "" {
		searchpath = append(searchpath, nearest)
	}
	searchpath = append(searchpath, filepath.Join(home, filename))
	// TODO(mattmoor): Consider adding a system-side file, e.g. /etc/mink.yaml

	// Perform our own search handling in order to configure our own precedence.
	found := false
	for _, path := range searchpath {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}

		if err := func() error { // Scope defer
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()

			// Read the first config we file, and merge the rest.
			if !found {
				err = viper.ReadConfig(f)
				found = true
			} else {
				err = viper.MergeConfig(f)
			}
			if err != nil {
				return err
			}
			return nil
		}(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	viper.SetEnvPrefix("MINK")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv() // read in environment variables that match
}

func nearestConfig(filename string) string {
	// Find working directory.
	wd, err := os.Getwd()
	if err != nil {
		// avoid color since we don't know if it should be enabled yet
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for ; wd != filepath.Dir(wd); wd = filepath.Dir(wd) {
		p := filepath.Join(wd, filename)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			continue
		}
		return p
	}
	return ""
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}
}
