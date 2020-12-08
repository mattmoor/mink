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
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/mattmoor/mink/pkg/builds"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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
	opts := &InstallOptions{}

	cmd := &cobra.Command{
		Use:     "install",
		Short:   "Installs mink on the current cluster context.",
		PreRunE: opts.Validate,
		RunE:    opts.Execute,
	}
	opts.AddFlags(cmd)

	return cmd
}

// InstallOptions implements Interface for the `kn im install` command.
type InstallOptions struct {
	Domain   string
	InMemory bool
	Replicas int
}

// InstallOptions implements Interface
var _ Interface = (*InstallOptions)(nil)

// AddFlags implements Interface
func (opts *InstallOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().String("domain", "", "The domain to configure the cluster to use.")
	cmd.Flags().Bool("disable-imc", false, "Whether to install the in-memory channel.")
	cmd.Flags().Int("replicas", 1, "The number of controlplane replicas to run.")
}

// Validate implements Interface
func (opts *InstallOptions) Validate(cmd *cobra.Command, args []string) error {
	viper.BindPFlags(cmd.Flags())

	opts.InMemory = !viper.GetBool("disable-imc")
	opts.Domain = viper.GetString("domain")

	opts.Replicas = viper.GetInt("replicas")
	if opts.Replicas == 0 {
		opts.Replicas = 1
	}

	return nil
}

// Execute implements Interface
func (opts *InstallOptions) Execute(cmd *cobra.Command, args []string) error {
	cmd.Print("Cleaning up any old jobs.\n")
	if err := cleanupJobs(cmd); err != nil {
		return err
	}

	if err := install(cmd, "mink core", CoreReleaseURI); err != nil {
		return err
	}

	controlplane := []string{"statefulsets/controlplane"}
	for _, cp := range controlplane {
		if err := wait(cmd, cp); err != nil {
			return err
		}
	}
	if opts.Domain != "" {
		cmd.Printf("Configuring Serving to use %s for DNS.\n", opts.Domain)
		if err := patchResource(cmd, "configmap/config-domain", fmt.Sprintf(`{"data":{%q:""}}`, opts.Domain)); err != nil {
			return err
		}
	} else {
		cmd.Print("Waiting for mink webhook to be ready.\n")
		if err := patchResource(cmd, "configmap/config-network", `{"data":{"bogus":"value"}}`); err != nil {
			return err
		}
	}

	if opts.Replicas > 1 {
		cmd.Printf("Configuring high-availability to use %d replicas.\n", opts.Replicas)
		if err := patchResource(cmd, "configmap/config-leader-election", fmt.Sprintf(`{"data":{"buckets":"%d"}}`, opts.Replicas)); err != nil {
			return err
		}
		for _, cp := range controlplane {
			// Scale to zero first, so pods aren't left with stale config.
			cmd.Printf("Scaling down %s.\n", cp)
			if err := patchResource(cmd, cp, `{"spec":{"replicas":0}}`); err != nil {
				return err
			}
			if err := wait(cmd, cp); err != nil {
				return err
			}

			cmd.Printf("Scaling up %s to %d replicas.\n", cp, opts.Replicas)
			if err := patchResource(cmd, cp, fmt.Sprintf(`{"spec":{"replicas":%d}}`, opts.Replicas)); err != nil {
				return err
			}
			if err := wait(cmd, cp); err != nil {
				return err
			}
		}
	}
	if err := wait(cmd, "daemonsets/dataplane"); err != nil {
		return err
	}

	if opts.InMemory {
		if err := install(cmd, "in-memory channel", InMemoryReleaseURI); err != nil {
			return err
		}
		if err := wait(cmd, "deployments/imc-controller"); err != nil {
			return err
		}
		if err := wait(cmd, "deployments/imc-dispatcher"); err != nil {
			return err
		}
	}

	cmd.Print("Waiting for jobs to complete.\n")
	if err := waitJobs(cmd); err != nil {
		return err
	}
	if err := cleanupJobs(cmd); err != nil {
		return err
	}

	// TODO(mattmoor): Consider waiting for Jobs
	// TODO(mattmoor): Clean up Jobs.

	if opts.Domain != "" {
		// TODO(mattmoor): expose masterURL and kubeconfig flags.
		cfg, err := builds.GetConfig("", "")
		if err != nil {
			return err
		}
		client, err := kubernetes.NewForConfig(cfg)
		if err != nil {
			return err
		}

		svc, err := client.CoreV1().Services("mink-system").Get(context.Background(), "envoy-external", metav1.GetOptions{})
		if err != nil {
			return err
		}
		cmd.Print("Please configure the following DNS records to complete setup:\n")
		// Instruct the user on how to configure their DNS records.
		for _, ing := range svc.Status.LoadBalancer.Ingress {
			switch {
			case ing.IP != "":
				cmd.Printf("  *.%s == A %s\n", opts.Domain, ing.IP)
			case ing.Hostname != "":
				cmd.Printf("  *.%s == CNAME %s\n", opts.Domain, ing.Hostname)
			}
		}
	}

	cmd.Print("mink installation complete!\n")
	return nil
}

func cleanupJobs(cmd *cobra.Command) error {
	argv := []string{"delete", "jobs", "-n", "mink-system", "--all"}

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

func patchResource(cmd *cobra.Command, resource, patch string) error {
	argv := []string{
		"patch", resource,
		"--namespace", "mink-system",
		"--type", "merge",
		"--patch", patch,
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

func wait(cmd *cobra.Command, thing string) error {
	argv := []string{
		"rollout", "status",
		"--timeout", "5m",
		"--namespace", "mink-system",
		thing,
	}
	cmd.Printf("Waiting for %s rollout to complete.\n", thing)

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

func waitJobs(cmd *cobra.Command) error {
	argv := []string{
		"wait", "jobs",
		"--timeout", "5m",
		"--namespace", "mink-system",
		"--for", "condition=complete",
		"--all",
	}

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
