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

var runPipelineExample = fmt.Sprintf(`
  # Create a PipelineRun instantiating the pipeline "build-stuff" and passing
  # values to named arguments.
  %[1]s run pipeline build-stuff -- --arg1=val1 --arg2=val2
`, ExamplePrefix())

// NewRunPipelineCommand implements 'kn-im run pipeline' command
func NewRunPipelineCommand(ctx context.Context) *cobra.Command {
	opts := &RunPipelineOptions{
		RunOptions: RunOptions{
			BaseBuildOptions: BaseBuildOptions{BundleOptions: BundleOptions{ctx: ctx}},
			resource:         "pipeline",
		},
	}

	cmd := &cobra.Command{
		Use:          "pipeline NAME",
		Short:        "Create a PipelineRun to execute a pipeline.",
		Example:      runPipelineExample,
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

// RunPipelineOptions implements Interface for the `kn im run pipeline` command.
type RunPipelineOptions struct {
	// Inherit all of the base run options.
	RunOptions
}

// RunPipelineOptions implements Interface
var _ Interface = (*RunPipelineOptions)(nil)

// AddFlags implements Interface
func (opts *RunPipelineOptions) AddFlags(cmd *cobra.Command) {
	// Add the bundle flags to our surface.
	opts.BaseBuildOptions.AddFlags(cmd)

}

// Validate implements Interface
func (opts *RunPipelineOptions) Validate(cmd *cobra.Command, args []string) error {
	// Validate the bundle arguments.
	if err := opts.BaseBuildOptions.Validate(cmd, args); err != nil {
		return err
	}
	return nil
}

// Execute implements Interface
func (opts *RunPipelineOptions) Execute(cmd *cobra.Command, args []string) error {
	ctx := opts.GetContext(cmd)

	// We take one positional argument, pass that as the pipeline name.
	taskCmd, err := opts.buildCmd(ctx, args[0], opts.detectProcessors)
	if err != nil {
		return err
	}

	// Pass the remaining arguments to the sub-command.
	// These are all after the --
	taskCmd.SetArgs(args[1:])

	return taskCmd.Execute()
}

// buildCmd constructs a cobra.Command for the named pipeline.
func (opts *RunPipelineOptions) buildCmd(ctx context.Context, pipelineName string, detector signatureDetector) (*cobra.Command, error) {
	client := pipelineclient.Get(ctx)

	// Load the pipeline definition.
	ns := opts.Namespace
	pipeline, err := client.TektonV1beta1().Pipelines(ns).Get(ctx, pipelineName, metav1.GetOptions{})
	if apierrs.IsNotFound(err) {
		return nil, fmt.Errorf("pipeline %q not found: %w", fmt.Sprintf("%s/%s", ns, pipelineName), err)
	} else if err != nil {
		return nil, err
	}

	var processors []Processor
	pipelineCmd := &cobra.Command{
		Use:   "mink run pipeline " + pipeline.Name,
		Short: pipeline.Spec.Description,
		RunE: func(cmd *cobra.Command, args []string) error {
			pr := &v1beta1.PipelineRun{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ns,
					GenerateName: "mink-" + pipeline.Name + "-",
				},
				Spec: v1beta1.PipelineRunSpec{
					PipelineRef: &v1beta1.PipelineRef{
						Name: pipeline.Name,
					},
				},
			}

			for _, processor := range processors {
				ps, err := processor.PreRun(pipeline.Spec.Params)
				if err != nil {
					return err
				}
				pr.Spec.Params = append(pr.Spec.Params, ps...)
			}

			pr, err := builds.RunPipeline(ctx, pr, &options.LogOptions{
				ActivityTimeout: activityTimeout,
				Params:          &cli.TektonParams{},
				Stream: &cli.Stream{
					// Send Out to stderr so we can capture the digest for composition.
					Out: cmd.OutOrStderr(),
					Err: cmd.OutOrStderr(),
				},
				Follow: true,
			}, builds.WithPipelineServiceAccount(ctx, opts.ServiceAccount, opts.references...))
			if err != nil {
				return err
			}

			trr := p2tResults(pr.Status.PipelineResults)
			for _, processor := range processors {
				if err := processor.PostRun(trr); err != nil {
					return err
				}
			}
			return nil
		},
	}

	// Process the results of the pipeline.
	results := make(sets.String, len(pipeline.Spec.Results))
	for _, result := range pipeline.Spec.Results {
		results.Insert(result.Name)
	}

	// Based on the signature determine which processors to wire in.
	processors = detector(pipelineCmd, pipeline.Spec.Params, results)

	return pipelineCmd, nil
}
