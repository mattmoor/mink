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
	"context"
	"fmt"

	"github.com/mattmoor/mink/pkg/builds"
	"github.com/spf13/cobra"
	"github.com/tektoncd/cli/pkg/cli"
	"github.com/tektoncd/cli/pkg/options"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	pipelineclient "github.com/tektoncd/pipeline/pkg/client/injection/client"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

var runTaskExample = fmt.Sprintf(`
  # Create a TaskRun instantiating the task "build-stuff" and passing
  # values to named arguments.
  %[1]s run task build-stuff -- --arg1=val1 --arg2=val2
`, ExamplePrefix())

// NewRunTaskCommand implements 'kn-im run task' command
func NewRunTaskCommand(ctx context.Context) *cobra.Command {
	opts := &RunTaskOptions{
		RunOptions: RunOptions{
			BaseBuildOptions: BaseBuildOptions{BundleOptions: BundleOptions{ctx: ctx}},
			resource:         "task",
		},
	}

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
	// Inherit all of the base run options.
	RunOptions
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
	ctx := opts.GetContext(cmd)

	// We take one positional argument, pass that as the task name.
	taskCmd, err := opts.buildCmd(ctx, args[0], opts.detectProcessors)
	if err != nil {
		return err
	}

	// Pass the remaining arguments to the sub-command.
	// These are all after the --
	taskCmd.SetArgs(args[1:])

	return taskCmd.Execute()
}

// buildCmd constructs a cobra.Command for the named task.
func (opts *RunTaskOptions) buildCmd(ctx context.Context, taskName string, detector signatureDetector) (*cobra.Command, error) {
	client := pipelineclient.Get(ctx)

	// Load the task definition.
	ns := opts.Namespace
	task, err := client.TektonV1beta1().Tasks(ns).Get(ctx, taskName, metav1.GetOptions{})
	if apierrs.IsNotFound(err) {
		return nil, fmt.Errorf("task %q not found: %w", fmt.Sprintf("%s/%s", ns, taskName), err)
	} else if err != nil {
		return nil, err
	}

	var processors []Processor
	taskCmd := &cobra.Command{
		Use:   "mink run task " + task.Name,
		Short: task.Spec.Description,
		RunE: func(cmd *cobra.Command, args []string) error {
			tr := &v1beta1.TaskRun{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ns,
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
				ActivityTimeout: activityTimeout,
				Params:          &cli.TektonParams{},
				Stream: &cli.Stream{
					// Send Out to stderr so we can capture the digest for composition.
					Out: cmd.OutOrStderr(),
					Err: cmd.OutOrStderr(),
				},
				Follow: true,
			}, builds.WithTaskServiceAccount(ctx, opts.ServiceAccount, opts.references...))
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
	processors = detector(taskCmd, task.Spec.Params, results)

	return taskCmd, nil
}
