// Copyright 2018 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package commands

import (
	"log"
	"os"

	"github.com/google/ko/pkg/commands/options"
	"github.com/spf13/cobra"
)

// addResolve augments our CLI surface with resolve.
func addResolve(topLevel *cobra.Command) {
	po := &options.PublishOptions{}
	fo := &options.FilenameOptions{}
	so := &options.SelectorOptions{}
	sto := &options.StrictOptions{}
	bo := &options.BuildOptions{}

	resolve := &cobra.Command{
		Use:   "resolve -f FILENAME",
		Short: "Print the input files with image references resolved to built/pushed image digests.",
		Long:  `This sub-command finds import path references within the provided files, builds them into Go binaries, containerizes them, publishes them, and prints the resulting yaml.`,
		Example: `
  # Build and publish import path references to a Docker
  # Registry as:
  #   ${KO_DOCKER_REPO}/<package name>-<hash of import path>
  # When KO_DOCKER_REPO is ko.local, it is the same as if
  # --local and --preserve-import-paths were passed.
  ko resolve -f config/

  # Build and publish import path references to a Docker
  # Registry preserving import path names as:
  #   ${KO_DOCKER_REPO}/<import path>
  # When KO_DOCKER_REPO is ko.local, it is the same as if
  # --local was passed.
  ko resolve --preserve-import-paths -f config/

  # Build and publish import path references to a Docker
  # daemon as:
  #   ko.local/<import path>
  # This always preserves import paths.
  ko resolve --local -f config/`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := createCancellableContext()
			builder, err := makeBuilder(ctx, bo)
			if err != nil {
				log.Fatalf("error creating builder: %v", err)
			}
			publisher, err := makePublisher(po)
			if err != nil {
				log.Fatalf("error creating publisher: %v", err)
			}
			defer publisher.Close()
			if err := resolveFilesToWriter(ctx, builder, publisher, fo, so, sto, os.Stdout); err != nil {
				log.Fatal(err)
			}
		},
	}
	options.AddPublishArg(resolve, po)
	options.AddFileArg(resolve, fo)
	options.AddSelectorArg(resolve, so)
	options.AddBuildOptions(resolve, bo)
	topLevel.AddCommand(resolve)
}
