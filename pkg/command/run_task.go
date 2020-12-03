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

	"github.com/mattmoor/mink/pkg/builds"
	"github.com/spf13/cobra"
	"github.com/tektoncd/cli/pkg/cli"
	"github.com/tektoncd/cli/pkg/options"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	tektonclientset "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"knative.dev/pkg/signals"
)

var runTaskExample = fmt.Sprintf(`
  # Create a TaskRun instantiating the task "build-stuff" and passing
  # values to named arguments.
  %[1]s run task build-stuff -- --arg1=val1 --arg2=val2
`, ExamplePrefix())

// NewRunTaskCommand implements 'kn-im run task' command
func NewRunTaskCommand() *cobra.Command {
	opts := &RunTaskOptions{}

	cmd := &cobra.Command{
		Use:          "task NAME",
		Short:        "Create a TaskRun to execute a task.",
		Example:      runTaskExample,
		SilenceUsage: true,
		Args: func(cmd *cobra.Command, args []string) error {
			// The index of --
			dashIdx := cmd.ArgsLenAtDash()

			posArgs := args
			if dashIdx != -1 {
				posArgs = posArgs[:dashIdx]
			}
			// We want exactly one positional argument
			// before a possible -- token.
			return cobra.ExactArgs(1)(cmd, posArgs)
		},
		PreRunE: opts.Validate,
		RunE:    opts.Execute,
	}

	opts.AddFlags(cmd)

	return cmd
}

// RunTaskOptions implements Interface for the `kn im run task` command.
type RunTaskOptions struct {
	// Inherit all of the base build options.
	BaseBuildOptions
}

// RunTaskOptions implements Interface
var _ Interface = (*RunTaskOptions)(nil)

// AddFlags implements Interface
func (opts *RunTaskOptions) AddFlags(cmd *cobra.Command) {
	// Add the bundle flags to our surface.
	opts.BaseBuildOptions.AddFlags(cmd)

}

// Validate implements Interface
func (opts *RunTaskOptions) Validate(cmd *cobra.Command, args []string) error {
	// Validate the bundle arguments.
	if err := opts.BaseBuildOptions.Validate(cmd, args); err != nil {
		return err
	}
	return nil
}

// Execute implements Interface
func (opts *RunTaskOptions) Execute(cmd *cobra.Command, args []string) error {
	ctx := signals.NewContext()

	// TODO(mattmoor): expose masterURL and kubeconfig flags.
	cfg, err := builds.GetConfig("", "")
	if err != nil {
		return err
	}
	client, err := tektonclientset.NewForConfig(cfg)
	if err != nil {
		return err
	}

	// Load the task definition.
	taskName := args[0]
	task, err := client.TektonV1beta1().Tasks(Namespace()).Get(ctx, taskName, metav1.GetOptions{})
	if apierrs.IsNotFound(err) {
		cmd.Printf("Task %q not found\n", fmt.Sprintf("%s/%s", Namespace(), taskName))
		return err
	} else if err != nil {
		return err
	}

	var processors []Processor

	taskCmd := &cobra.Command{
		Use:   "mink run task " + task.Name,
		Short: task.Spec.Description,
		RunE: func(cmd *cobra.Command, args []string) error {
			tr := &v1beta1.TaskRun{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    Namespace(),
					GenerateName: "mink-" + task.Name + "-",
				},
				Spec: v1beta1.TaskRunSpec{
					TaskRef: &v1beta1.TaskRef{
						Name: task.Name,
					},
				},
			}

			for _, processor := range processors {
				ps, err := processor.PreRun(task.Spec.Params)
				if err != nil {
					return err
				}
				tr.Spec.Params = append(tr.Spec.Params, ps...)
			}

			tr, err := builds.RunTask(ctx, tr, &options.LogOptions{
				Params: &cli.TektonParams{},
				Stream: &cli.Stream{
					// Send Out to stderr so we can capture the digest for composition.
					Out: cmd.OutOrStderr(),
					Err: cmd.OutOrStderr(),
				},
				Follow: true,
			}) //, builds.WithTaskServiceAccount(opts.ServiceAccount))
			if err != nil {
				return err
			}

			for _, processor := range processors {
				if err := processor.PostRun(tr.Status.TaskRunResults); err != nil {
					return err
				}
			}
			return nil
		},
	}

	// Process the results of the task.
	results := make(sets.String, len(task.Spec.Results))
	for _, result := range task.Spec.Results {
		results.Insert(result.Name)
	}

	// Based on the signature determine which processors to wire in.
	processors, err = detectProcessors(taskCmd, task.Spec.Params, results)
	if err != nil {
		return err
	}

	taskCmd.SetArgs(args[1:])

	return taskCmd.Execute()
}
