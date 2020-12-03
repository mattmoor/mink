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

package builds

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/tektoncd/cli/pkg/options"
	tknv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	tektonclientset "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
)

// CancelablePipelineOption is a function option that can be used to customize a pipelinerun in
// certain ways prior to execution.  Each option may return a cancel function, which
// can be used to clean up any temporary artifacts created in support of this pipeline run.
type CancelablePipelineOption func(context.Context, *tknv1beta1.PipelineRun) (context.CancelFunc, error)

// RunPipeline executes the provided PipelineRun with the provided options applied, and returns
// the final PipelineRun state (or error) upon completion.
func RunPipeline(ctx context.Context, tr *tknv1beta1.PipelineRun, opt *options.LogOptions, opts ...CancelablePipelineOption) (*tknv1beta1.PipelineRun, error) {
	// TODO(mattmoor): expose masterURL and kubeconfig flags.
	cfg, err := GetConfig("", "")
	if err != nil {
		return nil, err
	}
	client, err := tektonclientset.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	for _, o := range opts {
		cancel, err := o(ctx, tr)
		if err != nil {
			return nil, err
		}
		defer cancel()
	}

	tr, err = client.TektonV1beta1().PipelineRuns(tr.Namespace).Create(ctx, tr, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	// TODO(mattmoor): From here down assumes opt.Follow, but if we want to have
	// a --no-wait or something then we should have an early-out here.
	defer client.TektonV1beta1().PipelineRuns(tr.Namespace).Delete(context.Background(), tr.Name, metav1.DeleteOptions{})

	opt.PipelineRunName = tr.Name
	if err := streamLogs(ctx, opt); err != nil {
		return nil, err
	}

	// Spin waiting for the final status.
	for {
		// See if our context has been cancelled
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Fetch the final state of the build.
		tr, err = client.TektonV1beta1().PipelineRuns(tr.Namespace).Get(ctx, tr.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		// Return an error if the build failed.
		cond := tr.Status.GetCondition(apis.ConditionSucceeded)
		if cond.IsFalse() {
			return nil, fmt.Errorf("%s: %s", cond.Reason, cond.Message)
		} else if !cond.IsTrue() {
			continue
		}

		return tr, nil
	}
}

// WithPipelineServiceAccount is used to adjust the PipelineRun to execute as a particular
// service account, as specified by the user.  It supports a special "me" sentinel
// which configures a temporary ServiceAccount infused with the local credentials
// for the container registry hosting the image we will publish to (and to which
// the source is published).
func WithPipelineServiceAccount(sa string, refs ...name.Reference) CancelablePipelineOption {
	return func(ctx context.Context, tr *tknv1beta1.PipelineRun) (context.CancelFunc, error) {
		if sa != "me" {
			tr.Spec.ServiceAccountName = sa
			return func() {}, nil
		}
		return nil, errors.New("--as=me is not yet supported for pipelines")
	}
}
