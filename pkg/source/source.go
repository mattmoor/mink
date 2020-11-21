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

package source

import (
	"context"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/mattmoor/mink/pkg/kontext"
	tknv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

// GitLocation allows cloning from git
type GitLocation struct {
	// URL the git repo URL
	URL string

	// Revision is the git revision to clone
	Revision string

	// Verbose enable verbose logging for git
	Verbose bool
}

var (
	// based on a simplification of the git clone task from the tekton catalog:
	// https://github.com/tektoncd/catalog/tree/master/task/git-clone
	gitCloneImage = "gcr.io/tekton-releases/github.com/tektoncd/pipeline/cmd/git-init:v0.18.0"

	gitCloneScript = `#!/bin/sh
set -eu -o pipefail
if [[ "$VERBOSE" == "true" ]] ; then
  set -x
fi
/ko-app/git-init \
  -url $GIT_URL \
  -revision $GIT_REVISION \
  -path /workspace \
  -sslVerify=$GIT_SSL_VERIFY \
  -submodules=$GIT_SUBMODULES \
  -depth $GIT_DEPTH
`
)

// CreateSourceSteps creates the source step(s) to get the source code from an image or git
func CreateSourceSteps(ctx context.Context, directory string, tag name.Tag, location *GitLocation) ([]tknv1beta1.Step, []name.Reference, error) {
	if location == nil {
		// lets bundle the source into a container image
		kontext, err := kontext.Bundle(ctx, directory, tag)
		if err != nil {
			return nil, nil, err
		}
		return []tknv1beta1.Step{{
			Container: corev1.Container{
				Name:       "extract-bundle",
				Image:      kontext.String(),
				WorkingDir: "/workspace",
			},
		}}, []name.Reference{tag, kontext}, nil
	}

	verbose := ""
	if location.Verbose {
		verbose = "true"
	}
	// lets git clone instead
	revision := location.Revision
	if revision == "" {
		revision = "HEAD"
	}
	return []tknv1beta1.Step{{
		Container: corev1.Container{
			Name:       "extract-bundle",
			Image:      gitCloneImage,
			WorkingDir: "/",
			Env: []corev1.EnvVar{
				{
					Name:  "GIT_URL",
					Value: location.URL,
				},
				{
					Name:  "GIT_REVISION",
					Value: revision,
				},
				{
					Name:  "VERBOSE",
					Value: verbose,
				},
				{
					Name:  "GIT_SSL_VERIFY",
					Value: "true",
				},
				{
					Name:  "GIT_SUBMODULES",
					Value: "false",
				},
				{
					Name:  "GIT_DEPTH",
					Value: "1",
				},
			},
		},
		Script: gitCloneScript,
	}}, []name.Reference{tag}, nil
}
