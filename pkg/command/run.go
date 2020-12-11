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
	"net/url"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	minkcli "github.com/mattmoor/mink/pkg/cli"
	"github.com/mattmoor/mink/pkg/constants"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"k8s.io/apimachinery/pkg/util/sets"
)

var (
	specialParams = sets.NewString(constants.ImageTargetParam, constants.SourceBundleParam)

	// TODO(mattmoor): Uncomment this if/when we use it.
	// specialResults = sets.NewString(constants.ImageDigestResult)
)

// NewRunCommand implements 'kn-im run' command
func NewRunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "<todo>",
	}

	cmd.AddCommand(NewRunTaskCommand())
	cmd.AddCommand(NewRunPipelineCommand())

	return cmd
}

// RunOptions is a base for the RunFooOptions commands.
type RunOptions struct {
	// Inherit all of the base build options.
	BaseBuildOptions

	resource string

	references []name.Reference
}

// Processor is an interface for augmenting a vanilla run execution with
// additional pre/post functionality "mixed in".
type Processor interface {
	PreRun(params []v1beta1.ParamSpec) ([]v1beta1.Param, error)
	PostRun(results []v1beta1.TaskRunResult) error
}

func p2tResults(prr []v1beta1.PipelineRunResult) []v1beta1.TaskRunResult {
	trr := make([]v1beta1.TaskRunResult, 0, len(prr))
	for _, res := range prr {
		trr = append(trr, v1beta1.TaskRunResult(res))
	}
	return trr
}

// ProcessorFuncs is a helper for implementing Processor
type ProcessorFuncs struct {
	PreRunFunc  func([]v1beta1.ParamSpec) ([]v1beta1.Param, error)
	PostRunFunc func([]v1beta1.TaskRunResult) error
}

var _ Processor = (*ProcessorFuncs)(nil)

// PreRun implements Processor
func (pf *ProcessorFuncs) PreRun(params []v1beta1.ParamSpec) ([]v1beta1.Param, error) {
	if pf.PreRunFunc != nil {
		return pf.PreRunFunc(params)
	}
	return nil, nil
}

// PostRun implements Processor
func (pf *ProcessorFuncs) PostRun(results []v1beta1.TaskRunResult) error {
	if pf.PostRunFunc != nil {
		return pf.PostRunFunc(results)
	}
	return nil
}

func processParams(cmd *cobra.Command, params []v1beta1.ParamSpec) Processor {
	for _, param := range params {
		// Elide turning "special" parameters into arguments.
		if specialParams.Has(param.Name) {
			continue
		}

		switch param.Type {
		case v1beta1.ParamTypeArray:
			if param.Default != nil {
				cmd.Flags().StringSlice(param.Name, param.Default.ArrayVal, param.Description)
			} else {
				cmd.Flags().StringSlice(param.Name, nil, param.Description)
			}
		default:
			if param.Default != nil {
				cmd.Flags().String(param.Name, param.Default.StringVal, param.Description)
			} else {
				cmd.Flags().String(param.Name, "", param.Description)
			}
		}
	}

	return &ProcessorFuncs{
		PreRunFunc: func(params []v1beta1.ParamSpec) ([]v1beta1.Param, error) {
			ps := make([]v1beta1.Param, 0, len(params))
			for _, param := range params {
				// Elide turning "special" parameters into arguments.
				if specialParams.Has(param.Name) {
					continue
				}

				f := cmd.Flags().Lookup(param.Name)

				switch param.Type {
				case v1beta1.ParamTypeArray:
					v := f.Value.(pflag.SliceValue).GetSlice()
					if param.Default == nil && len(v) == 0 {
						return nil, minkcli.ErrMissingFlag(param.Name)
					}
					ps = append(ps, v1beta1.Param{
						Name: param.Name,
						// NewArrayOrString doesn't pick the correct type when there is a single argument.
						Value: v1beta1.ArrayOrString{
							Type:     v1beta1.ParamTypeArray,
							ArrayVal: v,
						}})
				default:
					if param.Default == nil && f.Value.String() == "" {
						return nil, minkcli.ErrMissingFlag(param.Name)
					}
					ps = append(ps, v1beta1.Param{
						Name:  param.Name,
						Value: *v1beta1.NewArrayOrString(f.Value.String()),
					})
				}
			}
			return ps, nil
		},
	}
}

