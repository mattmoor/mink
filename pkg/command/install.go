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
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
)

var (
	// CoreReleaseURI holds the URI for the release yaml to install.
	// This will generally hold a path to config/core, which is infused
	// via linker flags at build time.  For development, this will
	// typically hold a local file path, but for releases will point to
	// a URL.  See ./hack/build-flags.sh for the linker magic.
	CoreReleaseURI = ""

	// InMemoryReleaseURI is the analog of CoreReleaseURI for config/in-memory.
	InMemoryReleaseURI = ""
)

// NewInstallCommand implements 'kn-im install' command
func NewInstallCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Installs mink on the current cluster context.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cleanupJobs(cmd); err != nil {
				return err
			}

			if err := install(cmd, "mink core", CoreReleaseURI); err != nil {
				return err
			}
			if err := waitNonJob(cmd, "mink core"); err != nil {
				return err
			}
			if err := awaitWebhook(cmd); err != nil {
				return err
			}

			if err := install(cmd, "in-memory channel", InMemoryReleaseURI); err != nil {
				return err
			}
			if err := waitNonJob(cmd, "in-memory channel"); err != nil {
				return err
			}

			// TODO(mattmoor): Consider waiting for Jobs, but how to deal with default-domain on KinD?

			cmd.Print("mink installation complete!\n")
			return nil
		},
	}
}

func cleanupJobs(cmd *cobra.Command) error {
	argv := []string{"delete", "jobs", "-n", "mink-system", "--all"}
	cmd.Print("Cleaning up any old jobs.\n")

	kubectlCmd := exec.Command("kubectl", argv...)

	// Pass through our environment
	kubectlCmd.Env = os.Environ()

	// For debugging.
	buf := &bytes.Buffer{}
	kubectlCmd.Stderr = buf
	kubectlCmd.Stdout = buf

	if err := kubectlCmd.Run(); err != nil {
		cmd.PrintErr(buf.String())
		return err
	}
	return nil
}

func install(cmd *cobra.Command, label, uri string) error {
	cmd.Printf("Installing %s from: %s\n", label, uri)
	argv := []string{"apply", "-f", uri}

	kubectlCmd := exec.Command("kubectl", argv...)

	// Pass through our environment
	kubectlCmd.Env = os.Environ()

	// For debugging.
	buf := &bytes.Buffer{}
	kubectlCmd.Stderr = buf
	kubectlCmd.Stdout = buf

	if err := kubectlCmd.Run(); err != nil {
		cmd.PrintErr(buf.String())
		return err
	}
	return nil
}

func awaitWebhook(cmd *cobra.Command) error {
	cmd.Print("Waiting for mink webhook to be ready.\n")
	argv := []string{
		"patch", "configmap/config-network",
		"--namespace", "mink-system",
		"--type", "merge",
		"--patch", `{"data":{"bogus":"value"}}`,
	}

	timeout := time.After(2 * time.Minute)
	for {
		kubectlCmd := exec.Command("kubectl", argv...)

		// Pass through our environment
		kubectlCmd.Env = os.Environ()

		// For debugging.
		buf := &bytes.Buffer{}
		kubectlCmd.Stderr = buf
		kubectlCmd.Stdout = buf

		err := kubectlCmd.Run()
		if err == nil {
			return nil
		}

		select {
		case <-timeout:
			cmd.PrintErr(buf.String())
			return err
		default:
			// Wait between attempts.
			time.Sleep(1 * time.Second)
		}
	}
}

func waitNonJob(cmd *cobra.Command, label string) error {
	argv := []string{
		"wait", "pods",
		"--timeout", "5m",
		"--namespace", "mink-system",
		"--for", "condition=Ready",
		"--selector", "!job-name",
	}
	cmd.Printf("Waiting for %s to be ready.\n", label)

	kubectlCmd := exec.Command("kubectl", argv...)

	// Pass through our environment
	kubectlCmd.Env = os.Environ()

	// For debugging.
	buf := &bytes.Buffer{}
	kubectlCmd.Stderr = buf
	kubectlCmd.Stdout = buf

	if err := kubectlCmd.Run(); err != nil {
		cmd.PrintErr(buf.String())
		return err
	}
	return nil
}
