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
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/tektoncd/cli/pkg/cmd/pipelinerun"
	"github.com/tektoncd/cli/pkg/cmd/taskrun"
	"github.com/tektoncd/cli/pkg/options"
	tknv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Run executes the provided TaskRun with the provided options applied, and returns
// the fully-qualified image digest (or error) upon completion.
func Run(ctx context.Context, image string, tr *tknv1beta1.TaskRun, opt *options.LogOptions, opts ...CancelableTaskOption) (name.Digest, error) {
	tr, err := RunTask(ctx, tr, opt, opts...)
	if err != nil {
		return name.Digest{}, err
	}

	for _, result := range tr.Status.TaskRunResults {
		if result.Name != "IMAGE-DIGEST" {
			continue
		}
		value := strings.TrimSpace(result.Value)

		// Extract the IMAGE-DIGEST result.
		return name.NewDigest(image + "@" + value)
	}
	return name.Digest{}, errors.New("taskrun did not produce an IMAGE-DIGEST result")
}

func streamLogs(ctx context.Context, opt *options.LogOptions) error {
	// TODO(mattmoor): This should take a context so that it can be cancelled.
	errCh := make(chan error)
	go func() {
		defer close(errCh)
		switch {
		case opt.TaskrunName != "":
			errCh <- taskrun.Run(opt)
		case opt.PipelineRunName != "":
			errCh <- pipelinerun.Run(opt)
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// GetConfig is forked out of sharedmain because linking knative.dev/pkg/metrics spews logs.
func GetConfig(masterURL, kubeconfig string) (*rest.Config, error) {
	if kubeconfig == "" {
		kubeconfig = os.Getenv("KUBECONFIG")
	}
	// If we have an explicit indication of where the kubernetes config lives, read that.
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	}
	// If not, try the in-cluster config.
	if c, err := rest.InClusterConfig(); err == nil {
		return c, nil
	}
	// If no in-cluster config, try the default location in the user's home directory.
	if usr, err := user.Current(); err == nil {
		if c, err := clientcmd.BuildConfigFromFlags("", filepath.Join(usr.HomeDir, ".kube", "config")); err == nil {
			return c, nil
		}
	}

	return nil, fmt.Errorf("could not create a valid kubeconfig")
}