func newResultProcessor(cmd *cobra.Command, results sets.String) Processor {
	// TODO(mattmoor): Incorporate the output descriptions.
	cmd.Flags().StringP("output", "o", "", "options: "+strings.Join(results.List(), ", "))
	return &ProcessorFuncs{
		PostRunFunc: func(results []v1beta1.TaskRunResult) error {
			result := cmd.Flags().Lookup("output").Value.String()
			if result == "" {
				return nil
			}
			for _, r := range results {
				if r.Name != result {
					continue
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s\n", strings.TrimSpace(r.Value))
				return nil
			}
			return fmt.Errorf("unable to find result %q", result)
		},
	}
}

type signatureDetector func(cmd *cobra.Command, params []v1beta1.ParamSpec, results sets.String) []Processor

func (opts *RunOptions) detectProcessors(cmd *cobra.Command, params []v1beta1.ParamSpec, results sets.String) (processors []Processor) {
	if len(params) > 0 {
		processors = append(processors, processParams(cmd, params))
	}
	if len(results) > 0 {
		processors = append(processors, newResultProcessor(cmd, results))
	}

	paramNames := make(sets.String, len(params))
	for _, param := range params {
		paramNames.Insert(param.Name)
	}

	if paramNames.Has(constants.SourceBundleParam) {
		processors = append(processors, &ProcessorFuncs{
			PreRunFunc: func(params []v1beta1.ParamSpec) ([]v1beta1.Param, error) {
				// Bundle up the source context in an image.
				sourceDigest, err := opts.bundle(context.Background())
				if err != nil {
					return nil, err
				}
				opts.references = append(opts.references, sourceDigest)

				return []v1beta1.Param{{
					Name:  constants.SourceBundleParam,
					Value: *v1beta1.NewArrayOrString(sourceDigest.String()),
				}}, nil
			},
		})
	}

	if paramNames.Has(constants.ImageTargetParam) {
		var tag name.Tag
		processors = append(processors, &ProcessorFuncs{
			PreRunFunc: func(params []v1beta1.ParamSpec) ([]v1beta1.Param, error) {
				var err error
				tag, err = opts.tag(imageNameContext{
					URL: url.URL{
						Scheme: opts.resource,
					},
				})
				if err != nil {
					return nil, err
				}
				opts.references = append(opts.references, tag)

				return []v1beta1.Param{{
					Name:  constants.ImageTargetParam,
					Value: *v1beta1.NewArrayOrString(tag.String()),
				}}, nil
			},
		})

		if results.Has(constants.ImageDigestResult) {
			processors = append(processors, &ProcessorFuncs{
				PostRunFunc: func(results []v1beta1.TaskRunResult) error {
					if result := cmd.Flags().Lookup("output").Value.String(); result != "" {
						return nil
					}
					for _, r := range results {
						if r.Name != constants.ImageDigestResult {
							continue
						}
						fmt.Fprintf(cmd.OutOrStdout(), "%s@%s\n", tag.String(), strings.TrimSpace(r.Value))
						return nil
					}
					return fmt.Errorf("unable to find result %q", constants.ImageDigestResult)
				},
			})
		}
	}

	return processors
}

// ValidationErrorProcessor constructs a Processor that surfaces a validation error.
func ValidationErrorProcessor(f string, args ...interface{}) Processor {
	return &ProcessorFuncs{
		PostRunFunc: func([]v1beta1.TaskRunResult) error {
			return fmt.Errorf(f, args...)
		},
	}
}
